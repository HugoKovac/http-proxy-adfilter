package data

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"gitlab.com/eyeo/network-filtering/router-adfilter-go/internal/types"
)


func ensureCategoryExists(db *sql.DB, categoryName, description string) (string, error) {
	// Check if the category exists
	var existingDescription string
	err := db.QueryRow("SELECT description FROM category WHERE category_name = $1", categoryName).Scan(&existingDescription)

	if err == sql.ErrNoRows {
		// If not exists, insert the new category
		_, err := db.Exec("INSERT INTO category (category_name, description) VALUES ($1, $2)", categoryName, description)
		if err != nil {
			return "", fmt.Errorf("failed to insert category: %w", err)
		}
		log.Printf("Inserted new category: %s\n", categoryName)
		return categoryName, nil
	} else if err != nil {
		return "", fmt.Errorf("failed to query category: %w", err)
	}

	// If the category already exists, return the existing category name
	return categoryName, nil
}

func insertDomain(db *sql.DB, domainName, categoryName string) error {
	// Check if the domain already exists
	var existingDomain string
	err := db.QueryRow("SELECT domain_name FROM domain WHERE domain_name = $1", domainName).Scan(&existingDomain)

	if err == sql.ErrNoRows {
		// If not exists, insert the new domain
		_, err := db.Exec("INSERT INTO domain (domain_name) VALUES ($1)", domainName)
		if err != nil {
			return fmt.Errorf("failed to insert domain: %w", err)
		}
		log.Printf("Inserted new domain: %s\n", domainName)
	}

	// Now insert the domain-category association in the junction table
	_, err = db.Exec("INSERT INTO domain_category (domain_name, category_name) VALUES ($1, $2) ON CONFLICT DO NOTHING", domainName, categoryName)
	if err != nil {
		return fmt.Errorf("failed to associate domain with category: %w", err)
	}

	return nil
}

func fakeFetch() (data []types.DomainList, err error) {
	file, err := os.Open("./tests/gambling_list.json")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err		
	}

	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	return data, nil
}

func GetCategorizedDomainList(db *sql.DB) {
	domainLists, err := fakeFetch()
	if err != nil {
		log.Println(err)
		return
	}
	
	var wg sync.WaitGroup

	for _, d := range domainLists {
		wg.Add(1)

		go func(d types.DomainList) {
			defer wg.Done()

			categoryID, err := ensureCategoryExists(db, d.Name, d.Description)
			if err != nil {
				log.Printf("Error ensuring category exists: %v\n", err)
				return
			}

			for _, domain := range d.List {
				err := insertDomain(db, domain, categoryID)
				if err != nil {
					log.Printf("Error inserting domain %s: %v\n", domain, err)
				} else {
					log.Printf("Successfully associated domain %s with category %s\n", domain, d.Name)
				}
			}
		}(d) 
	}

	wg.Wait()
}
