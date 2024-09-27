package main

import (
	"log"

	"gitlab.com/eyeo/network-filtering/router-adfilter-go/internal/db"
)

func main() {
	db := db.NewDatabase()
	
	if err := db.Ping(); err != nil {
		log.Panic(err)
	}

	createTables := `
		-- Create the client table
		CREATE TABLE client (
			client_mac MACADDR PRIMARY KEY,  -- MAC addresses and Index
			ip INET NOT NULL UNIQUE -- IPv4 support
		);

		-- Table for categories
		CREATE TABLE category (
			category_name VARCHAR(255) PRIMARY KEY, -- Category name and Index
			description TEXT            -- Category description
		);

		-- Table for domains
		CREATE TABLE domain (
			domain_name VARCHAR(255) PRIMARY KEY -- Domain name as the primary key
		);

		-- Junction table for domain and category relationship (many-to-many)
		CREATE TABLE domain_category (
			domain_name VARCHAR(255),    -- Foreign key reference to domain
			category_name VARCHAR(255),             -- Foreign key reference to category
			PRIMARY KEY (domain_name, category_name),  -- Composite primary key for uniqueness
			FOREIGN KEY (domain_name) REFERENCES domain(domain_name) ON DELETE CASCADE,
			FOREIGN KEY (category_name) REFERENCES category(category_name) ON DELETE CASCADE
		);

		-- Create the junction table for client-category association
		CREATE TABLE client_category (
			client_mac MACADDR NOT NULL,
			category_name VARCHAR(255) NOT NULL,
			PRIMARY KEY (client_mac, category_name),
			FOREIGN KEY (client_mac) REFERENCES client(client_mac) ON DELETE CASCADE,
			FOREIGN KEY (category_name) REFERENCES category(category_name) ON DELETE CASCADE
		);
	`

	result, err := db.Exec(createTables)
	if err != nil {
		log.Panicln(err)
	}
	log.Println(result)
	
	db.Close()	
}