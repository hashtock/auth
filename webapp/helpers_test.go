package webapp_test

import (
	"net/http"
	"net/url"

	"github.com/hashtock/service-tools/serialize"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/faux"

	"github.com/hashtock/auth/core"
	"github.com/hashtock/auth/webapp"
)

////////////////////////////
// Storage implementation //
////////////////////////////

type mapStorage struct {
	Data      map[string]*core.User
	NextError error
}

func (m *mapStorage) nextError() (err error) {
	err = m.NextError
	m.NextError = nil
	return err
}

func (m *mapStorage) GetUserBySession(sessionId string) (*core.User, error) {
	if m.NextError != nil {
		return nil, m.nextError()
	}

	value, ok := m.Data[sessionId]
	if !ok {
		return nil, core.ErrSessionNotFound
	}
	return value, nil
}
func (m *mapStorage) AddUserToSession(sessionId string, user *core.User) error {
	if m.NextError != nil {
		return m.nextError()
	}

	m.Data[sessionId] = user
	return nil
}
func (m *mapStorage) DeleteSession(sessionId string) error {
	if m.NextError != nil {
		return m.nextError()
	}

	delete(m.Data, sessionId)
	return nil
}

////////////////
// Serializer //
////////////////

type serializerLog struct {
	obj interface{}
}

func (s *serializerLog) JSON(rw http.ResponseWriter, status int, obj interface{}) {
	s.obj = obj
	ser := serialize.WebAPISerializer{}
	ser.JSON(rw, status, obj)
}

////////////////////////
// Test auth provider //
////////////////////////

type testProvider struct {
	faux.Provider

	NextError error
}

func (t *testProvider) BeginAuth(state string) (goth.Session, error) {
	return &faux.Session{
		Name:  "name",
		Email: "email",
	}, nil
}

func (t *testProvider) UnmarshalSession(data string) (goth.Session, error) {
	if t.NextError != nil {
		return nil, t.NextError
	}

	return t.Provider.UnmarshalSession(data)
}

//////////////////
// Test handler //
//////////////////

func makeHandler() (http.Handler, *serializerLog, *mapStorage) {
	url, _ := url.Parse("http://localhost:1234")
	serializer := new(serializerLog)
	storage := &mapStorage{
		Data: make(map[string]*core.User),
	}

	options := webapp.Options{
		AppAddress: url,
		Storage:    storage,
		Serializer: serializer,
	}

	handler := webapp.Handlers(options)
	return handler, serializer, storage
}
