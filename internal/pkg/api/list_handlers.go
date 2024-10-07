package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/boltdb/bolt"
	"gitlab.com/eyeo/network-filtering/router-adfilter-go/internal/data"
	macClients "gitlab.com/eyeo/network-filtering/router-adfilter-go/internal/pkg/mac_clients"
)

func getCategoryLists(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	list, err := data.GetCategoryLists(db);
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
	}

	if err = json.NewEncoder(w).Encode(list); err != nil {
		http.Error(w, "", http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
}

func getSubLists(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	client, err := macClients.GetInfoFromIP(r.RemoteAddr)
	if err != nil {
		http.Error(w, "Getting client info: " + err.Error(), http.StatusInternalServerError)
		return
	}
	list, err := data.GetSubscribedCategoryLists(db, client.MAC.String());
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
	}

	if err = json.NewEncoder(w).Encode(list); err != nil {
		http.Error(w, "", http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
}

func addSubList(w http.ResponseWriter, r *http.Request, boltdb *bolt.DB) {
	r.ParseForm()
	category := r.FormValue("category")
	if category == "" {
		http.Error(w, "Expecting catgory parameter", http.StatusBadRequest)
		return
	}

	client, err := macClients.GetInfoFromIP(r.RemoteAddr)
	if err != nil {
		http.Error(w, "Getting client info: " + err.Error(), http.StatusInternalServerError)
		return
	}
	if err := data.CreateMacClient(boltdb, client); err != nil {
		http.Error(w, "Checking if client exist: " + err.Error(), http.StatusInternalServerError)
		return
	}
	if err = boltdb.Update(func(tx *bolt.Tx) (err error) {
		// Get related bucker
		b := tx.Bucket([]byte("client_categories"))

		err = data.AppendValue(b, client.MAC.String(), category)
		return err
	}); err != nil {
		log.Println(err)
		http.Error(w, "subscribe to the list", http.StatusBadRequest)
		return
	}

	w.WriteHeader(201)
}

func delSubList(w http.ResponseWriter, r *http.Request, boltdb *bolt.DB) {
	r.ParseForm()
	category := r.FormValue("category")
	if category == "" {
		http.Error(w, "Expecting catgory parameter", http.StatusBadRequest)
		return
	}

	client, err := macClients.GetInfoFromIP(r.RemoteAddr)
	if err != nil {
		http.Error(w, "Getting client info: " + err.Error(), http.StatusInternalServerError)
		return
	}
	if err = boltdb.Update(func(tx *bolt.Tx) (err error) {
		// Get related bucker
		b := tx.Bucket([]byte("client_categories"))

		log.Println(category, client.MAC.String())
		err = data.DelValue(b, client.MAC.String(), category)
		return err
	}); err != nil {
		log.Println(err)
		http.Error(w, "subscribe to the list", http.StatusBadRequest)
		return
	}

}

