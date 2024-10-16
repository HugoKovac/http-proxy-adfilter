package api

import (
	"log"
	"net/http"

	"github.com/boltdb/bolt"
)

const (
	HOST = "0.0.0.0"
	PORT = "9000"
)

func ListenHandler(boltdb *bolt.DB) {
	http.HandleFunc("/get_sub_lists", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			getSubLists(w, r, boltdb)
			return
		}
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	})
	http.HandleFunc("/add_sub_list", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			addSubList(w, r, boltdb)
			return
		}
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	})
	http.HandleFunc("/del_sub_list", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			delSubList(w, r, boltdb)
			return
		}
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	})

	log.Fatal(http.ListenAndServe(HOST + ":" + PORT, nil))
}
