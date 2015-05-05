package webapp

import (
	"fmt"
	"net/http"
	"text/template"

	"code.google.com/p/go-uuid/uuid"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"

	"github.com/hashtock/service-tools/serialize"
)

var (
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
	Serializer   serialize.Serializer
	Providers    map[string]string
}

type LoginProvider struct {
	Name string `json:"name"`
	URI  string `json:"uri"`
}

//////////////
// Handlers //
//////////////

func (a *authService) who(rw http.ResponseWriter, req *http.Request) {
	user, ok := a.getCurrentUser(req)
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

func (a *authService) logout(rw http.ResponseWriter, req *http.Request) {
	a.Serializer.JSON(rw, http.StatusOK, "Noop - ToDo")
}

/////////////
// Helpers //
/////////////

func getSessionId(session *sessions.Session) string {
	sessionId := ""
	if value, ok := session.Values[sessionIdKey]; ok {
		sessionId = value.(string)
	}
	return sessionId
}

func (a *authService) getCurrentUser(req *http.Request) (user goth.User, ok bool) {
	session, err := a.SessionStore.Get(req, a.SessionName)
	if err != nil {
		ok = false
		return
	}

	sessionId := getSessionId(session)
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

	sessionId := getSessionId(session)
	if sessionId == "" {
		sessionId = uuid.New()
	}

	users[sessionId] = user
	session.Values[sessionIdKey] = sessionId

	session.Save(req, rw)
}
