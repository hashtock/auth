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
	dbUrl = "localhost"
)

func init() {
	storage.DialTimout = 200 * time.Microsecond
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
		t.Skip("Skip the test as waitint for bad url takes some time")
	}

	mgoStorage, err := storage.NewMongoStorage("fake-url", "test_db_123")

	assert.Nil(t, mgoStorage)
	assert.Error(t, err, "no reachable servers")
}
