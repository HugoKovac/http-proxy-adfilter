package proxy

import (
	"io"
	"log"
	"net/http"

	"github.com/boltdb/bolt"
	"gitlab.com/eyeo/network-filtering/router-adfilter-go/internal/pkg/filter"
)

func buildHTTPRequest(r *http.Request) {
	r.URL.Scheme = "http"
	r.URL.Host = r.Host
}

func handleHTTP(w http.ResponseWriter, r *http.Request, boltdb *bolt.DB) {
	err := filter.Filter(w, r, boltdb)
	if err != nil {
		log.Println(err)
		return
	}
	resp, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		log.Printf("error with HTTP Roundtrip\nRequest: %#v\n", r)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()
	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}