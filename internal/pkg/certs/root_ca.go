package certs

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io"
	"log"
	"math/big"
	"os"
	"time"
)

func (c *Cert) GenerateRootCA(rootPath string, keyPath string) (err error) {
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
		
		RootCa, err := x509.ParseCertificate(block.Bytes)
		c.RootCa = *RootCa
		if err != nil {
			return err
		}
		
	} else {
		// Prepare certificate template
		c.RootCa = x509.Certificate{
			SerialNumber: big.NewInt(2024),
			Subject: pkix.Name{
				Organization:  []string{"Eyeo"},
				Country:       []string{"DE"},
				Province:      []string{"Berlin"},
				Locality:      []string{"Berlin"},
				StreetAddress: []string{""},
				PostalCode:    []string{""},
			},
			NotBefore:             time.Now(),
			NotAfter:              time.Now().Add(10 * 365 * 24 * time.Hour), // Valid for 10 years
			KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
			IsCA:                  true,
			BasicConstraintsValid: true,
		}
		log.Printf("Generated new certificate: %s", rootPath)
	}
		
	// Self-sign the certificate (sign with its own private key)
	c.RootDER, err = x509.CreateCertificate(rand.Reader, &c.RootCa, &c.RootCa, &c.rootKey.PublicKey, c.rootKey)
	if err != nil {
		return err
	}

	// Save the certificate to a file (PEM encoded)
	certOut, err := os.Create(rootPath)
	if err != nil {
		return err
	}
	defer certOut.Close()
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: c.RootDER})

	// Save the private key to a file (PEM encoded)
	keyOut, err := os.Create(keyPath)
	if err != nil {
		return err
	}
	defer keyOut.Close()
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(c.rootKey)})

	return nil
}
