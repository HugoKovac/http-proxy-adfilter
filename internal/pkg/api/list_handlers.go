package api

import (
	"database/sql"
	"log"
	"net/http"

	"gitlab.com/eyeo/network-filtering/router-adfilter-go/internal/data"
	macClients "gitlab.com/eyeo/network-filtering/router-adfilter-go/internal/pkg/mac_clients"
)

func getLists(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	/*
		? How are the list stored (name, id, description, domains) | Pull a json and store it in db
		Pull data from where it's stored and return it
	*/
}

func getSubLists(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	/*
		gest lists
		return the name or id
	*/
}

func addSubList(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	/*
		check if client exist, if not create and then add category to client
	*/
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
	/*
		del a list to client in db
	*/
}

func Handler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	log.Println("Method", r.Method)
	log.Println("Path", r.URL.Path)

	switch r.Method {
	case "GET":
		switch r.URL.Path {
		case "/get_lists":
			getLists(w, r, db)
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
