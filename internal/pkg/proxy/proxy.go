package proxy

import (
	"crypto/tls"
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/boltdb/bolt"
	"gitlab.com/eyeo/network-filtering/router-adfilter-go/internal/pkg/filter"
)

func handleTunneling(w http.ResponseWriter, r *http.Request, boltdb *bolt.DB) {
	log.Println("HTTPS")
	err := filter.Filter(w, r, boltdb)
	if err != nil {
		log.Println(err)
		return
	}
	dest_conn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	client_conn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}
	go transfer(dest_conn, client_conn)
	go transfer(client_conn, dest_conn)
}

func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()
	io.Copy(destination, source)
}

func buildHTTPRequest(r *http.Request) {
	r.URL.Scheme = "http"
	r.URL.Host = r.Host
}

func handleHTTP(w http.ResponseWriter, r *http.Request, boltdb *bolt.DB) {
	log.Println("HTTP")
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

func serverHandler(tls bool) func(*bolt.DB, http.ResponseWriter, *http.Request) {
	if !tls {
		return func(boltdb *bolt.DB, w http.ResponseWriter, r *http.Request) {
			handleHTTP(w, r, boltdb)
		}
	} else {
		return func(boltdb *bolt.DB, w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodConnect {
				handleTunneling(w, r, boltdb)
			} else {
				handleHTTP(w, r, boltdb)
			}
		}
	}
}

func createHTTPserver(boltdb *bolt.DB, host string, port string, isTLS bool) (httpsServer *http.Server) {
	httpsServer = &http.Server{
		Addr: host + ":" + port,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			buildHTTPRequest(r)
			serverHandler(isTLS)(boltdb, w, r)
		}),
		// Disable HTTP/2.
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	return
}

func ListenProxy(boltdb *bolt.DB) {
	var pemPath string
	flag.StringVar(&pemPath, "pem", "cert.pem", "path to pem file")
	var keyPath string
	flag.StringVar(&keyPath, "key", "key.pem", "path to key file")
	var host string
	flag.StringVar(&host, "host", "0.0.0.0", "interface to listen on")
	var port string
	flag.StringVar(&port, "port", "7080", "HTTP proxy port")
	var tlsPort string
	flag.StringVar(&tlsPort, "sport", "7443", "HTTPS proxy port")
	flag.Parse()
	//todo: check input, move it and put data in struct

	httpServer := createHTTPserver(boltdb, host, port, false)
	httpsServer := createHTTPserver(boltdb, host, tlsPort, true)

	httpChannel := make(chan error)
	httpsChannel := make(chan error)
	go func(err chan error) {
		err <- httpServer.ListenAndServe()
	}(httpChannel)
	go func(err chan error, pemPath string, keyPath string) {
		err <- httpsServer.ListenAndServeTLS(pemPath, keyPath)
	}(httpsChannel, pemPath, keyPath)

	select {
	case httpError := <-httpChannel:
		log.Fatal(httpError)
	case httpsError := <-httpsChannel:
		log.Fatal(httpsError)
	}
}
