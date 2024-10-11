package proxy

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"sync"
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

type flags struct {
	pemPath string
	keyPath string
	host    string
	port    string
	tlsPort string
}

type cert struct {
	rootDER   []byte
	rootCa    x509.Certificate
	rootKey   *rsa.PrivateKey
	certCache map[string]*tls.Certificate
	mu        sync.Mutex
}

type proxy struct {
	flags      *flags
	Boltdb     *bolt.DB
	httpServer *http.Server
	cert       cert
}

func parseFlags() *flags {
	var flags flags
	//todo: check input, move it and put data in struct
	flag.StringVar(&flags.pemPath, "pem", "ssl/EyeoCA.pem", "path to pem file")
	flag.StringVar(&flags.keyPath, "key", "ssl/EyeoKey.pem", "path to key file")
	flag.StringVar(&flags.host, "host", "0.0.0.0", "interface to listen on")
	flag.StringVar(&flags.port, "port", "7080", "HTTP proxy port")
	flag.StringVar(&flags.tlsPort, "sport", "7443", "HTTPS proxy port")
	flag.Parse()

	return &flags
}

func (p *proxy) runHTTP(errorChan chan error) chan bool {
	p.httpServer = &http.Server{
		Addr: p.flags.host + ":" + p.flags.port,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			buildHTTPRequest(r)
			handleHTTP(w, r, p.Boltdb)
		}),
		// Disable HTTP/2.
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	stopChan := make(chan bool)
	go func(stopChan chan bool, errorChan chan error) {
		err := p.httpServer.ListenAndServe()
		if err != nil {
			errorChan <- err
			return
		}
		for {
			select {
			case <-stopChan:
				return
			default:
				time.Sleep(time.Second)
			}
		}
	}(stopChan, errorChan)

	return stopChan
}

func (p *proxy) runHTTPS(errorChan chan error) chan bool {
	err := p.cert.generateRootCA(p.flags.pemPath, p.flags.keyPath)
	if err != nil {
		log.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM([]byte(p.cert.rootDER))

	p.cert.certCache = make(map[string]*tls.Certificate)
	p.httpServer = &http.Server{
		Addr: p.flags.host + ":" + p.flags.tlsPort,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			buildHTTPRequest(r)
			if r.Method == http.MethodConnect {
				handleTunneling(w, r, p.Boltdb)
			} else {
				handleHTTP(w, r, p.Boltdb)
			}
		}),
		TLSConfig: &tls.Config{
			GetCertificate: getCertificateFunc(&p.cert),
			RootCAs:        caCertPool,
		},
		// Disable HTTP/2.
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	stopChan := make(chan bool)
	go func(stopChan chan bool, errorChan chan error) {
		err := p.httpServer.ListenAndServeTLS(p.flags.pemPath, p.flags.keyPath)
		if err != nil {
			errorChan <- err
			return
		}
		for {
			select {
			case <-stopChan:
				return
			default:
				time.Sleep(time.Second)
			}
		}
	}(stopChan, errorChan)

	return stopChan
}

func ListenProxy(boltdb *bolt.DB) (stopHTTP chan bool, stopHTTPS chan bool, errorChan chan error) {
	flags := parseFlags()

	httpServer := proxy{
		flags:  flags,
		Boltdb: boltdb,
	}
	httpsServer := proxy{
		flags: flags,
		Boltdb: boltdb,
	}

	errorChan = make(chan error)
	stopHTTP = httpServer.runHTTP(errorChan)
	stopHTTPS = httpsServer.runHTTPS(errorChan)

	return stopHTTP, stopHTTPS, errorChan
}
