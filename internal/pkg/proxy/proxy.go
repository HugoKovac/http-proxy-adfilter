package proxy

import (
	"database/sql"
	"log"
	"net/http"

	"gitlab.com/eyeo/network-filtering/router-adfilter-go/internal/pkg/api"
	"gitlab.com/eyeo/network-filtering/router-adfilter-go/internal/pkg/filter"
)

type handler struct {
	db *sql.DB
}

func (h handler) ServeHTTP(originalWriter http.ResponseWriter, originalRequest *http.Request) {
	HeaderHandler := NewHeaderHandler()
	requestHandler := NewRequestHandler()

	if originalRequest.Host == "localhost:8080"{
		api.Handler(originalWriter, originalRequest, h.db)
		return
	}

	err := filter.Filter(originalRequest)
	if err != nil {
		log.Println(err)
		return
	}
	HeaderHandler.PreRequest(originalRequest)

	proxyResp, err := requestHandler.Do(originalWriter, originalRequest)
	if err != nil {
		log.Println("Requesting Error: ", err)
		return
	}

	log.Println("Response Status", proxyResp.Status)

	HeaderHandler.PostRequest(proxyResp, originalWriter)
}

func ListenProxy(db *sql.DB) {
	h := handler{db}

	log.Fatal(http.ListenAndServe("localhost:8080", h))
}
