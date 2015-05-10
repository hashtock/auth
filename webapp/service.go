package webapp

import (
	"fmt"
	"net/http"

	"github.com/markbates/goth/gothic"

	"github.com/hashtock/auth/core"
	"github.com/hashtock/auth/storage"
	"github.com/hashtock/service-tools/serialize"
)

type authService struct {
	Serializer serialize.Serializer
	Storage    storage.UserSessioner
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
	user, err := getCurrentUser(req, a.Storage)
	if err != nil {
		a.Serializer.JSON(rw, http.StatusUnauthorized, err)
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
	authUser, err := gothic.CompleteUserAuth(rw, req)
	if err != nil {
		fmt.Fprintln(rw, err)
		return
	}

	user := core.User{
		Name:   authUser.Name,
		Email:  authUser.Email,
		Avatar: authUser.AvatarURL,
	}

	if err := setCurrentUser(rw, req, &user, a.Storage); err != nil {
		a.Serializer.JSON(rw, http.StatusInternalServerError, err)
		return
	}

	a.Serializer.JSON(rw, http.StatusOK, user)
}

func (a *authService) logout(rw http.ResponseWriter, req *http.Request) {
	removeSession(rw, req, a.Storage)
	rw.WriteHeader(http.StatusOK)
}
