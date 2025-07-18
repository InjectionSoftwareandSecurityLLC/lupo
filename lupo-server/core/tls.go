package core
import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"strconv"
	"path/filepath"
	"time"
)

func GenerateSelfSignedCert(ip, host string, days int) error {
	certDir := "tls-certs"
	keyPath := filepath.Join(certDir, "lupo-server.key")
	crtPath := filepath.Join(certDir, "lupo-server.crt")
	pemPath := filepath.Join(certDir, "lupo-server.pem")

	if _, err := os.Stat(keyPath); err == nil {
		if _, err := os.Stat(crtPath); err == nil {
			LogData("🔒 TLS key and cert already exist in " + certDir + "skipping generation.")
			return nil
		}
	}

	LogData("📜 Generating self-signed TLS certificate for host=" + host +", ip=" + ip + " (valid " + strconv.Itoa(days) + ")")

	if err := os.MkdirAll(certDir, 0700); err != nil {
		return fmt.Errorf("failed to create cert directory: %w", err)
	}

	priv, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return fmt.Errorf("failed to generate key: %w", err)
	}

	notBefore := time.Now()
	notAfter := notBefore.AddDate(0, 0, days)
	serialNumber, _ := rand.Int(rand.Reader, big.NewInt(1<<62))

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Country:            []string{"US"},
			Province:           []string{"Lupo"},
			Locality:           []string{"Lupo"},
			Organization:       []string{"Lupo"},
			OrganizationalUnit: []string{"Lupo"},
			CommonName:         host,
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	if ipAddr := net.ParseIP(ip); ipAddr != nil {
		template.IPAddresses = []net.IP{ipAddr}
	}
	if host != "" {
		template.DNSNames = []string{host}
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %w", err)
	}

	// Write private key
	keyOut, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to open key file: %w", err)
	}
	defer keyOut.Close()
	privBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return fmt.Errorf("unable to marshal EC key: %w", err)
	}
	pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes})

	// Write certificate
	certOut, err := os.Create(crtPath)
	if err != nil {
		return fmt.Errorf("failed to open cert file: %w", err)
	}
	defer certOut.Close()
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	// Write PEM file (key + cert)
	pemOut, err := os.Create(pemPath)
	if err != nil {
		return fmt.Errorf("failed to create PEM file: %w", err)
	}
	defer pemOut.Close()
	pem.Encode(pemOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes})
	pem.Encode(pemOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	LogData("✅ TLS certs written to " + certDir)
	return nil
}
