package webapp

import (
	"net/http"
	"time"

	"code.google.com/p/go-uuid/uuid"
	"github.com/markbates/goth"
)

const (
	SessionName   = "auth_session_id"
	SessionTimout = 7 * 24 * time.Hour
)

func getSessionId(req *http.Request) string {
	cookie, err := req.Cookie(SessionName)
	if err != nil {
		return ""
	}

	return cookie.Value
}

func getCurrentUser(req *http.Request) (user goth.User, ok bool) {
	sessionId := getSessionId(req)
	if sessionId == "" {
		ok = false
		return
	}

	user, ok = users[sessionId]
	return
}

func setCurrentUser(rw http.ResponseWriter, req *http.Request, user goth.User) {
	sessionId := getSessionId(req)
	if sessionId == "" {
		sessionId = uuid.New()
	}

	users[sessionId] = user

	cookie := http.Cookie{
		Name:    SessionName,
		Value:   sessionId,
		Path:    "/",
		MaxAge:  int(SessionTimout.Seconds()),
		Expires: time.Now().Add(SessionTimout),
	}
	http.SetCookie(rw, &cookie)
}

func removeSession(rw http.ResponseWriter, req *http.Request) {
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
}
