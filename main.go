package main

import (
	"log"

	"github.com/zeebo/errs"

	"github.com/openmtg/edh-go/persistence"
	"github.com/openmtg/edh-go/server"

	"github.com/kelseyhightower/envconfig"
)

func main() {
	var cfg server.Conf
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err)
	}
	db, err := persistence.NewDB(cfg.PostgresURL)
	if err != nil {
		log.Fatal(errs.Wrap(err))
	}
	log.Println("successfully opened database connection")
	kv, err := persistence.NewRedis(cfg.RedisURL, "")
	if err != nil {
		log.Fatalf("failed to start redis: %s", errs.Wrap(err))
	}
	log.Println("created new redis store")
	s, err := server.NewGraphQLServer(kv, db, cfg)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("graphQL server attempting to start")
	err = s.Serve("/graphql", cfg.DefaultPort)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("serving /graphql at :%d", cfg.DefaultPort)
	log.Printf("serving graphiql playground at :%d/playground", cfg.DefaultPort)
}
