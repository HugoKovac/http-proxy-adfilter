package api

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/boltdb/bolt"
)

const (
	HOST = "localhost"
	PORT = "8080"
)

func ListenHandler(db *sql.DB, boltdb *bolt.DB) {
	http.HandleFunc("/get_category_list", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			getCategoryLists(w, r, db)
			return
		}
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	})
	http.HandleFunc("/get_sub_lists", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			getSubLists(w, r, db)
			return
		}
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	})
	http.HandleFunc("/add_sub_list", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			addSubList(w, r, db)
			return
		}
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	})
	http.HandleFunc("/del_sub_list", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			delSubList(w, r, db)
			return
		}
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	})

	log.Fatal(http.ListenAndServe(HOST + ":" + PORT, nil))
}
