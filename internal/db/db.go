package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

// TODO: pull from cloud vault
const (
	host     = "localhost"
	port     = 5432
	user     = "user-name"
	password = "strong-password"
	dbname   = "postgres"
  )

func NewDatabase() (db *sql.DB) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to DB")

	return db
}