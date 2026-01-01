package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"log"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

var (
	outDir = flag.String("out", "certs", "Output directory for certificates")
)

func main() {
	flag.Parse()

	if err := os.MkdirAll(*outDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	log.Println("Generating CA...")
	caCert, caKey, err := generateCA()
	if err != nil {
		log.Fatalf("Failed to generate CA: %v", err)
	}
	saveCertAndKey("ca", caCert, caKey)

	log.Println("Generating Server Certificate...")
	srvCert, srvKey, err := generateCert("pulse-server", []string{"localhost", "127.0.0.1"}, caCert, caKey, true)
	if err != nil {
		log.Fatalf("Failed to generate server cert: %v", err)
	}
	saveCertAndKey("server", srvCert, srvKey)

	log.Println("Generating Agent Certificate...")
	// For MVP, we generate one agent cert. In production, this would be per-agent.
	agentCert, agentKey, err := generateCert("pulse-agent", nil, caCert, caKey, false)
	if err != nil {
		log.Fatalf("Failed to generate agent cert: %v", err)
	}
	saveCertAndKey("agent", agentCert, agentKey)

	log.Println("Done! Certificates generated in", *outDir)
}

func generateCA() (*x509.Certificate, *rsa.PrivateKey, error) {
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}

	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{Organization: []string{"Pulse CA"}},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(10 * 365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return nil, nil, err
	}

	cert, err := x509.ParseCertificate(certBytes)
	return cert, key, err
}

func generateCert(cn string, sans []string, caCert *x509.Certificate, caKey *rsa.PrivateKey, isServer bool) ([]byte, *rsa.PrivateKey, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      pkix.Name{CommonName: cn},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	if isServer {
		template.ExtKeyUsage = append(template.ExtKeyUsage, x509.ExtKeyUsageServerAuth)
		for _, san := range sans {
			if ip := net.ParseIP(san); ip != nil {
				template.IPAddresses = append(template.IPAddresses, ip)
			} else {
				template.DNSNames = append(template.DNSNames, san)
			}
		}
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, template, caCert, &key.PublicKey, caKey)
	return certBytes, key, err
}

func saveCertAndKey(name string, cert interface{}, key *rsa.PrivateKey) {
	var certBytes []byte
	switch c := cert.(type) {
	case *x509.Certificate:
		certBytes = c.Raw
	case []byte:
		certBytes = c
	}

	certOut, _ := os.Create(filepath.Join(*outDir, name+".crt"))
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	certOut.Close()

	keyOut, _ := os.Create(filepath.Join(*outDir, name+".key"))
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	keyOut.Close()
}
