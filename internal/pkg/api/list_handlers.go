package api

import (
	"database/sql"
	"log"
	"net/http"
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
	One client can have many lists
	add a list to client in db
*/
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
		switch r.URL.Path{
		case "/get_lists":
			getLists(w,r,db)
		case "/get_sub_lists":
			getSubLists(w,r,db)
		default:
			http.Error(w, "Not Found", http.StatusNotFound)
		}
	case "POST":
		switch r.URL.Path{
		case "/add_sub_list":
			addSubList(w,r,db)
		case "/del_sub_list":
			delSubList(w,r,db)
		default:
			http.Error(w, "Not Found", http.StatusNotFound)
		}
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}