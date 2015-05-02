package webapp

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth/gothic"

	"github.com/hashtock/auth/conf"
)

// Handlers build http handler for service API
func Handlers(cfg *conf.Config) http.Handler {
	auth := authService{
		// SessionName: cfg.SessionName,
		SessionName:  "test",
		SessionStore: sessions.NewCookieStore([]byte("something-very-secret")),
	}

	m := mux.NewRouter()

	m.HandleFunc("/", auth.index).Methods("GET")
	m.HandleFunc("/auth/{provider}/", gothic.BeginAuthHandler).Methods("GET")
	callbackRoute := m.HandleFunc("/auth/{provider}/callback", auth.authCallback).Methods("GET")

	if err := setupGoth(cfg, callbackRoute); err != nil {
		fmt.Println("Could not configure Auth providers.", err)
		os.Exit(1)
	}

	return m
}
