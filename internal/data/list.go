package data

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	macClients "gitlab.com/eyeo/network-filtering/router-adfilter-go/internal/pkg/mac_clients"
	"gitlab.com/eyeo/network-filtering/router-adfilter-go/internal/types"
)

func GetSubscribedCategoryLists(db *sql.DB, mac string) (list []types.CategoryList, err error) {
	rows, err := db.Query(`SELECT cat.*
		FROM client c
		JOIN client_category cc on c.client_mac = cc.client_mac
		JOIN category cat on cc.category_name = cat.category_name
		WHERE c.client_mac = ?`, mac)
	if err != nil {
		return list, err
	}
	for rows.Next() {
		var index types.CategoryList

		err := rows.Scan(&index.CategoryName, &index.Description)
		if err != nil {
			log.Println(err)
		}
		list = append(list, index)
	} 
	return list, nil
	
}

func DelSubscribtion(db *sql.DB, category string, mac string) (err error){
	_, err = db.Exec(`DELETE FROM client_category
		WHERE client_mac = ? AND category_name = ?`, mac, category);
	return err
}

func EnsureClientExists(db *sql.DB, client macClients.Client) error {
	// Convert MAC and IP to strings
	macStr := client.MAC.String()
	ipStr := client.IP.String()

	// Check if the client already exists
	var existingMAC string
	log.Printf("Checking if client exists: MAC = %s, IP = %s\n", macStr, ipStr)

	err := db.QueryRow("SELECT client_mac FROM client WHERE client_mac = ?", macStr).Scan(&existingMAC)

	if err == sql.ErrNoRows {
		// If not exists, insert the new client
		_, err := db.Exec("INSERT INTO client (client_mac, ip) VALUES (?, ?)", macStr, ipStr)
		if err != nil {
			return fmt.Errorf("failed to insert client: %w", err)
		}
		log.Printf("Inserted new client: %s\n", macStr)
	} else if err != nil {
		return fmt.Errorf("failed to query client: %w", err)
	}

	// Client exists or has been inserted successfully
	return nil
}

func GetCategoryLists(db *sql.DB) (list []types.CategoryList, err error) {
	rows, err := db.Query("SELECT * FROM category")
	if err != nil {
		return list, err
	}
	for rows.Next() {
		var index types.CategoryList

		err := rows.Scan(&index.CategoryName, &index.Description)
		if err != nil {
			log.Println(err)
		}
		list = append(list, index)
	} 
	return list, nil
}


func AppendCategoryToClient(db *sql.DB, clientMAC string, categoryName string) error {
	// Insert the client-category association into the junction table
	_, err := db.Exec("INSERT INTO client_category (client_mac, category_name) VALUES (?, ?) ON CONFLICT DO NOTHING", clientMAC, categoryName)
	if err != nil {
		return fmt.Errorf("failed to associate client with category: %w", err)
	}
	// log.Printf("Successfully associated client %s with category %s\n", clientMAC, categoryName)

	return nil
}

func ensureCategoryExists(db *sql.DB, categoryName, description string) (string, error) {
	// Check if the category exists
	var existingDescription string
	err := db.QueryRow("SELECT description FROM category WHERE category_name = ?", categoryName).Scan(&existingDescription)

	if err == sql.ErrNoRows {
		// If not exists, insert the new category
		_, err := db.Exec("INSERT INTO category (category_name, description) VALUES (?, ?)", categoryName, description)
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

func insertDomain(txn *sql.Tx, domainName, categoryName string) error {
	// Check if the domain already exists
	var existingDomain string
	err := txn.QueryRow("SELECT domain_name FROM domain WHERE domain_name = ?", domainName).Scan(&existingDomain)

	if err == sql.ErrNoRows {
		// If not exists, insert the new domain
		_, err := txn.Exec("INSERT INTO domain (domain_name) VALUES (?)", domainName)
		if err != nil {
			txn.Rollback()
			return fmt.Errorf("failed to insert domain: %w", err)
		}
	}

	// Now insert the domain-category association in the junction table
	_, err = txn.Exec("INSERT INTO domain_category (domain_name, category_name) VALUES (?, ?) ON CONFLICT DO NOTHING", domainName, categoryName)
	if err != nil {
		return fmt.Errorf("failed to associate domain with category: %w", err)
	}

	return nil
}

func CheckClientDomain(db *sql.DB, clientMAC string, domainName string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM client_category cc
			JOIN domain_category dc ON cc.category_name = dc.category_name
			JOIN client c ON cc.client_mac = c.client_mac
			WHERE c.client_mac = ? AND dc.domain_name = ?
		);
	`
	var exists bool
	log.Println(clientMAC, domainName)
	err := db.QueryRow(query, clientMAC, domainName).Scan(&exists)
	if err != nil {
		log.Println("PSQLLITE ERROR")
		return false, err
	}
	return exists, nil
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

			txn, err := db.Begin()
			if err != nil {
				log.Println(err)
				return
			}

			for _, domain := range d.List {
				err := insertDomain(txn, domain, categoryID)
				if err != nil {
					log.Printf("Error inserting domain %s: %v\n", domain, err)
				} else {
					// log.Printf("Successfully associated domain %s with category %s\n", domain, d.Name)
				}
			}

			if err := txn.Commit(); err != nil {
				log.Println(err)
				return
			}

			log.Println("Successfully imported: ", d.Name)
		}(d) 
	}

	wg.Wait()
}
