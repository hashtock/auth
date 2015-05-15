package main

import (
	"log"
	"net/http"

	"github.com/hashtock/service-tools/serialize"

	"github.com/hashtock/auth/conf"
	"github.com/hashtock/auth/storage"
	"github.com/hashtock/auth/webapp"
)

func main() {
	cfg := conf.GetConfig()

	mongoStorage, err := storage.NewMongoStorage(cfg.DB, cfg.DBName)
	if err != nil {
		log.Fatalln("Could not configure storage. ", err)
	}

	handlerOptions := webapp.Options{
		Serializer:         &serialize.WebAPISerializer{},
		Storage:            mongoStorage,
		AppAddress:         cfg.AppAddress,
		GoogleClientID:     cfg.GoogleClientID,
		GoogleClientSecret: cfg.GoogleClientSecret,
		SessionSecret:      cfg.SessionSecret,
	}

	handler := webapp.Handlers(handlerOptions)

	err = http.ListenAndServe(cfg.ServeAddress, handler)
	if err != nil {
		log.Fatalln(err)
	}
}
