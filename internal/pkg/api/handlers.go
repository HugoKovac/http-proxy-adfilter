package api

import (
	"net/http"
	"log"
)

func getLists(w http.ResponseWriter, r *http.Request) {
/* 

*/
}

func getSubLists(w http.ResponseWriter, r *http.Request) {
/* 

*/
}

func addSubList(w http.ResponseWriter, r *http.Request) {
/* 

*/
}

func delSubList(w http.ResponseWriter, r *http.Request) {
/* 
	
*/
}

func Handler(w http.ResponseWriter, r *http.Request) {
	log.Println("Method", r.Method)
	log.Println("Path", r.URL.Path)

	switch r.Method {
	case "GET":
		switch r.URL.Path{
		case "/get_lists":
			getLists(w,r)
		case "/get_sub_lists":
			getSubLists(w,r)
		default:
			http.Error(w, "Not Found", http.StatusNotFound)
		}
	case "POST":
		switch r.URL.Path{
		case "/add_sub_list":
			addSubList(w,r)
		case "/del_sub_list":
			delSubList(w,r)
		default:
			http.Error(w, "Not Found", http.StatusNotFound)
		}
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}