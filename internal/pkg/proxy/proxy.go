package proxy

import (
	"crypto/tls"
	"flag"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/boltdb/bolt"
)


type flags struct {
	pemPath string
	keyPath string
	host    string
	port    string
	tlsPort string
}

type proxy struct {
	flags      *flags
	Boltdb     *bolt.DB
	httpServer *http.Server
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
	stopChan := make(chan bool)

	go func(stopChan chan bool, errorChan chan error) {
		listener, err := net.Listen("tcp", p.flags.host+":"+p.flags.tlsPort)
		if err != nil {
			errorChan <- err
			return
		}
		defer listener.Close()

		for {
			select {
			case <-stopChan:
				return
			default:
				clientConn, err := listener.Accept()
				if err != nil {
					log.Printf("Error accepting connection: %v", err)
					continue
				}

				go handleConnection(clientConn, p.Boltdb)
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
