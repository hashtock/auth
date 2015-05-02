package main

import (
	"log"
	"net/http"

	"github.com/hashtock/auth/conf"
	"github.com/hashtock/auth/webapp"
)

func main() {
	cfg := conf.GetConfig()

	handler := webapp.Handlers(cfg)

	err := http.ListenAndServe(cfg.ServeAddress, handler)
	if err != nil {
		log.Fatalln(err)
	}
}
