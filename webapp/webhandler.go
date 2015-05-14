package webapp

import (
	"net/http"
	"net/url"

	"github.com/gorilla/pat"
	"github.com/gorilla/sessions"
	"github.com/hashtock/service-tools/serialize"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/gplus"

	"github.com/hashtock/auth/conf"
	"github.com/hashtock/auth/storage"
)

const (
	loginRoute    = "login"
	callbackRoute = "callback"
)

// Handlers build http handler for service API
func Handlers(cfg *conf.Config, sessioner storage.UserSessioner) http.Handler {
	auth := authService{
		Serializer: serialize.WebAPISerializer{},
		Storage:    sessioner,
	}

	m := pat.New()

	m.Get("/who/", auth.who)
	m.Get("/providers/", auth.providers)
	m.Get("/logout/", auth.logout)
	m.Get("/login/{provider}/callback", auth.authCallback).Name(callbackRoute)
	m.Get("/login/{provider}/", gothic.BeginAuthHandler).Name(loginRoute)

	// Use our secret key
	gothic.AppKey = cfg.SessionSecret
	gothic.Store = sessions.NewCookieStore([]byte(gothic.AppKey))

	// Set up the provider(s)
	gplusLogin, gplusCallback := urlForProvider(cfg.AppAddress, m, "gplus")
	auth.Providers = map[string]string{"gplus": gplusLogin}
	goth.UseProviders(
		gplus.New(cfg.GoogleClientID, cfg.GoogleClientSecret, gplusCallback),
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
