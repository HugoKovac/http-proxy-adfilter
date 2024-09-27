package proxy

import (
	"log"
	"net/http"
	"gitlab.com/eyeo/network-filtering/router-adfilter-go/internal/pkg/filter"
	"gitlab.com/eyeo/network-filtering/router-adfilter-go/internal/pkg/api"
)

type handler struct {
}

func (h handler) ServeHTTP(originalWriter http.ResponseWriter, originalRequest *http.Request) {
	HeaderHandler := NewHeaderHandler()
	requestHandler := NewRequestHandler()

	if originalRequest.Host == "localhost:8080"{
		api.Handler(originalWriter, originalRequest)
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

func ListenProxy() {
	h := handler{}

	log.Fatal(http.ListenAndServe("localhost:8080", h))
}
