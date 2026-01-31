//go:generate go run github.com/99designs/gqlgen generate

package server

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/99designs/gqlgen/handler"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
)

// Conf takes configuration values and loads them from the environment into our struct.
type Conf struct {
	PostgresURL string `envconfig:"DATABASE_URL" default:"postgres://edhgo:edhgo@localhost:5432/edhgo?sslmode=disable"`
	DefaultPort int    `envconfig:"PORT" default:"8080"`
}

// var userCtxKey = &contextKey{"user"}
// type contextKey struct {
// 	name string
// }

// graphQLServer binds the whole app together and implements the GraphQL interfac
type graphQLServer struct {
	mutex sync.RWMutex

	logger *slog.Logger

	// Persistence layers
	db *sql.DB

	// games holds a reference to *FullGames in the server.
	// * games rely on both Game and Board channels.
	games map[string]*FullGame

	// boards holds references to *FullBoards which track BoardObservers
	boards map[string]*FullBoardstate
}

// NewGraphQLServer creates a new server to attach the database, game engine,
// and graphql connections together
func NewGraphQLServer(
	db *sql.DB,
	cfg Conf,
	logger *slog.Logger,
) (*graphQLServer, error) {
	if logger == nil {
		logger = slog.Default()
	}
	return &graphQLServer{
		mutex:  sync.RWMutex{},
		logger: logger,
		db:     db,
		games:  map[string]*FullGame{},
		boards: map[string]*FullBoardstate{},
	}, nil
}

type requestIDContextKey struct{}

func (s *graphQLServer) loggerFor(ctx context.Context) *slog.Logger {
	if s.logger == nil {
		return slog.Default()
	}
	if ctx == nil {
		return s.logger
	}
	if rid, ok := ctx.Value(requestIDContextKey{}).(string); ok && rid != "" {
		return s.logger.With("request_id", rid)
	}
	return s.logger
}

type statusRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.wroteHeader = true
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(p []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	return r.ResponseWriter.Write(p)
}

func (r *statusRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := r.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("response writer does not support hijacking")
	}
	return h.Hijack()
}

func (r *statusRecorder) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (r *statusRecorder) Push(target string, opts *http.PushOptions) error {
	if p, ok := r.ResponseWriter.(http.Pusher); ok {
		return p.Push(target, opts)
	}
	return http.ErrNotSupported
}

func (s *graphQLServer) withRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-Id")
		if requestID == "" {
			requestID = uuid.NewString()
		}
		w.Header().Set("X-Request-Id", requestID)
		ctx := context.WithValue(r.Context(), requestIDContextKey{}, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *graphQLServer) withRequestLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		dur := time.Since(start)
		logger := s.loggerFor(r.Context()).With(
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration_ms", dur.Milliseconds(),
		)
		if rec.status >= 500 {
			logger.Error("http request")
			return
		}
		logger.Info("http request")
	})
}

// Serve is a blocking function that runs the server until anything returns an error.
// It sets up the muxed routing, exposes the prometheus endpoint, and serves the
// GraphQL playground and API at the given route and port.
func (s *graphQLServer) Serve(route string, port int) error {
	mux := http.NewServeMux()
	gqlHandler := handler.GraphQL(NewExecutableSchema(Config{Resolvers: s}),
		handler.WebsocketUpgrader(websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}),
		handler.WebsocketInitFunc(func(ctx context.Context, initPayload transport.InitPayload) (context.Context, *transport.InitPayload, error) {
			user, err := parseAuthFromInitPayload(initPayload)
			if err != nil {
				return ctx, nil, err
			}
			return withAuth(ctx, user), nil, nil
		}),
		handler.WebsocketKeepAliveDuration(time.Second*10),
	)
	mux.Handle(
		route,
		gqlHandler,
	)
	h := cors.AllowAll().Handler(s.withAuthContext(mux))
	h = s.withRequestLogging(h)
	h = s.withRequestID(h)
	mux.Handle("/playground", playground.Handler("GraphQL", route))
	mux.Handle("/prometheus", promhttp.Handler())
	s.logger.Info("serving graphiql", "url", fmt.Sprintf("http://localhost:%d/playground", port))
	return http.ListenAndServe(fmt.Sprintf(":%d", port), h)
}

// withAuthContext validates bearer tokens (if present) and attaches the user to the request context.
func (s *graphQLServer) withAuthContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := parseAuthFromRequest(r)
		if err != nil {
			s.loggerFor(r.Context()).Warn("invalid auth token", "err", err)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		ctx := withAuth(r.Context(), user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *graphQLServer) Mutation() MutationResolver {
	return s
}

func (s *graphQLServer) Query() QueryResolver {
	return s
}

func (s *graphQLServer) Subscription() SubscriptionResolver {
	return s
}
