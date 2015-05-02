package webapp

import (
	"fmt"
	"net/http"
	"text/template"

	"code.google.com/p/go-uuid/uuid"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
)

var (
	loginTemplate  = `<p><a href="/auth/gplus/">Log in with Google</a></p>`
	oldUserTemplte = `<p>Wecome back {{ .Name }}</p>`
	newUserTemplte = `<p>Wecome {{ .Name }}</p>`
)

const sessionIdKey string = "session_id"

var users map[string]goth.User

func init() {
	users = make(map[string]goth.User, 0)
}

type authService struct {
	SessionName  string
	SessionStore sessions.Store
}

func (a *authService) getCurrentUser(req *http.Request) (user goth.User, ok bool) {
	session, err := a.SessionStore.Get(req, a.SessionName)
	if err != nil {
		ok = false
		return
	}

	sessionId := ""
	if value, ok := session.Values[sessionIdKey]; ok {
		sessionId = value.(string)
	}

	if sessionId == "" {
		ok = false
		return
	}

	user, ok = users[sessionId]
	return
}

func (a *authService) setCurrentUser(rw http.ResponseWriter, req *http.Request, user goth.User) {
	session, err := a.SessionStore.Get(req, a.SessionName)
	if err != nil {
		return
	}

	sessionId := ""
	if value, ok := session.Values[sessionIdKey]; ok {
		sessionId = value.(string)
	}

	if sessionId == "" {
		sessionId = uuid.New()
	}

	users[sessionId] = user
	session.Values[sessionIdKey] = sessionId

	session.Save(req, rw)
}

func (a *authService) authCallback(rw http.ResponseWriter, req *http.Request) {
	if currentUser, ok := a.getCurrentUser(req); ok {
		t, _ := template.New("old").Parse(oldUserTemplte)
		t.Execute(rw, currentUser)
		return
	}

	user, err := gothic.CompleteUserAuth(rw, req)
	if err != nil {
		fmt.Fprintln(rw, err)
		return
	}

	a.setCurrentUser(rw, req, user)

	t, _ := template.New("new").Parse(newUserTemplte)
	t.Execute(rw, user)
}

func (a *authService) index(rw http.ResponseWriter, req *http.Request) {
	user, ok := a.getCurrentUser(req)
	if ok {
		t, _ := template.New("old").Parse(oldUserTemplte)
		t.Execute(rw, user)
	} else {
		t, _ := template.New("login").Parse(loginTemplate)
		t.Execute(rw, nil)
	}
}
