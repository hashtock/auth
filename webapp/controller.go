package webapp

import (
	"errors"
	"log"
	"net/http"
	"time"

	"code.google.com/p/go-uuid/uuid"
	"github.com/hashtock/service-tools/serialize"
	"github.com/markbates/goth/gothic"

	"github.com/hashtock/auth/core"
)

const (
	SessionName   = "auth_session_id"
	SessionTimout = 7 * 24 * time.Hour
)

var (
	ErrUserNotLoggedIn = errors.New("User not logged in")
)

type authController struct {
	Serializer serialize.Serializer
	Storage    core.UserSessioner
	Providers  map[string]string
}

func getSessionId(req *http.Request) string {
	cookie, err := req.Cookie(SessionName)
	if err != nil {
		return ""
	}

	return cookie.Value
}

func (a *authController) who(rw http.ResponseWriter, req *http.Request) {
	sessionId := getSessionId(req)
	if sessionId == "" {
		a.Serializer.JSON(rw, http.StatusUnauthorized, ErrUserNotLoggedIn)
		return
	}

	user, err := a.Storage.GetUserBySession(sessionId)
	if err != nil {
		errCode := http.StatusInternalServerError
		if err == core.ErrSessionNotFound {
			err = ErrUserNotLoggedIn
			errCode = http.StatusUnauthorized
		}

		a.Serializer.JSON(rw, errCode, err)
		return
	}

	a.Serializer.JSON(rw, http.StatusOK, user)
}

func (a *authController) providers(rw http.ResponseWriter, req *http.Request) {
	a.Serializer.JSON(rw, http.StatusOK, a.Providers)
}

func (a *authController) authCallback(rw http.ResponseWriter, req *http.Request) {
	authUser, err := gothic.CompleteUserAuth(rw, req)
	if err != nil {
		a.Serializer.JSON(rw, http.StatusInternalServerError, err)
		return
	}

	user := &core.User{
		Name:   authUser.Name,
		Email:  authUser.Email,
		Avatar: authUser.AvatarURL,
	}

	sessionId := getSessionId(req)
	if sessionId == "" {
		sessionId = uuid.New()
	}

	if err := a.Storage.AddUserToSession(sessionId, user); err != nil {
		a.Serializer.JSON(rw, http.StatusInternalServerError, err)
		return
	}

	cookie := http.Cookie{
		Name:     SessionName,
		Value:    sessionId,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   int(SessionTimout.Seconds()),
		Expires:  time.Now().Add(SessionTimout),
	}
	http.SetCookie(rw, &cookie)

	a.Serializer.JSON(rw, http.StatusOK, user)
}

func (a *authController) logout(rw http.ResponseWriter, req *http.Request) {
	sessionId := getSessionId(req)
	if sessionId == "" {
		// No session - nothing to do
		rw.WriteHeader(http.StatusOK)
		return
	}

	cookie := http.Cookie{
		Name:    SessionName,
		Value:   "",
		Path:    "/",
		MaxAge:  -1,
		Expires: time.Now().Add(-time.Hour),
	}
	http.SetCookie(rw, &cookie)

	if err := a.Storage.DeleteSession(sessionId); err != nil {
		// While cleanup operation failed, user session is gone now, so continue
		log.Printf("Could not remove session %v from storage.", err)
	}

	rw.WriteHeader(http.StatusOK)
}
