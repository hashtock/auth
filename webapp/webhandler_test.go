package webapp_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/hashtock/service-tools/serialize"
	"github.com/stretchr/testify/assert"

	"github.com/hashtock/auth/core"
	"github.com/hashtock/auth/webapp"
)

type mapStorage struct {
	Data map[string]*core.User
}

func (m *mapStorage) GetUserBySession(sessionId string) (*core.User, error) {
	value, ok := m.Data[sessionId]
	if !ok {
		return nil, core.ErrSessionNotFound
	}
	return value, nil
}
func (m *mapStorage) AddUserToSession(sessionId string, user *core.User) error {
	m.Data[sessionId] = user
	return nil
}
func (m *mapStorage) DeleteSession(sessionId string) error {
	delete(m.Data, sessionId)
	return nil
}

type serializerLog struct {
	obj interface{}
}

func (s *serializerLog) JSON(rw http.ResponseWriter, status int, obj interface{}) {
	s.obj = obj
	ser := serialize.WebAPISerializer{}
	ser.JSON(rw, status, obj)
}

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

func TestWhoNotLoggedIn(t *testing.T) {
	handler, serializer, _ := makeHandler()
	w := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", "/who/", nil)
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, "{}", w.Body.String())
	assert.EqualValues(t, webapp.ErrUserNotLoggedIn, serializer.obj)
}

func TestWhoLoggedIn(t *testing.T) {
	handler, serializer, storage := makeHandler()
	w := httptest.NewRecorder()

	sessionId := "session"
	user := &core.User{Name: "user", Email: "email", Avatar: "avatar"}
	storage.AddUserToSession(sessionId, user)

	req, _ := http.NewRequest("GET", "/who/", nil)
	req.AddCookie(&http.Cookie{Name: webapp.SessionName, Value: sessionId})
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.EqualValues(t, user, serializer.obj)
}
