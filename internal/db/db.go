package db

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

const file string = "activities.db"

func NewDatabase() (db *sql.DB) {
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

	return db
}
