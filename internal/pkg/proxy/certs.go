package proxy


import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"log"
	"os"
	"time"
)

func generateSelfSignedCert() (certFile, keyFile string, err error) {
	certFile = "cert.pem"
	keyFile = "key.pem"

	// Check if certificate and key files already exist
	if _, err := os.Stat(certFile); err == nil {
		log.Printf("Certificate already exists: %s", certFile)
		return certFile, keyFile, nil
	}
	if _, err := os.Stat(keyFile); err == nil {
		log.Printf("Private key already exists: %s", keyFile)
		return certFile, keyFile, nil
	}

	// Generate private key if not found
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", err
	}

	// Prepare certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "localhost",
			Organization: []string{"My Organization"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // Valid for 1 year
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	template.DNSNames = []string{"localhost"} // Include SANs here

	// Self-sign the certificate (sign with its own private key)
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return "", "", err
	}

	// Save the certificate to a file (PEM encoded)
	certOut, err := os.Create(certFile)
	if err != nil {
		return "", "", err
	}
	defer certOut.Close()
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	// Save the private key to a file (PEM encoded)
	keyOut, err := os.Create(keyFile)
	if err != nil {
		return "", "", err
	}
	defer keyOut.Close()
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})

	log.Printf("Generated new certificate: %s and key: %s", certFile, keyFile)
	return certFile, keyFile, nil
}
