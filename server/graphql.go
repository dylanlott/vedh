//go:generate go run github.com/99designs/gqlgen generate

package server

import (
	"bufio"
	"context"
	"crypto/subtle"
	"database/sql"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/99designs/gqlgen/handler"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/openmtg/edh-go/pkg/ratelimit"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
)

const (
	readHeaderTimeout         = 5 * time.Second
	readTimeout               = 15 * time.Second
	writeTimeout              = 30 * time.Second
	idleTimeout               = 60 * time.Second
	maxHeaderBytes            = 1 << 20 // 1 MiB
	maxGraphQLBodyBytes int64 = 1 << 20 // 1 MiB
)

// Conf takes configuration values and loads them from the environment into our struct.
type Conf struct {
	PostgresURL    string `envconfig:"DATABASE_URL" default:"postgres://edhgo:edhgo@localhost:5432/edhgo?sslmode=disable"`
	DefaultPort    int    `envconfig:"PORT" default:"8080"`
	AllowedOrigins string `envconfig:"ALLOWED_ORIGINS" default:"http://localhost:5173,http://127.0.0.1:5173,http://localhost:8080,http://127.0.0.1:8080"`
	MetricsEnabled bool   `envconfig:"METRICS_ENABLED" default:"false"`
	MetricsToken   string `envconfig:"METRICS_TOKEN" default:""`

	// DeckImportRatePerMinute and DeckImportRateBurst bound the token
	// bucket server/ratelimit.go constructs for previewDeck and
	// trackProductEvent (pkg/ratelimit.Registry). Tuned generously per
	// 01-RESEARCH.md Open Question 2: previewDeck sits on the activation
	// critical path, and a limit set too tight would throttle the very
	// funnel this phase exists to measure. The idle-eviction window that
	// bounds the registry's own growth is deliberately not a third
	// variable here — see pkg/ratelimit's idleWindow constant.
	DeckImportRatePerMinute int `envconfig:"DECK_IMPORT_RATE_PER_MINUTE" default:"30"`
	DeckImportRateBurst     int `envconfig:"DECK_IMPORT_RATE_BURST" default:"10"`

	// DeckProviderAllowedHosts is the comma-separated exact-match host
	// allowlist server/deck_providers.go's redirect hook (and, once plan
	// 01-07 wires an adapter, the initial fetch) consult. Defaults to
	// empty, which is the correct default: with nothing allowlisted,
	// nothing is fetchable.
	DeckProviderAllowedHosts string `envconfig:"DECK_PROVIDER_ALLOWED_HOSTS" default:""`

	// DeckProviderEnabled is the provider kill switch (D-16/T-01-28).
	// Defaults to false: a deploy must be safe before this is
	// deliberately turned on, and the release must be able to drop to
	// paste-only activation independently of any other switch.
	// providerEnabled() (server/deck_providers.go) also requires a
	// non-empty DeckProviderAllowedHosts even when this is true.
	DeckProviderEnabled bool `envconfig:"DECK_PROVIDER_ENABLED" default:"false"`

	// GuestCreationEnabled is the guest-creation kill switch (DEC-A,
	// REQ-ACT-005/T-02-02). Unlike DeckProviderEnabled, this defaults to
	// TRUE: guest creation dials no third party -- it is an internal,
	// rate-limited, bcrypt-backed write on our own users table, not an
	// SSRF surface with unread terms of service. Shipping it false by
	// default (mirroring DeckProviderEnabled "for symmetry") would leave
	// this entire phase's activation funnel inert on merge, with no way
	// for Phase 5's release gate to exercise the guest journey without an
	// out-of-band flag flip. The switch exists so an operator can KILL
	// the funnel under abuse, not so the funnel starts dead. Do not "fix"
	// this default to false by analogy with DeckProviderEnabled -- the
	// two switches guard structurally different risks.
	GuestCreationEnabled bool `envconfig:"GUEST_CREATION_ENABLED" default:"true"`

	// GuestSessionRatePerMinute and GuestSessionRateBurst are declared per
	// REQ-ACT-005's rate-limiting requirement for the guestSession
	// mutation (T-02-01). DEC-C deliberately keeps pkg/ratelimit.Registry
	// to a single (perMinute, burst) budget shared by every surface on
	// s.limiter, rather than fragmenting idle eviction and the
	// vedh_rate_limit_total mental model with a second Registry or a
	// per-surface budget map. SurfaceGuestSession therefore draws on the
	// same shared budget as SurfaceDeckImport/SurfaceProductEvent
	// (s.limiter, constructed from DeckImportRatePerMinute/Burst below);
	// these two fields exist to make the intended guest-specific budget
	// operator-visible and are the values a future per-surface Registry
	// enhancement would consume without a Conf shape change.
	GuestSessionRatePerMinute int `envconfig:"GUEST_SESSION_RATE_PER_MINUTE" default:"20"`
	GuestSessionRateBurst     int `envconfig:"GUEST_SESSION_RATE_BURST" default:"5"`
}

