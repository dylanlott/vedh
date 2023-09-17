//go:generate go run github.com/99designs/gqlgen generate

package server

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/99designs/gqlgen/handler"
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
) (*graphQLServer, error) {

	return &graphQLServer{
		mutex:  sync.RWMutex{},
		db:     db,
		games:  map[string]*FullGame{},
		boards: map[string]*FullBoardstate{},
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
	mux.Handle("/playground", playground.Handler("GraphQL", route))
	mux.Handle("/prometheus", promhttp.Handler())
	log.Printf("serving graphiql at localhost:%d/playground", port)
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
