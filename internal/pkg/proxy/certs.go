package proxy

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io"
	"log"
	"math/big"
	"os"
	"time"
)


func (c *cert) generateRootCA(rootPath string, keyPath string) (err error) {
	// Check if certificate and key files already exist
	if _, err := os.Stat(keyPath); err == nil {
		keyFile, err := os.Open(keyPath)
		if err != nil {
			return nil
		}
		keyContent, err := io.ReadAll(keyFile)
		if err != nil {
			return nil
		}
		block, _ := pem.Decode(keyContent)
		c.rootKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil
		}

	} else {
		// Generate private key if not found
		c.rootKey, err = rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return err
		}
		log.Printf("Generated new key: %s", keyPath)
	}
	if _, err := os.Stat(rootPath); err == nil {
		rootFile, err := os.Open(rootPath)
		if err != nil {
			return nil
		}
		rootContent, err := io.ReadAll(rootFile)
		if err != nil {
			return nil
		}
		block, _ := pem.Decode(rootContent)
		
		rootCa, err := x509.ParseCertificate(block.Bytes)
		c.rootCa = *rootCa
		if err != nil {
			return err
		}
		
	} else {
		// Prepare certificate template
		c.rootCa = x509.Certificate{
			SerialNumber: big.NewInt(1),
			Subject: pkix.Name{
				CommonName: "proxy",
				Organization: []string{"Eyeo"},
			},
			NotBefore:             time.Now(),
			NotAfter:              time.Now().Add(365 * 24 * time.Hour), // Valid for 1 year
			KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
			ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			BasicConstraintsValid: true,
		}
		c.rootCa.DNSNames = []string{"proxy"} // Include SANs here
		log.Printf("Generated new certificate: %s", rootPath)
	}
		
	// Self-sign the certificate (sign with its own private key)
	c.rootDER, err = x509.CreateCertificate(rand.Reader, &c.rootCa, &c.rootCa, &c.rootKey.PublicKey, c.rootKey)
	if err != nil {
		return err
	}

	// Save the certificate to a file (PEM encoded)
	certOut, err := os.Create(rootPath)
	if err != nil {
		return err
	}
	defer certOut.Close()
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: c.rootDER})

	// Save the private key to a file (PEM encoded)
	keyOut, err := os.Create(keyPath)
	if err != nil {
		return err
	}
	defer keyOut.Close()
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(c.rootKey)})

	return nil
}

func (c *cert) intermediateCa(domain string) (*tls.Certificate, error) {
	defer func ()  {
		if r := recover(); r != nil {
			log.Printf("Recovered in intermediateCa for domain %s: %v", domain, r)
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

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	// Sign the certificate with the CA
	certDER, err := x509.CreateCertificate(rand.Reader, &cert, &c.rootCa, &privateKey.PublicKey, c.rootKey)
	if err != nil {
		return nil, err
	}


	// PEM encode the certificate and private key
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})

	// Create a tls.Certificate to use in tls.Config
	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}

	// Cache the certificate
	c.mu.Lock()
	c.certCache[domain] = &tlsCert
	c.mu.Unlock()

	log.Printf("Generated certificate for domain: %s", domain)
	return &tlsCert, nil

}

func getCertificateFunc(cert *cert) func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	return func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		domain := hello.ServerName

		cert.mu.Lock()
		domainCert, exists := cert.certCache[domain]
		cert.mu.Unlock()

		if exists {
			return domainCert, nil
		}

		// Generate a new certificate for the domain
		return cert.intermediateCa(domain)
	}
}