// var userCtxKey = &contextKey{"user"}
// type contextKey struct {
// 	name string
// }

// graphQLServer binds the whole app together and implements the GraphQL interfac
type graphQLServer struct {
	mutex sync.RWMutex

	logger *slog.Logger
	cfg    Conf

	// Persistence layers
	db *sql.DB

	// games holds a reference to *FullGames in the server.
	// * games rely on both Game and Board channels.
	games map[string]*FullGame

	// boards holds references to *FullBoards which track BoardObservers
	boards map[string]*FullBoardstate

	// allowedOrigins holds normalized origins permitted for CORS and websocket checks.
	allowedOrigins map[string]struct{}

	// limiter is the per-surface, per-client token-bucket registry
	// server/ratelimit.go's allowRequest consults for previewDeck and
	// trackProductEvent. A nil limiter (e.g. a graphQLServer built as a
	// struct literal by a test rather than through NewGraphQLServer)
	// allows every request, so tests that do not care about rate
	// limiting are unaffected.
	limiter *ratelimit.Registry

	// deckProviderAllowedHosts holds the parsed, lower-cased exact-match
	// host allowlist for the outbound deck-provider fetch client
	// (server/deck_providers.go), mirroring allowedOrigins' shape for the
	// CORS origin allowlist above.
	deckProviderAllowedHosts map[string]struct{}
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
		mutex:                    sync.RWMutex{},
		logger:                   logger,
		cfg:                      cfg,
		db:                       db,
		games:                    map[string]*FullGame{},
		boards:                   map[string]*FullBoardstate{},
		allowedOrigins:           parseAllowedOrigins(cfg.AllowedOrigins),
		limiter:                  ratelimit.NewRegistry(cfg.DeckImportRatePerMinute, cfg.DeckImportRateBurst),
		deckProviderAllowedHosts: parseAllowedHosts(cfg.DeckProviderAllowedHosts),
	}, nil
}

// parseAllowedHosts parses a comma-separated host list into a lower-cased
// set, mirroring parseAllowedOrigins' shape immediately below for the CORS
// origin allowlist. An empty or all-blank input yields an empty set.
func parseAllowedHosts(raw string) map[string]struct{} {
	allowed := make(map[string]struct{})
	for _, part := range strings.Split(raw, ",") {
		host := strings.ToLower(strings.TrimSpace(part))
		if host == "" {
			continue
		}
		allowed[host] = struct{}{}
	}
	return allowed
}

func parseAllowedOrigins(raw string) map[string]struct{} {
	allowed := make(map[string]struct{})
	for _, part := range strings.Split(raw, ",") {
		origin := strings.TrimSpace(part)
		if origin == "" {
			continue
		}
		normalized, ok := normalizeOrigin(origin)
		if !ok {
			continue
		}
		allowed[normalized] = struct{}{}
	}
	return allowed
}

func normalizeOrigin(origin string) (string, bool) {
	u, err := url.Parse(strings.TrimSpace(origin))
	if err != nil || u == nil {
		return "", false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", false
	}
	if u.Host == "" {
		return "", false
	}
	return u.Scheme + "://" + u.Host, true
}

type requestIDContextKey struct{}

// remoteAddrContextKey holds the caller's IP address (port stripped) in the
// request context, for server/ratelimit.go's client-key derivation to fall
// back on when a caller supplies no session ID.
type remoteAddrContextKey struct{}

// withClientAddr attaches the caller's remote address to the request
// context under remoteAddrContextKey. It deliberately reads only
// r.RemoteAddr, never the X-Forwarded-For header: X-Forwarded-For is
// client-supplied, and trusting it by default would let an unauthenticated
// caller mint an unlimited number of distinct rate-limit keys simply by
// varying the header on each request. A deployment sitting behind a
// trusted reverse proxy that itself sets or strips this header would need
// an explicit opt-in to trust it — no such opt-in exists yet because no
// proxy of that kind fronts this deployment today.
func (s *graphQLServer) withClientAddr(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), remoteAddrContextKey{}, remoteHostFromRequest(r))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// remoteHostFromRequest returns r.RemoteAddr with any port stripped, or the
// raw value if it is not in host:port form.
func remoteHostFromRequest(r *http.Request) string {
	if r == nil {
		return ""
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return strings.TrimSpace(r.RemoteAddr)
	}
	return host
}

