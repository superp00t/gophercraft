package config

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"os"
	"os/exec"
	"time"
)

func genPair(public, private string) error {
	vf := 365 * 24 * time.Hour
	validFor := &vf
	var err error
	var notBefore time.Time
	notBefore = time.Now()

	notAfter := notBefore.Add(*validFor)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return fmt.Errorf("failed to generate serial number: %s", err)
	}
	rootKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}
	keyToFile(private, rootKey)

	rootTemplate := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Gophercraft"},
			CommonName:   "GC",
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &rootTemplate, &rootTemplate, &rootKey.PublicKey, rootKey)
	if err != nil {
		return err
	}
	certToFile(public, derBytes)
	return nil
}

// keyToFile writes a PEM serialization of |key| to a new file called
// |filename|.
func keyToFile(filename string, key *ecdsa.PrivateKey) {
	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	b, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to marshal ECDSA private key: %v", err)
		os.Exit(2)
	}
	if err := pem.Encode(file, &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}); err != nil {
		panic(err)
	}
}

func certToFile(filename string, derBytes []byte) {
	certOut, err := os.Create(filename)
	if err != nil {
		log.Fatalf("failed to open cert.pem for writing: %s", err)
	}
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		log.Fatalf("failed to write data to cert.pem: %s", err)
	}
	if err := certOut.Close(); err != nil {
		log.Fatalf("error closing cert.pem: %s", err)
	}
}

// debugCertToFile writes a PEM serialization and OpenSSL debugging dump of
// |derBytes| to a new file called |filename|.
func debugCertToFile(filename string, derBytes []byte) {
	cmd := exec.Command("openssl", "x509", "-text", "-inform", "DER")

	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	cmd.Stdout = file
	cmd.Stderr = os.Stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		panic(err)
	}

	if err := cmd.Start(); err != nil {
		panic(err)
	}
	if _, err := stdin.Write(derBytes); err != nil {
		panic(err)
	}
	stdin.Close()
	if err := cmd.Wait(); err != nil {
		panic(err)
	}
}
