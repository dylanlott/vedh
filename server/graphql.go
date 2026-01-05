//go:generate go run github.com/99designs/gqlgen generate

package server

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/99designs/gqlgen/handler"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
)

// Conf takes configuration values and loads them from the environment into our struct.
type Conf struct {
	RedisURL    string `envconfig:"REDIS_URL" default:"redis://localhost:6379"`
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
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
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
	mux.Handle(
		route,
		handler.GraphQL(NewExecutableSchema(Config{Resolvers: s}),
			handler.WebsocketUpgrader(websocket.Upgrader{
				CheckOrigin: func(r *http.Request) bool {
					return true
				},
			}),
			handler.WebsocketKeepAliveDuration(time.Second*10),
		),
	)
	h := cors.AllowAll().Handler(s.auth(mux))
	h = s.withRequestLogging(h)
	h = s.withRequestID(h)
	mux.Handle("/playground", playground.Handler("GraphQL", route))
	mux.Handle("/prometheus", promhttp.Handler())
	s.logger.Info("serving graphiql", "url", fmt.Sprintf("http://localhost:%d/playground", port))
	return http.ListenAndServe(fmt.Sprintf(":%d", port), h)
}

// auth is a middleware responsible for passing header and cookie info to the
// context
func (s *graphQLServer) auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)

		// TODO: all of this code below needs to be vetted to find where
		// Gorilla's websocket is getting clobbered and fix that

		// username, err := r.Cookie("username")
		// if err != nil {
		// 	// log.Printf("username not present: %s", err)
		// 	// This means they're not attempting to auth, so we need to let them through
		// 	// to either sign up or login.
		// 	next.ServeHTTP(w, r)
		// 	return
		// }
		// token, err := r.Cookie("token")
		// if err != nil {
		// 	log.Printf("error parsing token: %s", err)
		// 	next.ServeHTTP(w, r)
		// }
		// if token == nil {
		// 	// allow unauthed users in for login/signup
		// 	next.ServeHTTP(w, r)
		// 	return
		// }

		// // If they have a token, we have to compare it.
		// log.Printf("token: %s", token)
		// // find and authenticate
		// rows, err := s.db.Query("SELECT * FROM users WHERE username = ?", username)
		// if err != nil {
		// 	log.Printf("error querying users: %s", err)
		// 	next.ServeHTTP(w, r)
		// 	return
		// }

		// for rows.Next() {
		// 	cols, err := rows.Columns()
		// 	if err != nil {
		// 		log.Printf("error getting rows: %s", err)
		// 	}
		// 	log.Printf("columns: %s", cols)
		// 	continue
		// }

		// // TODO: Assign the found user to the request
		// // ctx := context.WithValue(r.Context(), userCtxKey, user)
		// // r = r.WithContext(ctx)
		// next.ServeHTTP(w, r)
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

// func use(h http.HandlerFunc, middleware ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
// 	for _, m := range middleware {
// 		h = m(h)
// 	}

// 	return h
// }
