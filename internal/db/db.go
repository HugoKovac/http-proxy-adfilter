package db

import (
	"database/sql"
	"log"
	"time"

	_ "modernc.org/sqlite"
	"github.com/boltdb/bolt"
)

const file string = "activities.db"

func NewDatabase() (db *sql.DB, boltdb *bolt.DB) {
	var err error
	db, err = sql.Open("sqlite", file)

	if err != nil {
		log.Fatal(err)
	}
	const create string = `
		CREATE TABLE IF NOT EXISTS activities (
			id INTEGER NOT NULL PRIMARY KEY,
			time DATETIME NOT NULL,
			description TEXT
			);`
	if _, err := db.Exec(create); err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec("PRAGMA synchronous = OFF;")
	if err != nil {
		log.Fatal("Failed to set synchronous mode:", err)
	}

	_, err = db.Exec("PRAGMA journal_mode = WAL;")
	if err != nil {
		log.Fatal("Failed to set journal mode:", err)
	}

	_, err = db.Exec("PRAGMA cache_size = -20000;") // Adjust size as needed
	if err != nil {
		log.Fatal("Failed to set cache size:", err)
	}

	log.Println("Connected to DB")

	boltdb, err = bolt.Open("my.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal(err)
	}

	return db, boltdb
}
