package webapp

import (
	"fmt"
	"net/http"

	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"

	"github.com/hashtock/service-tools/serialize"
)

var users map[string]goth.User

func init() {
	users = make(map[string]goth.User, 0)
}

type authService struct {
	Serializer serialize.Serializer
	Providers  map[string]string
}

type LoginProvider struct {
	Name string `json:"name"`
	URI  string `json:"uri"`
}

//////////////
// Handlers //
//////////////

func (a *authService) who(rw http.ResponseWriter, req *http.Request) {
	user, ok := getCurrentUser(req)
	if !ok {
		a.Serializer.JSON(rw, http.StatusUnauthorized, nil)
		return
	}

	a.Serializer.JSON(rw, http.StatusOK, user)
}

func (a *authService) providers(rw http.ResponseWriter, req *http.Request) {
	providers := make([]LoginProvider, 0)

	for name, uri := range a.Providers {
		providers = append(providers, LoginProvider{name, uri})
	}

	a.Serializer.JSON(rw, http.StatusOK, providers)
}

func (a *authService) authCallback(rw http.ResponseWriter, req *http.Request) {
	user, err := gothic.CompleteUserAuth(rw, req)
	if err != nil {
		fmt.Fprintln(rw, err)
		return
	}

	setCurrentUser(rw, req, user)
	a.Serializer.JSON(rw, http.StatusOK, user)
}

func (a *authService) logout(rw http.ResponseWriter, req *http.Request) {
	removeSession(rw, req)
	rw.WriteHeader(http.StatusOK)
}
