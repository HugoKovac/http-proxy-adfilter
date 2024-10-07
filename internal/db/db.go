package db

import (
	"log"
	"time"

	_ "modernc.org/sqlite"
	"github.com/boltdb/bolt"
)

const file string = "activities.db"

func NewDatabase() (boltdb *bolt.DB) {
	var err error

	boltdb, err = bolt.Open("my.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal(err)
	}

	return boltdb
}
