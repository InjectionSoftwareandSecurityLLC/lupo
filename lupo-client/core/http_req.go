package core

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

// WolfPackServer - configuration struct to tell the client how to communicate with the lupo server
type WolfPackServer struct {
	protocol string
	rhost    string
	rport    int
	userName string
	psk      string
	cert     string
}

var WolfPackServerConfig *WolfPackServer

var WolfPackHTTP *http.Client

// AuthURL - Primary auth URL scheme, populated by the config
var AuthURL string

// InitializeWolfPackRequests - Initializes a https request client that can be used for authenticated requests throughout the lupo client
func InitializeWolfPackRequests(configFile *string) error {

	err := ReadConfigFile(configFile)

	if err != nil {
		return err
	}

	AuthURL = WolfPackServerConfig.protocol + WolfPackServerConfig.rhost + ":" + strconv.Itoa(WolfPackServerConfig.rport) + "/?psk=" + WolfPackServerConfig.psk + "&user=" + WolfPackServerConfig.userName

	rootCert := WolfPackServerConfig.cert

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
		InitializeWolfPackRequests(configFile)

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

	return nil
}

// ReadConfigFile - reads in the configuration json file for the lupo client
func ReadConfigFile(configFile *string) error {

	var jsonFile *os.File
	var err error

	if *configFile != "" {
		jsonFile, err = os.Open(*configFile)
	}

	// if we os.Open returns an error then handle it
	if err != nil {
		return err
	}

	configJSON, _ := ioutil.ReadAll(jsonFile)

	var config map[string]interface{}

	err = json.Unmarshal(configJSON, &config)
	if err != nil {
		return errors.New("Could not parse json config file")
	}

	// JSON marshaling directly into struct hasn't worked so far, instead we'll type enforce the dynamic map and setup the config struct to use elsewhere
	WolfPackServerConfig = &WolfPackServer{
		protocol: config["protocol"].(string),
		rhost:    config["rhost"].(string),
		rport:    int(config["rport"].(float64)),
		userName: config["userName"].(string),
		psk:      config["psk"].(string),
		cert:     config["cert"].(string),
	}

	jsonFile.Close()

	return nil
}
