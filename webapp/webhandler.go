package webapp

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/hashtock/service-tools/serialize"
	"github.com/markbates/goth/gothic"

	"github.com/hashtock/auth/conf"
)

// Handlers build http handler for service API
func Handlers(cfg *conf.Config) http.Handler {
	auth := authService{
		SessionName:  cfg.SessionName,
		SessionStore: sessions.NewCookieStore([]byte(cfg.SessionSecret)),
		Serializer:   serialize.WebAPISerializer{},
	}

	m := mux.NewRouter()

	// Externall endpoints
	m.HandleFunc("/who/", auth.who).Methods("GET")
	m.HandleFunc("/providers/", auth.providers).Methods("GET")
	m.HandleFunc("/logout/", auth.logout).Methods("GET")
	loginRoute := m.HandleFunc("/login/{provider}/", gothic.BeginAuthHandler).Methods("GET")

	// Internal endpoint
	callbackRoute := m.HandleFunc("/login/{provider}/callback", auth.authCallback).Methods("GET")

	providers, err := setupGoth(cfg, loginRoute, callbackRoute)
	if err != nil {
		fmt.Println("Could not configure Auth providers.", err)
		os.Exit(1)
	}
	auth.Providers = providers

	return m
}
