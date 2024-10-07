package data

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/boltdb/bolt"
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

func insertDomain(b *bolt.Bucket, domainName string, categoryName string) error {
	// Check is domain already have catgories
	value := b.Get([]byte(domainName))
	// if not create empty json array with domain struct
	if value == nil {
		buf, err := json.Marshal([]string{categoryName})
		if err != nil {
			return fmt.Errorf("format in json: %s", categoryName)
		}
		// associate domain name with the name of the category
		log.Println(domainName, buf)
		return b.Put([]byte(domainName), buf)
	} else {
		var categoriesArray []string
		err := json.Unmarshal(value, &categoriesArray)
		if err != nil {
			return fmt.Errorf("unmarshal %s's value", domainName)
		}
		for _, cat := range categoriesArray {
			if strings.Compare(cat, categoryName) == 0{
				return nil
			}
		}
		categoriesArray = append(categoriesArray, categoryName)
		buf, err := json.Marshal(categoriesArray)
		if err != nil {
			return fmt.Errorf("format in json: %s", categoryName)
		}
		// associate domain name with the name of the category
		return b.Put([]byte(domainName), buf)
	}
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

func GetCategorizedDomainList(db *sql.DB, boltdb *bolt.DB) {
	domainLists, err := fakeFetch()
	if err != nil {
		log.Println(err)
		return
	}
	
	for _, category := range domainLists {
		for _, j := range category.List{
			log.Println(j)
		}
	}

	var wg sync.WaitGroup

	for _, category := range domainLists { // iterate in domainLists.list
		wg.Add(1)

		// Create new thread for each list
		go func(category types.DomainList) {
			defer wg.Done()
			boltdb.Update(func(tx *bolt.Tx) error {
				// Get related bucker
				b := tx.Bucket([]byte("domain_categories"))

				// For all domains
				for _, domain := range category.List {
					insertDomain(b, domain, category.Name)
				}
				return nil
			})

			log.Println("Successfully imported: ", category.Name)
		}(category) 
	}

	wg.Wait()
}
