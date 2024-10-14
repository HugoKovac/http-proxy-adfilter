package proxy

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/boltdb/bolt"
	"gitlab.com/eyeo/network-filtering/router-adfilter-go/internal/data"
	macClients "gitlab.com/eyeo/network-filtering/router-adfilter-go/internal/pkg/mac_clients"
)

// ref: https://www.agwa.name/blog/post/writing_an_sni_proxy_in_go

// readOnlyConn implements the net.Conn interface using a reader
type readOnlyConn struct {
	reader io.Reader
}

func (conn readOnlyConn) Read(p []byte) (int, error)         { return conn.reader.Read(p) }
func (conn readOnlyConn) Write(p []byte) (int, error)        { return 0, io.ErrClosedPipe }
func (conn readOnlyConn) Close() error                       { return nil }
func (conn readOnlyConn) LocalAddr() net.Addr                { return nil }
func (conn readOnlyConn) RemoteAddr() net.Addr               { return nil }
func (conn readOnlyConn) SetDeadline(t time.Time) error      { return nil }
func (conn readOnlyConn) SetReadDeadline(t time.Time) error  { return nil }
func (conn readOnlyConn) SetWriteDeadline(t time.Time) error { return nil }

// peekClientHello peeks the ClientHello from the connection
func peekClientHello(reader io.Reader) (*tls.ClientHelloInfo, io.Reader, error) {
	peekedBytes := new(bytes.Buffer)
	hello, err := readClientHello(io.TeeReader(reader, peekedBytes))
	if err != nil {
		return nil, nil, err
	}
	return hello, io.MultiReader(peekedBytes, reader), nil
}

// readClientHello extracts ClientHelloInfo by initiating a TLS handshake
func readClientHello(reader io.Reader) (*tls.ClientHelloInfo, error) {
	var hello *tls.ClientHelloInfo

	err := tls.Server(readOnlyConn{reader: reader}, &tls.Config{
		GetConfigForClient: func(argHello *tls.ClientHelloInfo) (*tls.Config, error) {
			hello = new(tls.ClientHelloInfo)
			*hello = *argHello
			return nil, nil
		},
	}).Handshake()

	if hello == nil {
		return nil, err
	}
	return hello, nil
}

func resolveAddress(serverName string) (string, error) {
	ips, err := net.LookupIP(serverName)
	if err != nil {
		return "", fmt.Errorf("failed to resolve address for %s: %v", serverName, err)
	}
	// Return the first IP address found
	return ips[0].String(), nil
}

func handleConnection(clientConn net.Conn, boltdb *bolt.DB) {
	defer clientConn.Close()

	// Set a read deadline
	if err := clientConn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		log.Print(err)
		return
	}

	// Peek at the ClientHello to get the SNI and preserve the original stream
	clientHello, clientReader, err := peekClientHello(clientConn)
	if err != nil {
		log.Print(err)
		return
	}

	// Clear the read deadline
	if err := clientConn.SetReadDeadline(time.Time{}); err != nil {
		log.Print(err)
		return
	}

	// Verify the SNI is within the allowed domain
	client, err := macClients.GetInfoFromIP(clientConn.RemoteAddr().String())
	if err != nil {
		log.Println(err)
		return
	}
	blocked, err := data.CheckClientDomain(boltdb, client.MAC.String(), clientHello.ServerName)
	if err != nil {
		log.Println(err)
		return
	}
	if blocked == true {
		log.Println("Blocked: ", clientHello.ServerName)
		return
	}

	// Dial the backend server based on the SNI
	backendConn, err := net.DialTimeout("tcp", net.JoinHostPort(clientHello.ServerName, "443"), 5*time.Second)
	if err != nil {
		log.Print(err)
		return
	}
	defer backendConn.Close()

	// Start proxying data between client and backend
	var wg sync.WaitGroup
	wg.Add(2)

	// Proxy client to backend
	go func() {
		io.Copy(clientConn, backendConn)
		clientConn.(*net.TCPConn).CloseWrite()
		wg.Done()
	}()

	// Proxy backend to client
	go func() {
		io.Copy(backendConn, clientReader)
		backendConn.(*net.TCPConn).CloseWrite()
		wg.Done()
	}()

	wg.Wait()
}