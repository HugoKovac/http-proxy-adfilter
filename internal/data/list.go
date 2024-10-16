package data

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	macClients "gitlab.com/eyeo/network-filtering/router-adfilter-go/internal/pkg/mac_clients"
)

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
		return b.Put([]byte(key), buf)
	} else {
		var categoriesArray []string
		err := json.Unmarshal(pastValue, &categoriesArray)
		if err != nil {
			return fmt.Errorf("unmarshal %s's value", key)
		}
		for _, cat := range categoriesArray {
			if strings.Compare(cat, value) == 0 {
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

func GetClientCategoriesList(boltdb *bolt.DB, macAddr string) (list []string, err error) {
	err = boltdb.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("client_categories"))
		value := b.Get([]byte(macAddr))
		json.Unmarshal(value, &list)

		return nil
	})
	return list, err
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
			if strings.Compare(cat, value) == 0 {
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

func CheckClientDomain(boltdb *bolt.DB, clientMAC string, domainName string) (ok bool, err error) {
	var clientCategories []string
	var domainCategories []string

	err = boltdb.View(func(tx *bolt.Tx) error {
		clientBucket := tx.Bucket([]byte("client_categories"))
		domainBucket := tx.Bucket([]byte("domain_categories"))

		domainName = strings.Split(domainName, ":")[0]

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

func fetch(uri string) (string, error) {
	defer func() {
		if recover := recover(); recover != nil {
			log.Println(recover)
		}
	}()
	client := http.Client{}
	req, err := http.NewRequest("GET", "https://easylist-downloads.adblockplus.org/network/nlf/v1/"+uri, nil)
	req.Header.Add("Accept-Encoding", "text/plain")
	if err != nil {
		return "", fmt.Errorf("accept text header: %v", err)
	}
	req.SetBasicAuth("nasa_user", os.Getenv("LIST_PASS"))
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("do: %v", err)
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("%s got status: %s", uri, resp.Status)
	}
	list, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read all: %v", err)
	}

	return string(list), nil
}

func GetCategorizedDomainList(boltdb *bolt.DB, categoryNames []string) {
	// Threading can overload and a crash of the program
	// var wg sync.WaitGroup

	const batchSize = 300

	for _, category := range categoryNames {
		// wg.Add(1)

		// go func(category string) {
			// defer wg.Done()
			now := time.Now()
			totalCount := 0

			baseList, err := fetch(category + ".txt")
			if err != nil {
				log.Println("Error fetching: ", err)
				return
			}

			baseLines := strings.Split(baseList, "\n")

			for i := 0; i < len(baseLines); i += batchSize {
				// Process lines in chunks of batchSize (max 300 lines)
				end := i + batchSize
				if end > len(baseLines) {
					end = len(baseLines)
				}
				batchLines := baseLines[i:end]

				// Process the current batch
				count := 0
				err := boltdb.Batch(func(tx *bolt.Tx) error {
					b := tx.Bucket([]byte("domain_categories"))
					for _, domain := range batchLines {
						if _, err := url.Parse(domain); err == nil {
							count += 1
							AppendValue(b, domain, category)
						}
					}
					return nil
				})

				if err != nil {
					log.Printf("Error processing batch for %s: %v\n", category, err)
				}

				totalCount += count
			}

			log.Printf("%s: %d domains imported in %v\n", category, totalCount, time.Since(now))
		// }(category)
	}
	// wg.Wait()
}
