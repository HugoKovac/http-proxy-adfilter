package certs

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"time"
)

func (c *Cert) intermediateCa(domain string) (*tls.Certificate, error) {
	defer func ()  {
		if r := recover(); r != nil {
			log.Printf("Recovered in intermediateCa for domain %s: %v", domain, r)
		} else {
			log.Printf("Generated certificate for domain: %s", domain)
		}
	}()
	cert := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			CommonName:   domain,
			Organization: []string{"Eyeo"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // 1 year validity
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{domain},
	}

	// privateKey, err := rsa.GenerateKey(rand.Reader, 2048) // performance issue
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}


	// Sign the certificate with the CA
	certDER, err := x509.CreateCertificate(rand.Reader, &cert, &c.RootCa, &privateKey.PublicKey, c.rootKey)
	if err != nil {
		return nil, err
	}


	// PEM encode the certificate and private key
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	der, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, err
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: der})


	// Create a tls.Certificate to use in tls.Config
	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}

	// Cache the certificate
	c.Mu.Lock()
	c.CertCache[domain] = &tlsCert
	c.Mu.Unlock()

	return &tlsCert, nil
}

func GetCertificateFunc(cert *Cert) func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	return func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		domain := hello.ServerName

		cert.Mu.Lock()
		domainCert, exists := cert.CertCache[domain]
		cert.Mu.Unlock()

		if exists {
			return domainCert, nil
		}

		// Generate a new certificate for the domain
		return cert.intermediateCa(domain)
	}
}
