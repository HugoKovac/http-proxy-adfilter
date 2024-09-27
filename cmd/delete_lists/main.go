package main

import (
	"database/sql"
	"fmt"
	"log"

	"gitlab.com/eyeo/network-filtering/router-adfilter-go/internal/db"
)

func deleteAllData(db *sql.DB) error {
	// Get all tables in the public schema
	query := `
		SELECT tablename
		FROM pg_tables
		WHERE schemaname = 'public'
	`
	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("could not retrieve tables: %v", err)
	}
	defer rows.Close()

	// Collect all table names
	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return fmt.Errorf("could not scan table: %v", err)
		}
		tables = append(tables, table)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating over tables: %v", err)
	}

	// Truncate each table
	for _, table := range tables {
		truncateQuery := fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)
		_, err := db.Exec(truncateQuery)
		if err != nil {
			return fmt.Errorf("could not truncate table %s: %v", table, err)
		}
		log.Printf("Table %s truncated successfully.", table)
	}

	return nil
}


func main() {
	db := db.NewDatabase()
	
	if err := db.Ping(); err != nil {
		log.Panic(err)
	}

	defer db.Close()	

	err := deleteAllData(db)
	if err != nil {
		log.Fatalf("Error deleting all data: %v", err)
	}
	
}