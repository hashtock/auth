package webapp

import (
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/gorilla/pat"
	"github.com/gorilla/sessions"
	"github.com/hashtock/service-tools/serialize"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/gplus"

	"github.com/hashtock/auth/core"
)

const (
	loginRoute    = "login"
	callbackRoute = "callback"
)

type Options struct {
	Serializer         serialize.Serializer
	Storage            core.UserSessioner
	AppAddress         *url.URL
	GoogleClientID     string
	GoogleClientSecret string
	SessionSecret      string
}

// Handlers build http handler for service API
func Handlers(options Options) http.Handler {
	auth := authController{
		Serializer: options.Serializer,
		Storage:    options.Storage,
	}

	m := pat.New()
	if options.AppAddress.Path != "" {
		r := mux.NewRouter()
		sub := r.PathPrefix(options.AppAddress.Path).Subrouter()
		m.Router = *sub
	}

	m.Get("/who/", auth.who)
	m.Get("/providers/", auth.providers)
	m.Get("/logout/", auth.logout)
	m.Get("/login/{provider}/callback", auth.authCallback).Name(callbackRoute)
	m.Get("/login/{provider}/", gothic.BeginAuthHandler).Name(loginRoute)

	// Use our secret key
	gothic.AppKey = options.SessionSecret
	gothic.Store = sessions.NewCookieStore([]byte(gothic.AppKey))

	// Set up the provider(s)
	gplusLogin, gplusCallback := urlForProvider(options.AppAddress, m, "gplus")
	auth.Providers = map[string]string{"gplus": gplusLogin}
	goth.UseProviders(
		gplus.New(options.GoogleClientID, options.GoogleClientSecret, gplusCallback),
	)

	return m
}

func urlForProvider(appUrl *url.URL, router *pat.Router, providerName string) (login string, callback string) {
	loginUrl, _ := router.GetRoute(loginRoute).URL("provider", providerName)
	callbackUrl, _ := router.GetRoute(callbackRoute).URL("provider", providerName)

	// Make it absolute
	callbackUrl = appUrl.ResolveReference(callbackUrl)

	return loginUrl.String(), callbackUrl.String()
}
