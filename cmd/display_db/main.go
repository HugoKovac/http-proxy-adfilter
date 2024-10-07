package main

import (
	"log"

	"gitlab.com/eyeo/network-filtering/router-adfilter-go/internal/db"
	"github.com/boltdb/bolt"
)

func main() {
	_, boltdb := db.NewDatabase()
	defer boltdb.Close()
		
	boltdb.View(func(tx *bolt.Tx) error {
		tx.Bucket([]byte("domain_categories")).ForEach(func(k, v []byte) error {
			log.Printf("key: %s | value: %#v", k, string(v))
			return nil
		})
		return nil
	})
}