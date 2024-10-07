package proxy

import (
	"log"
	"net/http"

	// "gitlab.com/eyeo/network-filtering/router-adfilter-go/internal/pkg/api"
	"gitlab.com/eyeo/network-filtering/router-adfilter-go/internal/pkg/filter"
	"github.com/boltdb/bolt"
)

const (
	PORT = "8888"
	HOST = "0.0.0.0"
)

type handler struct {
	boltdb *bolt.DB
}

func (h handler) ServeHTTP(originalWriter http.ResponseWriter, originalRequest *http.Request) {
	HeaderHandler := NewHeaderHandler()
	requestHandler := NewRequestHandler()

	originalRequest.URL.Scheme = "http"
	originalRequest.URL.Host = originalRequest.Host
	originalRequest.URL.Path = originalRequest.RequestURI
	//TODO: Fill and check all URL vaiable like params

	err := filter.Filter(originalWriter, originalRequest, h.boltdb)
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

func ListenProxy(boltdb *bolt.DB) {
	h := handler{boltdb}

	log.Fatal(http.ListenAndServe(HOST + ":" + PORT, h))
}
