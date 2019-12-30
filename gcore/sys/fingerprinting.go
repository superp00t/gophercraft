package sys

import (
	context "context"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	fmt "fmt"
	"io/ioutil"
	"strings"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
)

func GetCertFileFingerprint(at string) (string, error) {
	b, err := ioutil.ReadFile(at)
	if err != nil {
		return "", err
	}

	pblock, _ := pem.Decode(b)

	cert, err := x509.ParseCertificate(pblock.Bytes)
	if err != nil {
		return "", nil
	}

	return GetCertFingerprint(cert)
}

func GetCertFingerprint(cert *x509.Certificate) (string, error) {
	der, err := x509.MarshalPKIXPublicKey(cert.PublicKey)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(der)
	hx := strings.ToUpper(hex.EncodeToString(hash[:]))

	return hx, nil
}

func GetPeerFingerprint(ctx context.Context) (string, error) {
	peer, ok := peer.FromContext(ctx)
	if !ok {
		return "", fmt.Errorf("could not extract peer information")
	}

	tlsInfo, ok := peer.AuthInfo.(credentials.TLSInfo)
	if !ok {
		return "", fmt.Errorf("peer does not have tls info")
	}

	certCount := len(tlsInfo.State.PeerCertificates)

	if certCount != 1 {
		return "", fmt.Errorf("invalid certificate count (%d)", certCount)
	}

	cert := tlsInfo.State.PeerCertificates[0]

	return GetCertFingerprint(cert)
}
