package proxy

import (
	"database/sql"
	"log"
	"net/http"

	"gitlab.com/eyeo/network-filtering/router-adfilter-go/internal/pkg/api"
	"gitlab.com/eyeo/network-filtering/router-adfilter-go/internal/pkg/filter"
)

const (
	PORT = "8888"
	HOST = "0.0.0.0"
)

type handler struct {
	db *sql.DB
}

func (h handler) ServeHTTP(originalWriter http.ResponseWriter, originalRequest *http.Request) {
	HeaderHandler := NewHeaderHandler()
	requestHandler := NewRequestHandler()

	originalRequest.URL.Scheme = "http"
	originalRequest.URL.Host = originalRequest.Host
	originalRequest.URL.Path = originalRequest.RequestURI
	//TODO: Fill and check all URL vaiable like params

	if originalRequest.Host == "192.168.10.1" {
		log.Println("Handler")
		api.Handler(originalWriter, originalRequest, h.db)
		return
	}

	err := filter.Filter(originalWriter, originalRequest, h.db)
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

	log.Fatal(http.ListenAndServe(HOST + ":" + PORT, h))
}
