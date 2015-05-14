package main

import (
	"log"
	"net/http"

	"github.com/hashtock/auth/conf"
	"github.com/hashtock/auth/storage"
	"github.com/hashtock/auth/webapp"
)

func main() {
	cfg := conf.GetConfig()

	storage, err := storage.NewMongoStorage(cfg.DB, cfg.DBName)
	if err != nil {
		log.Fatalln("Could not configure storage. ", err)
	}

	handler := webapp.Handlers(cfg, storage)

	err = http.ListenAndServe(cfg.ServeAddress, handler)
	if err != nil {
		log.Fatalln(err)
	}
}
