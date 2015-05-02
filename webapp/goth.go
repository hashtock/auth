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

func setupGoth(cfg *conf.Config, callbackRoute *mux.Route) error {
	gothic.GetState = getState
	gothic.GetProviderName = getProviderName
	gothic.AppKey = cfg.SessionSecret
	gothic.Store = sessions.NewCookieStore([]byte(gothic.AppKey))

	path, err := callbackRoute.URL("provider", "gplus")
	if err != nil {
		return err
	}

	googleAuthCallback, err := cfg.AppAddress.Parse(path.String())
	if err != nil {
		return err
	}

	goth.UseProviders(
		gplus.New(
			cfg.GoogleClientID,
			cfg.GoogleClientSecret,
			googleAuthCallback.String(),
		),
	)

	return nil
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
