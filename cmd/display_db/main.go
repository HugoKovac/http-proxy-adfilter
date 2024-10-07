package main

import (
	"log"

	"gitlab.com/eyeo/network-filtering/router-adfilter-go/internal/db"
	"github.com/boltdb/bolt"
)

func main() {
	boltdb := db.NewDatabase()
	defer boltdb.Close()
		
	boltdb.View(func(tx *bolt.Tx) error {
		log.Println("=====domain_categories=====")
		tx.Bucket([]byte("domain_categories")).ForEach(func(k, v []byte) error {
			log.Printf("key: %s | value: %#v", k, string(v))
			return nil
		})
		log.Println("=====mac_client=====")
		tx.Bucket([]byte("mac_client")).ForEach(func(k, v []byte) error {
			log.Printf("key: %s | value: %#v", k, string(v))
			return nil
		})
		log.Println("=====client_categories=====")
		tx.Bucket([]byte("client_categories")).ForEach(func(k, v []byte) error {
			log.Printf("key: %s | value: %#v", k, string(v))
			return nil
		})
		return nil
	})
}