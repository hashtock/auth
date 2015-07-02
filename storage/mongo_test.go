package storage_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2"

	"github.com/hashtock/auth/core"
	"github.com/hashtock/auth/storage"
)

const (
	dbUrl = "127.0.0.1"
)

func init() {
	storage.DialTimout = 200 * time.Millisecond
}

func newStorage(t *testing.T) (*storage.MgoStorage, string) {
	DBName := fmt.Sprintf("test_db_%v", time.Now().UnixNano())
	mgoStorage, err := storage.NewMongoStorage(dbUrl, DBName)

	if err != nil {
		t.Fatal(err)
	}

	return mgoStorage, DBName
}

func destroyStorage(t *testing.T, dbName string) {
	msession, err := mgo.DialWithTimeout(dbUrl, 1*time.Second)
	assert.NoError(t, err)

	session := msession.Copy()
	err = session.DB(dbName).DropDatabase()
	assert.NoError(t, err)
}

func TestMongoStorageInstance(t *testing.T) {
	mgoStorage, dbName := newStorage(t)
	defer destroyStorage(t, dbName)

	assert.NotNil(t, mgoStorage)
	assert.Implements(t, (*core.UserSessioner)(nil), mgoStorage)
}

func TestMongoStorageMissingUrl(t *testing.T) {
	mgoStorage, err := storage.NewMongoStorage("", "test_db_123")

	assert.Nil(t, mgoStorage)
	assert.EqualError(t, err, storage.ErrDBUrlMissing.Error())
}

func TestMongoStorageMissingDBName(t *testing.T) {
	mgoStorage, err := storage.NewMongoStorage(dbUrl, "")

	assert.Nil(t, mgoStorage)
	assert.EqualError(t, err, storage.ErrDBNameMissing.Error())
}

func TestMongoStorageWrongUrl(t *testing.T) {
	if testing.Short() {
		t.Skip("Skip the test as waiting for bad url takes some time")
	}

	mgoStorage, err := storage.NewMongoStorage("fake-url", "test_db_123")

	assert.Nil(t, mgoStorage)
	assert.Error(t, err, "no reachable servers")
}

func TestSingleSession(t *testing.T) {
	mgoStorage, dbName := newStorage(t)
	defer destroyStorage(t, dbName)

	user := &core.User{
		Name:  "Name",
		Email: "bob@example.com",
	}
	mgoStorage.AddUserToSession("session-1", user)
	fetchedUser, err := mgoStorage.GetUserBySession("session-1")
	assert.NoError(t, err)
	assert.EqualValues(t, user, fetchedUser)
}

func TestMultipleSessions(t *testing.T) {
	mgoStorage, dbName := newStorage(t)
	defer destroyStorage(t, dbName)

	user := &core.User{
		Name:  "Name",
		Email: "bob@example.com",
	}
	mgoStorage.AddUserToSession("session-1", user)
	mgoStorage.AddUserToSession("session-2", user)
	fetchedUser1, err1 := mgoStorage.GetUserBySession("session-1")
	fetchedUser2, err2 := mgoStorage.GetUserBySession("session-2")
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.EqualValues(t, user, fetchedUser1)
	assert.EqualValues(t, user, fetchedUser2)
}

func TestSingleSessionDeleting(t *testing.T) {
	mgoStorage, dbName := newStorage(t)
	defer destroyStorage(t, dbName)

	user := &core.User{
		Name:  "Name",
		Email: "bob@example.com",
	}
	mgoStorage.AddUserToSession("session-1", user)
	fetchedUser1, err1 := mgoStorage.GetUserBySession("session-1")
	assert.NoError(t, err1)
	assert.EqualValues(t, user, fetchedUser1)

	delErr := mgoStorage.DeleteSession("session-1")
	assert.NoError(t, delErr)

	_, err2 := mgoStorage.GetUserBySession("session-1")
	assert.Error(t, err2, core.ErrSessionNotFound.Error())
}

func TestMultipleSessionDeletingOne(t *testing.T) {
	mgoStorage, dbName := newStorage(t)
	defer destroyStorage(t, dbName)

	user := &core.User{
		Name:  "Name",
		Email: "bob@example.com",
	}
	mgoStorage.AddUserToSession("session-1", user)
	mgoStorage.AddUserToSession("session-2", user)
	fetchedUser1, err1 := mgoStorage.GetUserBySession("session-1")
	assert.NoError(t, err1)
	assert.EqualValues(t, user, fetchedUser1)

	delErr := mgoStorage.DeleteSession("session-1")
	assert.NoError(t, delErr)

	fetchedUser2, err2 := mgoStorage.GetUserBySession("session-2")
	assert.NoError(t, err2)
	assert.EqualValues(t, user, fetchedUser2)
}

func TestAddingMultipleSessionsDoesNotCreateMultipleUsers(t *testing.T) {
	mgoStorage, dbName := newStorage(t)
	defer destroyStorage(t, dbName)

	user := &core.User{
		Name:  "Name",
		Email: "bob@example.com",
		Admin: false,
	}

	mgoStorage.AddUserToSession("session-1", user)
	mgoStorage.MakeUserAnAdmin(user.Email)
	mgoStorage.AddUserToSession("session-2", user)
	fetchedUser1, err1 := mgoStorage.GetUserBySession("session-1")
	fetchedUser2, err2 := mgoStorage.GetUserBySession("session-2")
	assert.NoError(t, err1)
	assert.NoError(t, err2)

	user.Admin = true
	assert.EqualValues(t, user, fetchedUser1)
	assert.EqualValues(t, user, fetchedUser2)
}

func TestMarkAsAdmin(t *testing.T) {
	mgoStorage, dbName := newStorage(t)
	defer destroyStorage(t, dbName)

	user := &core.User{
		Name:  "Name",
		Email: "bob@example.com",
		Admin: false,
	}
	mgoStorage.AddUserToSession("session-1", user)
	mgoStorage.MakeUserAnAdmin(user.Email)
	adminUser, err := mgoStorage.GetUserBySession("session-1")
	assert.NoError(t, err)
	assert.True(t, adminUser.Admin)
}
