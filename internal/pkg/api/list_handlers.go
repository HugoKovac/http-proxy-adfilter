package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

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

func addSubList(w http.ResponseWriter, r *http.Request, db *sql.DB) {
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
	if err := data.EnsureClientExists(db, client); err != nil {
		http.Error(w, "Checking if client exist: " + err.Error(), http.StatusInternalServerError)
		return
	}
	if err := data.AppendCategoryToClient(db, client.MAC.String(), category); err != nil {
		log.Println(err)
		http.Error(w, "Category does not exist", http.StatusBadRequest)
		return
	}

	w.WriteHeader(201)
}

func delSubList(w http.ResponseWriter, r *http.Request, db *sql.DB) {
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
	if err = data.DelSubscribtion(db, category, client.MAC.String()) ;err != nil {
		log.Println(err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

}

func Handler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	log.Println("Method", r.Method)
	log.Println("Path", r.URL.Path)

	switch r.Method {
	case "GET":
		switch r.URL.Path {
		case "/get_category_list":
			getCategoryLists(w, r, db)
		case "/get_sub_lists":
			getSubLists(w, r, db)
		default:
			http.Error(w, "Not Found", http.StatusNotFound)
		}
	case "POST":
		switch r.URL.Path {
		case "/add_sub_list":
			addSubList(w, r, db)
		case "/del_sub_list":
			delSubList(w, r, db)
		default:
			http.Error(w, "Not Found", http.StatusNotFound)
		}
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}
