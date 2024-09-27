package data

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"gitlab.com/eyeo/network-filtering/router-adfilter-go/internal/types"
)


func ensureCategoryExists(db *sql.DB, name, description string) (int, error) {
	var categoryID int
	err := db.QueryRow(`SELECT id FROM category WHERE name = $1`, name).Scan(&categoryID)
	if err != nil {
		if err == sql.ErrNoRows {
			// Insert new category if it does not exist
			err = db.QueryRow(
				`INSERT INTO category (name, description) VALUES ($1, $2) RETURNING id`,
				name, description,
			).Scan(&categoryID)
			if err != nil {
				return 0, fmt.Errorf("could not insert category: %v", err)
			}
		} else {
			return 0, fmt.Errorf("error fetching category: %v", err)
		}
	}
	return categoryID, nil
}

func insertDomain(db *sql.DB, domain string, categoryID int) error {
	// Insert domain if it doesn't exist
	_, err := db.Exec(`INSERT INTO domain (domain_name) VALUES ($1) ON CONFLICT (domain_name) DO NOTHING`, domain)
	if err != nil {
		return fmt.Errorf("could not insert domain: %v", err)
	}

	// Associate the domain with the category
	_, err = db.Exec(`INSERT INTO domain_category (domain_name, category_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, domain, categoryID)
	if err != nil {
		return fmt.Errorf("could not associate domain with category: %v", err)
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
	for _, d := range domainLists {
		categoryID, err := ensureCategoryExists(db, d.Name, d.Description)
		if err != nil {
			log.Printf("Error ensuring category exists: %v\n", err)
			continue
		}

		for _, domain := range d.List {
			err := insertDomain(db, domain, categoryID)
			if err != nil {
				log.Printf("Error inserting domain %s: %v\n", domain, err)
			} else {
				log.Printf("Successfully associated domain %s with category %s\n", domain, d.Name)
			}
		}
	}
}
