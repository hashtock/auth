package storage

import (
	"errors"

	"github.com/hashtock/auth/core"
)

var (
	ErrSessionNotFound error = errors.New("Session not found")
)

type UserSessioner interface {
	GetUser(sessionId string) (*core.User, error)
	SetUser(sessionId string, user *core.User) error
	DelUser(sessionId string) error
}
