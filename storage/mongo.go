package storage

import (
	"errors"
	"fmt"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/hashtock/auth/core"
)

const (
	userColection = "user"
)

var (
	ErrDBUrlMissing  = errors.New("Url to Mongodb not provided")
	ErrDBNameMissing = errors.New("Name of database for Mongodb not provided")

	DialTimout = 5 * time.Second
)

type MgoStorage struct {
	session *mgo.Session
	db      string
	dbName  string
}

type mongoUser struct {
	core.User `bson:",inline"`
	Session   []string `bson:"session"`
}

func NewMongoStorage(dbUrl string, dbName string) (*MgoStorage, error) {
	if dbUrl == "" {
		return nil, ErrDBUrlMissing
	} else if dbName == "" {
		return nil, ErrDBNameMissing
	}

	msession, err := mgo.DialWithTimeout(dbUrl, DialTimout)
	if err != nil {
		return nil, err
	}

	mgostorage := &MgoStorage{
		db:      dbUrl,
		dbName:  dbName,
		session: msession,
	}

	return mgostorage, nil
}

func (m *MgoStorage) userColection() *mgo.Collection {
	session := m.session.Copy()
	return session.DB(m.dbName).C(userColection)
}

func (m *MgoStorage) GetUserBySession(sessionId string) (*core.User, error) {
	col := m.userColection()
	defer col.Database.Session.Close()

	selector := bson.M{"session": sessionId}
	mUser := mongoUser{}

	err := col.Find(selector).One(&mUser)
	if err == mgo.ErrNotFound {
		err = core.ErrSessionNotFound
	}
	return &mUser.User, err
}

func (m *MgoStorage) AddUserToSession(sessionId string, user *core.User) (err error) {
	col := m.userColection()
	defer col.Database.Session.Close()

	selector := bson.M{"email": user.Email}

	cnt, err := col.Find(selector).Count()
	if err != nil {
		return err
	} else if cnt == 0 {
		// New user, need to create
		mUser := mongoUser{}
		mUser.User = *user
		mUser.Session = []string{sessionId}
		err = col.Insert(&mUser)
	} else {
		// User in place, try to add sessionId
		change := bson.M{
			"$addToSet": bson.M{
				"session": sessionId,
			},
		}

		err = col.Update(selector, change)
	}
	return err
}

func (m *MgoStorage) DeleteSession(sessionId string) error {
	col := m.userColection()
	defer col.Database.Session.Close()

	selector := bson.M{"session": sessionId}

	change := bson.M{
		"$pull": bson.M{
			"session": sessionId,
		},
	}

	err := col.Update(selector, change)
	if err == mgo.ErrNotFound {
		err = core.ErrSessionNotFound
	}
	return err
}

func (m *MgoStorage) MakeUserAnAdmin(email string) error {
	col := m.userColection()
	defer col.Database.Session.Close()

	selector := bson.M{"email": email}

	cnt, err := col.Find(selector).Count()
	if err != nil {
		return err
	} else if cnt == 0 {
		return fmt.Errorf("User with email %#v not found", email)
	}

	change := bson.M{
		"$set": bson.M{
			"admin": true,
		},
	}
	err = col.Update(selector, change)

	return err
}
