package db

import (
	"os"

	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

type Query bson.M

func Connect() (*DB, error) {
	dbHost := os.Getenv("MONGO_URL")
	return New(dbHost)
}

func New(host string) (*DB, error) {
	session, err := mgo.Dial(host)
	if err != nil {
		return nil, err
	}

	session.SetMode(mgo.Monotonic, true)
	db := &DB{session: session}
	err = db.ensureDefaultIndex()

	return db, err
}

type DB struct {
	name    string
	session *mgo.Session
}

func (db *DB) Close() {
	db.session.Close()
}

func (db *DB) DB() *mgo.Database {
	return db.session.DB(db.name)
}

func (db *DB) C(name string) *mgo.Collection {
	return db.DB().C(name)
}

func (db *DB) DropC(name string) error {
	return db.C(name).DropCollection()
}

func (db *DB) Upsert(name string, q Query, v interface{}) (updated bool, err error) {
	info, err := db.C(name).Upsert(q, v)
	updated = info != nil && info.UpsertedId != nil
	return
}

func (db *DB) EnsureIndexKey(colName string, keys ...string) error {
	for _, key := range keys {
		index := mgo.Index{
			Key:        []string{key},
			Background: true,
		}

		err := db.C(colName).EnsureIndex(index)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *DB) EnsureUniqueIndexKey(colName string, keys ...string) error {
	for _, key := range keys {
		index := mgo.Index{
			Key:        []string{key},
			Unique:     true,
			DropDups:   true,
			Background: true,
		}

		err := db.C(colName).EnsureIndex(index)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *DB) ensureDefaultIndex() error {
	err := db.EnsureUniqueIndexKey("new_builds", "lastbuildid")
	if err != nil {
		return err
	}

	return db.EnsureIndexKey("new_builds", "lastbuildstartedat")
}
