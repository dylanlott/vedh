//go:generate go run github.com/99designs/gqlgen

package server

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/dylanlott/edh-go/persistence"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/99designs/gqlgen/handler"
	"github.com/go-redis/redis"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"github.com/tinrab/retry"
	"github.com/zeebo/errs"
)

// Conf takes configuration values and loads them from the environment into our struct.
type Conf struct {
	RedisURL    string `envconfig:"REDIS_URL" default:"redis://localhost:6379"`
	PostgresURL string `envconfig:"DATABASE_URL" default:"postgres://edhgo:edhgodev@localhost:5432/edhgo?sslmode=disable"`
	DefaultPort int    `envconfig:"PORT" default:"8080"`
}

// var userCtxKey = &contextKey{"user"}
// type contextKey struct {
// 	name string
// }

// graphQLServer binds the whole app together and implements the GraphQL interfac
type graphQLServer struct {
	mutex sync.RWMutex

	// Persistence layers
	redisClient *redis.Client
	db          *sql.DB

	// TODO: make sure the games and boards maps are moved to redis so we can
	// survive a restart without losing game state.

	// TODO: Write a function for recovering boardstates and game channels
	// if we consider them currently active.

	// games holds a reference to *FullGames in the server.
	// * games rely on both Game and Board channels.
	games map[string]*FullGame

	// boards holds references to *FullBoards which track BoardObservers
	boards map[string]*FullBoardstate

	// Update channels are only stored in memory.
	// If the server crashes, we'll need to setup all the channels for
	// existing games again.
	messageChannels map[string]chan *Message
	userChannels    map[string]chan string
}

// NewGraphQLServer creates a new server to attach the database, game engine,
// and graphql connections together
func NewGraphQLServer(
	kv persistence.KV,
	db *sql.DB,
	cfg Conf,
) (*graphQLServer, error) {
	// TODO: Remove this redis client and wire chat up to KV interface instead
	log.Printf("attempting to connecto to redis at %s", cfg.RedisURL)
	opts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Printf("failed to get redis options: %s", err)
		return nil, errs.Wrap(err)
	}
	client := redis.NewClient(opts)

	retry.ForeverSleep(2*time.Second, func(_ int) error {
		_, err := client.Ping().Result()
		if err != nil {
			log.Printf("error connecting to redis: %+v\n", err)
		}
		return err
	})

	return &graphQLServer{
		mutex:           sync.RWMutex{},
		db:              db,
		redisClient:     client,
		games:           map[string]*FullGame{},
		boards:          map[string]*FullBoardstate{},
		messageChannels: map[string]chan *Message{},
		userChannels:    map[string]chan string{},
	}, nil
}

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
	// serve the graphql playground
	mux.Handle("/playground", playground.Handler("GraphQL", route))
	// serve prometheus metrics
	mux.Handle("/prometheus", promhttp.Handler())
	log.Printf("serving graphiql at localhost:%d/playground", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), h)
}

// auth is a middleware responsible for passing header and cookie info to context
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
