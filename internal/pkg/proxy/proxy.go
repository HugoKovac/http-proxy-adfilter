package proxy

import (
	"crypto/tls"
	"log"
	"net/http"

	"github.com/boltdb/bolt"
	"gitlab.com/eyeo/network-filtering/router-adfilter-go/internal/pkg/filter"
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

	// if CONNECT https
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

	certFile, keyFile, err := generateSelfSignedCert()

	if err != nil {
		log.Fatal(err)
	}

	tlsConfig := &tls.Config{
		MinVersion:               tls.VersionTLS12,
		PreferServerCipherSuites: true,
		CurvePreferences:         []tls.CurveID{tls.CurveP256, tls.X25519},
	}


	server := &http.Server{
		Addr:      HOST + ":" + PORT,
		Handler:   h,
		TLSConfig: tlsConfig,
	}

	log.Fatal(server.ListenAndServeTLS(certFile, keyFile))
}
