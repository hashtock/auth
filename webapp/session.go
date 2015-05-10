package webapp

import (
	"errors"
	"log"
	"net/http"
	"time"

	"code.google.com/p/go-uuid/uuid"

	"github.com/hashtock/auth/core"
	"github.com/hashtock/auth/storage"
)

const (
	SessionName   = "auth_session_id"
	SessionTimout = 7 * 24 * time.Hour
)

var (
	ErrUserNotLoggedIn = errors.New("User not logged in")
)

func getSessionId(req *http.Request) string {
	cookie, err := req.Cookie(SessionName)
	if err != nil {
		return ""
	}

	return cookie.Value
}

func getCurrentUser(req *http.Request, store storage.UserSessioner) (user *core.User, err error) {
	sessionId := getSessionId(req)
	if sessionId == "" {
		err = ErrUserNotLoggedIn
		return
	}

	user, err = store.GetUser(sessionId)
	return
}

func setCurrentUser(rw http.ResponseWriter, req *http.Request, user *core.User, store storage.UserSessioner) (err error) {
	sessionId := getSessionId(req)
	if sessionId == "" {
		sessionId = uuid.New()
	}

	if err = store.SetUser(sessionId, user); err != nil {
		return
	}

	cookie := http.Cookie{
		Name:    SessionName,
		Value:   sessionId,
		Path:    "/",
		MaxAge:  int(SessionTimout.Seconds()),
		Expires: time.Now().Add(SessionTimout),
	}
	http.SetCookie(rw, &cookie)

	return
}

func removeSession(rw http.ResponseWriter, req *http.Request, store storage.UserSessioner) {
	sessionId := getSessionId(req)
	if sessionId == "" {
		// No session - nothing to do
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

	if err := store.DelUser(sessionId); err != nil {
		log.Printf("Could not remove session %v from storage.", err)
	}
	return
}
