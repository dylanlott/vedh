package main

import (
	"log"

	"github.com/zeebo/errs"

	"github.com/dylanlott/edh-go/persistence"
	"github.com/dylanlott/edh-go/server"
	"github.com/kelseyhightower/envconfig"
)

func main() {
	var cfg server.Conf
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err)
	}
	db, err := persistence.NewAppDatabase("./persistence/migrations/", cfg.PostgresURL)
	if err != nil {
		log.Fatal(errs.Wrap(err))
	}
	log.Println("opened app database")
	cardDB, err := persistence.NewSQLite("./persistence/AllPrintings.sqlite")
	if err != nil {
		log.Fatalf(errs.Wrap(err).Error())
	}
	log.Println("opened card database")
	kv, err := persistence.NewRedis(cfg.RedisURL, "", nil)
	if err != nil {
		log.Fatalf("failed to start redis: %s", errs.Wrap(err))
	}
	log.Println("created new redis store")
	s, err := server.NewGraphQLServer(kv, db, cardDB, cfg)
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
