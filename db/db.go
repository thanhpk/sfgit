package db

import (
	"github.com/boltdb/bolt"
	"os"
	"time"
)

type LocalDB struct {
	db *bolt.DB
	root string
	database []byte
}

func NewLocalDB(root, database string) *LocalDB {
	d := &LocalDB{}
	d.root = root
	d.database = []byte(database)

	var err error
	d.db, err = bolt.Open(root + "git.db", 0600, nil)
	if err != nil {
		panic(err)
	}

	err = d.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(d.database)
		return err
	})
	if err != nil {
		panic(err)
	}
	return d
}

func (m LocalDB) Close() {
	err := m.db.Close()
	if err != nil {
		panic(err)
	}
}

func (m LocalDB) IsRepoExists(service, repo string) bool {
	_, err := os.Stat(m.root + service + "/" +  repo + "/.git")
	return err == nil
}

func (m LocalDB) Touch(service, repo string) (err error) {
	var ut time.Time
	err = m.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(m.database)
		utb := b.Get([]byte(service + "/" + repo))
		ut, _ = time.Parse(time.RFC3339Nano, string(utb))
		return nil
	})

	if err != nil {
		return err
	}

	if time.Now().Sub(ut).Nanoseconds() < 0 {
		return nil
	}

	err = m.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(m.database)
		err := b.Put([]byte(service + "/" + repo), []byte(time.Now().Format(time.RFC3339Nano)))
		if err != nil {
			return err
		}
		return nil
	})

	return err
}

func (m LocalDB) LastUpdate(service, repo string) time.Time {
	var t time.Time
	err := m.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(m.database)
		utb := b.Get([]byte(service + "/" + repo))
		if len(utb) == 0 {
			t = time.Time{}
			return nil
		}
		ut, _ := time.Parse(time.RFC3339Nano, string(utb))
		t = ut
		return nil
	})
	if err != nil {
		return time.Time{}
	}
	return t
}
