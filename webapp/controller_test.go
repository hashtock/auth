package webapp_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/gplus"
	"github.com/stretchr/testify/assert"

	"github.com/hashtock/auth/core"
	"github.com/hashtock/auth/webapp"
)

func TestWhoNotLoggedIn(t *testing.T) {
	handler, serializer, _ := makeHandler()
	w := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", "/who/", nil)
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, `"User not logged in"`, w.Body.String())
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

func TestWhoWrongSession(t *testing.T) {
	handler, serializer, _ := makeHandler()
	w := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", "/who/", nil)
	req.AddCookie(&http.Cookie{Name: webapp.SessionName, Value: "fake-session"})
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, `"User not logged in"`, w.Body.String())
	assert.EqualValues(t, webapp.ErrUserNotLoggedIn, serializer.obj)
}

func TestWhoOtherError(t *testing.T) {
	handler, serializer, storage := makeHandler()
	w := httptest.NewRecorder()

	err := errors.New("Some other error")
	storage.NextError = err

	req, _ := http.NewRequest("GET", "/who/", nil)
	req.AddCookie(&http.Cookie{Name: webapp.SessionName, Value: "fake-session"})
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, `"Some other error"`, w.Body.String())
	assert.EqualValues(t, err, serializer.obj)
}

func TestProviders(t *testing.T) {
	handler, serializer, _ := makeHandler()
	w := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", "/providers/", nil)
	handler.ServeHTTP(w, req)

	expectedProviders := map[string]string{
		"gplus": "/login/gplus/",
	}

	assert.Equal(t, http.StatusOK, w.Code)
	assert.EqualValues(t, expectedProviders, serializer.obj)
}

func TestProvidersAsRelativeToPath(t *testing.T) {
	handler, serializer, _ := makeHandlerSubPath("/some/path/")
	w := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", "/some/path/providers/", nil)
	handler.ServeHTTP(w, req)

	expectedProviders := map[string]string{
		"gplus": "/some/path/login/gplus/",
	}

	assert.Equal(t, http.StatusOK, w.Code)
	assert.EqualValues(t, expectedProviders, serializer.obj)
}

func TestLoginGoogle(t *testing.T) {
	handler, _, _ := makeHandler()
	provider, _ := goth.GetProvider("gplus")
	w := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", "/login/gplus/", nil)
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
	assert.Contains(t, w.HeaderMap["Location"][0], "https://accounts.google.com/o/oauth2/auth")
	assert.Equal(t, provider.(*gplus.Provider).CallbackURL, "http://localhost:1234/login/gplus/callback")
}

func TestLoginGoogleCallbackURL(t *testing.T) {
	handler, _, _ := makeHandlerSubPath("/auth/path/")
	provider, _ := goth.GetProvider("gplus")
	w := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", "/auth/path/login/gplus/", nil)
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
	assert.Contains(t, w.HeaderMap["Location"][0], "https://accounts.google.com/o/oauth2/auth")
	assert.Equal(t, provider.(*gplus.Provider).CallbackURL, "http://localhost:1234/auth/path/login/gplus/callback")
}

func TestLoginFakeProvider(t *testing.T) {
	handler, serializer, storage := makeHandler()

	flow := func(handler http.Handler, nextError error) *httptest.ResponseRecorder {
		// Login
		wLogin := httptest.NewRecorder()

		provider := &testProvider{NextError: nextError}
		goth.UseProviders(provider)

		req, _ := http.NewRequest("GET", "/login/faux/", nil)
		handler.ServeHTTP(wLogin, req)

		assert.Equal(t, http.StatusTemporaryRedirect, wLogin.Code)
		assert.Equal(t, wLogin.HeaderMap["Location"][0], "http://example.com/auth/")

		// Callback
		wCallback := httptest.NewRecorder()

		reqCallback, _ := http.NewRequest("GET", "/login/faux/callback", nil)
		reqCallback.Header["Cookie"] = wLogin.HeaderMap["Set-Cookie"]
		handler.ServeHTTP(wCallback, reqCallback)

		return wCallback
	}

	// Error in auth
	wAuthErr := flow(handler, errors.New("Auth error"))
	assert.EqualValues(t, errors.New("Auth error"), serializer.obj)
	assert.Equal(t, http.StatusInternalServerError, wAuthErr.Code)

	// Error in auth
	storage.NextError = errors.New("Storage error")
	wStorErr := flow(handler, nil)
	assert.EqualValues(t, errors.New("Storage error"), serializer.obj)
	assert.Equal(t, http.StatusInternalServerError, wStorErr.Code)

	// All ok
	wOk := flow(handler, nil)
	user := &core.User{Name: "name", Email: "email", Avatar: ""}
	assert.EqualValues(t, user, serializer.obj)
	assert.Equal(t, http.StatusOK, wOk.Code)
	cookie := wOk.HeaderMap["Set-Cookie"][0]
	assert.Contains(t, cookie, webapp.SessionName)
	assert.Contains(t, cookie, "Path=/;")
	assert.Contains(t, cookie, "HttpOnly")
	assert.Contains(t, cookie, "Max-Age=604800;")
}

func TestLogout(t *testing.T) {
	handler, _, storage := makeHandler()
	w := httptest.NewRecorder()

	sessionId := "session"
	user := &core.User{Name: "user", Email: "email", Avatar: "avatar"}
	storage.AddUserToSession(sessionId, user)

	// Sanity check
	assert.Len(t, storage.Data, 1)

	req, _ := http.NewRequest("GET", "/logout/", nil)
	req.AddCookie(&http.Cookie{Name: webapp.SessionName, Value: sessionId})
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.HeaderMap["Set-Cookie"][0], "auth_session_id=;") // Session id cleared from cookies
	assert.Len(t, storage.Data, 0)                                        // Session id removed also from storage
}

func TestLogoutUserWithoutSession(t *testing.T) {
	handler, _, _ := makeHandler()
	w := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", "/logout/", nil)
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Len(t, w.HeaderMap["Set-Cookie"], 0) // Cookies untouched
}

func TestLogoutErrorInStorage(t *testing.T) {
	handler, _, storage := makeHandler()
	w := httptest.NewRecorder()

	sessionId := "session"
	user := &core.User{Name: "user", Email: "email", Avatar: "avatar"}
	storage.AddUserToSession(sessionId, user)

	// Sanity check
	assert.Len(t, storage.Data, 1)
	storage.NextError = errors.New("Some other error")

	req, _ := http.NewRequest("GET", "/logout/", nil)
	req.AddCookie(&http.Cookie{Name: webapp.SessionName, Value: sessionId})
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Len(t, storage.Data, 1)                                        // Session id still there
	assert.Contains(t, w.HeaderMap["Set-Cookie"][0], "auth_session_id=;") // But session id cleared from cookies
}
