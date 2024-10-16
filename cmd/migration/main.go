package main

import (
	"log"

	"gitlab.com/eyeo/network-filtering/router-adfilter-go/internal/db"
	"github.com/boltdb/bolt"
)

func main() {
	boltdb := db.NewDatabase()
	defer boltdb.Close()
	
	err := boltdb.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("mac_client"))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte("client_categories"))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte("domain_categories"))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("Buckets created")
	}
}