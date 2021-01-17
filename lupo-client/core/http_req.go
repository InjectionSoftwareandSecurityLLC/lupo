package core

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
)

// Configuration
type WolfPackServer struct {
	updateInterval int
	protocol       string
	rhost          string
	rport          int
	userName       string
	psk            string
	data           string
}

var WolfPackServerConfig *WolfPackServer

var WolfPackHTTP *http.Client

var rootCert string = `-----BEGIN CERTIFICATE-----
MIICYjCCAeigAwIBAgIUW7t41gvdDM4PrOcQpERIamH7cvQwCgYIKoZIzj0EAwIw
XTELMAkGA1UEBhMCVVMxDTALBgNVBAgMBEx1cG8xDTALBgNVBAcMBEx1cG8xDTAL
BgNVBAoMBEx1cG8xDTALBgNVBAsMBEx1cG8xEjAQBgNVBAMMCWxvY2FsaG9zdDAe
Fw0yMTAxMTEwMzAyMThaFw0zMTAxMDkwMzAyMThaMF0xCzAJBgNVBAYTAlVTMQ0w
CwYDVQQIDARMdXBvMQ0wCwYDVQQHDARMdXBvMQ0wCwYDVQQKDARMdXBvMQ0wCwYD
VQQLDARMdXBvMRIwEAYDVQQDDAlsb2NhbGhvc3QwdjAQBgcqhkjOPQIBBgUrgQQA
IgNiAAR/5MWWRnNRZ7GbBx9oU98WrvYiXCWgRpkWvCaYZt4kFgnO7jZmYO5cae2W
OBGfJHcaFa85K+NhURQdD/m1LN1Vqwzp3pCyjgadUU94Y3rz/2vBPfOOyL9Ch19d
KNyDVMqjaTBnMB0GA1UdDgQWBBQAF4Pln4oYpsZ2z9sQTPF6B0PgbDAfBgNVHSME
GDAWgBQAF4Pln4oYpsZ2z9sQTPF6B0PgbDAPBgNVHRMBAf8EBTADAQH/MBQGA1Ud
EQQNMAuCCWxvY2FsaG9zdDAKBggqhkjOPQQDAgNoADBlAjEA8OO/tGsG9DY0Fqtd
JOfhv1XW+H7gA5H+f/8nToNGXxvYuXZjD7SHfz0+0li1J9eXAjBD3b1A0PCcZaee
3L92USeXWa2gFV4e1zjRmZbZTTwljLtydC8mSUOJH6KKzjn+tnQ=
-----END CERTIFICATE-----
`

// AuthURL - Primary auth URL scheme, needs to be parameterized
var AuthURL = "https://localhost:3074/?psk=wolfpack&user=3ndG4me"

func InitializeWolfPackRequests() {
	WolfPackServerConfig = &WolfPackServer{
		updateInterval: 1,
		protocol:       "https://",
		rhost:          "localhost",
		rport:          3074,
		userName:       "3ndG4me",
		psk:            "wolfpack",
		data:           "",
	}

	// If a root certificate is specified, use it
	config := &tls.Config{}
	if rootCert != "" {
		// Create new cert pool
		rootCAs := x509.NewCertPool()

		// Add cert to certpool
		rootCAs.AppendCertsFromPEM([]byte(rootCert))

		// Trust the certpool
		config = &tls.Config{
			InsecureSkipVerify: false,
			RootCAs:            rootCAs,
		}

	} else {

		// Recurse and try again, failure is not an option
		InitializeWolfPackRequests()

		/*
			// Otherwise accept any ssl cert
			config = &tls.Config{
				InsecureSkipVerify: true,
			}
		*/
	}

	// Create http client
	WolfPackHTTP = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: config,
		},
	}
}
