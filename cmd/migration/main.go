package main

import (
	"log"

	"gitlab.com/eyeo/network-filtering/router-adfilter-go/internal/db"
)

func main() {
	db := db.NewDatabase()
	
	err := db.Ping()
	if err != nil {
		log.Panic(err)
	}

	createTables := `
		-- Table for categories
		CREATE TABLE category (
			id SERIAL PRIMARY KEY,      -- Auto-incremented unique ID for each category
			name VARCHAR(255) NOT NULL, -- Category name
			description TEXT            -- Category description
		);

		-- Table for domains
		CREATE TABLE domain (
			domain_name VARCHAR(255) PRIMARY KEY -- Domain name as the primary key
		);

		-- Junction table for domain and category relationship (many-to-many)
		CREATE TABLE domain_category (
			domain_name VARCHAR(255),    -- Foreign key reference to domain
			category_id INT,             -- Foreign key reference to category
			PRIMARY KEY (domain_name, category_id),  -- Composite primary key for uniqueness
			FOREIGN KEY (domain_name) REFERENCES domain(domain_name) ON DELETE CASCADE,
			FOREIGN KEY (category_id) REFERENCES category(id) ON DELETE CASCADE
		);

	`

	result, err := db.Exec(createTables)
	if err != nil {
		log.Panicln(err)
	}
	log.Println(result)
	
	db.Close()	
}