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

func CreateMacClient(boltdb *bolt.DB, client macClients.Client) error {
	// Convert MAC and IP to strings
	macStr := client.MAC.String()
	ipStr := client.IP.String()

	// Check if already exist
	err := boltdb.Update(func(tx *bolt.Tx) (err error) {
		bucket := tx.Bucket([]byte("mac_client"))
		err = bucket.Put([]byte(macStr), []byte(ipStr))
		return err
	})

	return err
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

func AppendValue(b *bolt.Bucket, key string, value string) error {
	// Check is domain already have catgories
	pastValue := b.Get([]byte(key))
	// if not create empty json array with domain struct
	if pastValue == nil {
		buf, err := json.Marshal([]string{value})
		if err != nil {
			return fmt.Errorf("format in json: %s", value)
		}
		// associate domain name with the name of the category
		log.Println(key, buf)
		return b.Put([]byte(key), buf)
	} else {
		var categoriesArray []string
		err := json.Unmarshal(pastValue, &categoriesArray)
		if err != nil {
			return fmt.Errorf("unmarshal %s's value", key)
		}
		for _, cat := range categoriesArray {
			if strings.Compare(cat, value) == 0{
				return nil
			}
		}
		categoriesArray = append(categoriesArray, value)
		buf, err := json.Marshal(categoriesArray)
		if err != nil {
			return fmt.Errorf("format in json: %s", value)
		}
		// associate domain name with the name of the category
		return b.Put([]byte(key), buf)
	}
}

func DelValue(b *bolt.Bucket, key string, value string) error {
	// Check is domain already have catgories
	pastValue := b.Get([]byte(key))
	// if not create empty json array with domain struct
	if pastValue == nil {
		return nil
	} else {
		var categoriesArray []string
		err := json.Unmarshal(pastValue, &categoriesArray)
		if err != nil {
			return fmt.Errorf("unmarshal %s's value", key)
		}
		i := -1
		for key, cat := range categoriesArray {
			if strings.Compare(cat, value) == 0{
				i = key
				break
			}
		}
		if i == -1 {
			return fmt.Errorf("don't exist")
		}
		categoriesArray = append(categoriesArray[:i], categoriesArray[i+1:]...)
		buf, err := json.Marshal(categoriesArray)
		if err != nil {
			return fmt.Errorf("format in json: %s", value)
		}
		// associate domain name with the name of the category
		return b.Put([]byte(key), buf)
	}
}

func hasCommonElement(arr1, arr2 []string) bool {
	// Create a map to store the elements of the first array
	elementMap := make(map[string]bool)

	// Populate the map with elements from the first array
	for _, str := range arr1 {
		elementMap[str] = true
	}

	// Check if elements of the second array exist in the map
	for _, str := range arr2 {
		if elementMap[str] {
			return true
		}
	}

	return false
}

func CheckClientDomain(db *sql.DB, boltdb *bolt.DB, clientMAC string, domainName string) (ok bool, err error) {
	var clientCategories []string
	var domainCategories []string
	
	err = boltdb.View(func(tx *bolt.Tx) error {
		clientBucket := tx.Bucket([]byte("client_categories"))
		domainBucket := tx.Bucket([]byte("domain_categories"))

		rawClientCategories := clientBucket.Get([]byte(clientMAC))
		rawDomainCategories := domainBucket.Get([]byte(domainName))

		if rawClientCategories == nil || rawDomainCategories == nil {
			return nil // No categories means not blocked
		}

		if err := json.Unmarshal(rawClientCategories, &clientCategories); err != nil {
			return err
		}
		if err := json.Unmarshal(rawDomainCategories, &domainCategories); err != nil {
			return err
		}

		ok = hasCommonElement(clientCategories, domainCategories)

		return nil

	})
	return ok, err
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
			boltdb.Batch(func(tx *bolt.Tx) error {
				// Get related bucker
				b := tx.Bucket([]byte("domain_categories"))

				// For all domains
				for _, domain := range category.List {
					AppendValue(b, domain, category.Name)
				}
				return nil
			})

			log.Println("Successfully imported: ", category.Name)
		}(category) 
	}

	wg.Wait()
}
