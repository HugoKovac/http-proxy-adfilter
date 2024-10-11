package certs

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"sync"
)

type Cert struct {
	RootDER   []byte
	RootCa    x509.Certificate
	rootKey   *rsa.PrivateKey
	CertCache map[string]*tls.Certificate
	Mu        sync.Mutex
}