// remoteAddrFromContext reads the value withClientAddr attaches to the
// request context.
func remoteAddrFromContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	addr, ok := ctx.Value(remoteAddrContextKey{}).(string)
	return addr, ok
}

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

func (s *graphQLServer) isAllowedOrigin(origin string) bool {
	normalized, ok := normalizeOrigin(origin)
	if !ok {
		return false
	}
	_, found := s.allowedOrigins[normalized]
	return found
}

func (s *graphQLServer) isAllowedWebSocketOrigin(r *http.Request) bool {
	origin := ""
	if r != nil {
		origin = strings.TrimSpace(r.Header.Get("Origin"))
	}
	// Non-browser clients may not send Origin; keep these flows possible.
	if origin == "" {
		return true
	}
	if s.isAllowedOrigin(origin) {
		return true
	}
	s.loggerFor(context.Background()).Warn("websocket origin denied", "origin", origin)
	return false
}

func (s *graphQLServer) shouldExposeMetrics() bool {
	return s != nil && s.cfg.MetricsEnabled && strings.TrimSpace(s.cfg.MetricsToken) != ""
}

func (s *graphQLServer) withMetricsAuth(next http.Handler) http.Handler {
	token := strings.TrimSpace(s.cfg.MetricsToken)
	if token == "" {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const prefix = "Bearer "
		raw := strings.TrimSpace(r.Header.Get("Authorization"))
		if !strings.HasPrefix(raw, prefix) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		presented := strings.TrimSpace(strings.TrimPrefix(raw, prefix))
		if subtle.ConstantTimeCompare([]byte(presented), []byte(token)) != 1 {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Serve is a blocking function that runs the server until anything returns an error.
// It sets up the muxed routing, exposes the prometheus endpoint, and serves the
// GraphQL playground and API at the given route and port.
func (s *graphQLServer) Serve(route string, port int) error {
	mux := http.NewServeMux()
	gqlHandler := handler.GraphQL(NewExecutableSchema(Config{Resolvers: s}),
		handler.WebsocketUpgrader(websocket.Upgrader{
			CheckOrigin: s.isAllowedWebSocketOrigin,
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
	mux.Handle(route, s.withMaxGraphQLBody(gqlHandler))
	corsMiddleware := cors.New(cors.Options{
		AllowOriginFunc:  s.isAllowedOrigin,
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodOptions},
		AllowedHeaders:   []string{"Authorization", "Content-Type", "X-Request-Id"},
		ExposedHeaders:   []string{"X-Request-Id"},
		AllowCredentials: false,
		MaxAge:           600,
	})
	h := corsMiddleware.Handler(s.withAuthContext(mux))
	h = s.withClientAddr(h)
	h = s.withRequestLogging(h)
	h = s.withRequestID(h)
	mux.Handle("/playground", playground.Handler("GraphQL", route))
	if s.shouldExposeMetrics() {
		mux.Handle("/prometheus", s.withMetricsAuth(promhttp.Handler()))
	} else if s != nil && s.cfg.MetricsEnabled {
		s.logger.Warn("prometheus metrics disabled because METRICS_TOKEN is empty")
	} else {
		s.logger.Info("prometheus metrics disabled; set METRICS_ENABLED=true to expose /prometheus")
	}
	s.logger.Info("serving graphiql", "url", fmt.Sprintf("http://localhost:%d/playground", port))
	server := newHTTPServer(port, h)
	return server.ListenAndServe()
}

// withMaxGraphQLBody limits request allocation before gqlgen decodes a JSON
// operation. gqlgen's POST transport reads the full request body itself, so
// resolver-level rate limits are too late to protect the process from an
// oversized public request.
func (s *graphQLServer) withMaxGraphQLBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.Body != nil {
			r.Body = http.MaxBytesReader(w, r.Body, maxGraphQLBodyBytes)
		}
		next.ServeHTTP(w, r)
	})
}

func newHTTPServer(port int, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		MaxHeaderBytes:    maxHeaderBytes,
	}
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
