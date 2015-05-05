package webapp

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/gplus"

	"github.com/hashtock/auth/conf"
)

func setupGoth(cfg *conf.Config, loginRoute, callbackRoute *mux.Route) (map[string]string, error) {
	gothic.GetState = getState
	gothic.GetProviderName = getProviderName
	gothic.AppKey = cfg.SessionSecret
	gothic.Store = sessions.NewCookieStore([]byte(gothic.AppKey))

	providers := make(map[string]string, 0)

	// Google Auth
	const googleProvider string = "gplus"
	callbackPath, err := callbackRoute.URL("provider", googleProvider)
	if err != nil {
		return providers, err
	}

	googleAuthCallback, err := cfg.AppAddress.Parse(callbackPath.String())
	if err != nil {
		return providers, err
	}

	loginPath, err := loginRoute.URL("provider", googleProvider)
	if err != nil {
		return providers, err
	}
	providers[googleProvider] = loginPath.String()

	goth.UseProviders(
		gplus.New(
			cfg.GoogleClientID,
			cfg.GoogleClientSecret,
			googleAuthCallback.String(),
		),
	)

	return providers, nil
}

func getState(req *http.Request) string {
	return req.URL.Query().Get("state")
}

func getProviderName(req *http.Request) (string, error) {
	vars := mux.Vars(req)
	provider := vars["provider"]

	if provider == "" {
		return provider, errors.New("you must select a provider")
	}
	return provider, nil
}
