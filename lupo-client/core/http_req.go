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

var rootCert string = `


`

// AuthURL - Global variable used to handle requests through the client, additional paramaters can be appended on to it
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
