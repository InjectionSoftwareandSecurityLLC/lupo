package server

import (
	"fmt"
	"strings"
	"encoding/base64"
	"log"
	//"os"
	//"text/tabwriter"
	"strconv"
	"encoding/json"
	"errors"
	"sync"
	"time"
    "net"

	"github.com/InjectionSoftwareandSecurityLLC/lupo/lupo-server/core"
	"github.com/miekg/dns"
)

type DNSListener struct {
	chunkBuffers map[string]*MessageChunks
	chunkMutex   sync.RWMutex
	chunkMap     map[string]*ChunkTracker // Track chunks per implant
}

type MessageChunks struct {
	chunks    map[int]string
	total     int
	lastSeen  time.Time
}

type ChunkTracker struct {
	Chunks      map[int]string
	TotalChunks int
	LastUpdate  time.Time
}

var sessionsMu sync.RWMutex

// registrationFragments holds fragment buffers for requests that arrive before
// a session exists (e.g., initial registration). Keyed by remote address.
var registrationFragments = struct {
	sync.Mutex
	m map[string][]string
}{m: make(map[string][]string)}

func DNSServerHandler(w dns.ResponseWriter, r *dns.Msg) {

	var dnsParams core.DNSData
	remoteAddrFull := w.RemoteAddr().String()
	// Use remote IP (without port) to buffer registration fragments so clients
	// that use ephemeral source ports still reassemble correctly.
	remoteIP := remoteAddrFull
	if host, _, err := net.SplitHostPort(remoteAddrFull); err == nil {
		remoteIP = host
	}
	dnsParams.Register = false

	var additionalFunctions map[string]interface{}

	//fmt.Println("\n📥 Received from", remoteAddrFull)

	/*tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.Debug)
	fmt.Fprintln(tw, "Question Name\tType\tClass")
	for _, q := range r.Question {
		fmt.Fprintf(tw, "%s\t%s\t%s\n", q.Name, dns.TypeToString[q.Qtype], dns.ClassToString[q.Qclass])
	}
	tw.Flush()*/

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

	var fullBase64 string
	var session core.Session

	if ok {
		// session exists -> store fragments in session.SubDomainFragments
		session = sVal.(core.Session)

		// Lock session mutex to safely update fragments (narrow scope)
		sessionsMu.Lock()

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

		// Release the session lock now that fragments are updated and we know
		// whether the message is complete.
		// Persist updated session back to core.Sessions
		core.Sessions.Store(sessionID, session)
		sessionsMu.Unlock()

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

		// All chunks received - join into fullBase64 from session fragments
		fullBase64 = strings.Join(session.SubDomainFragments, "")

		// Clear fragments for next message
		sessionsMu.Lock()
		session.SubDomainFragments = nil
		// Persist cleared fragments
		core.Sessions.Store(sessionID, session)
		sessionsMu.Unlock()

	} else {
		// No session exists for provided sessionID. This is expected for
		// initial registration (we use remoteAddr as a buffer key).
		registrationFragments.Lock()
		frags, exists := registrationFragments.m[remoteIP]
		if !exists || len(frags) < totalChunks {
			newFrags := make([]string, totalChunks)
			if exists {
				copy(newFrags, frags)
			}
			frags = newFrags
		}

		frags[chunkIndex] = chunkPayloadBase64
		registrationFragments.m[remoteIP] = frags
		registrationFragments.Unlock()

		log.Printf("Received reg chunk %d/%d from %s", chunkIndex+1, totalChunks, remoteIP)

		// Check if all registration chunks received
		complete := true
		for _, frag := range frags {
			if frag == "" {
				complete = false
				break
			}
		}

		if !complete {
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

		// Assemble full payload and remove buffer
	registrationFragments.Lock()
	fullBase64 = strings.Join(frags, "")
	delete(registrationFragments.m, remoteIP)
	registrationFragments.Unlock()
	}


	// fullBase64 contains the reassembled payload (from session or registration buffers)
	jsonBytes, err := base64.RawURLEncoding.DecodeString(fullBase64)
	if err != nil {
		log.Printf("Failed to decode base64 full message for session %d: %v", sessionID, err)
		w.WriteMsg(msg)
		return
	}

	// Parse JSON message into dnsParams struct
	err = json.Unmarshal(jsonBytes, &dnsParams)
	if err != nil {
		log.Printf("Failed to unmarshal JSON for session %d: %v", sessionID, err)
		w.WriteMsg(msg)
		return
	}

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

			core.RegisterSession(core.SessionID, "DNS", implant, remoteAddrFull, 0, "", "", "", "")

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

			// Return plain JSON in TXT (consistent with HTTP handler responses)
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

			core.RegisterSession(core.SessionID, "DNS", implant, remoteAddrFull, 0, "", "", "", "")

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

			// Return plain JSON in TXT for persistence responses as well
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

			core.RegisterSession(core.SessionID, "DNS", implant, remoteAddrFull, 0, "", "", "", "")

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

// parseInt is a convenience wrapper used by the chunk-based handler
func parseInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

// handleImplantMessage processes a fully-assembled JSON message from an implant.
// This mirrors the logic in the DNSServerHandler but does not write a DNS reply
// (the chunk-based path already sent an acknowledgment). remoteAddr is used
// for session registration/logging.
func (d *DNSListener) handleImplantMessage(decoded string, remoteAddr net.Addr) (string, error) {
	var dnsParams core.DNSData
	var additionalFunctions map[string]interface{}

	err := json.Unmarshal([]byte(decoded), &dnsParams)
	if err != nil {
		log.Printf("Failed to unmarshal assembled JSON from DNS implant: %v", err)
		return "", err
	}

	raddr := ""
	if remoteAddr != nil {
		raddr = remoteAddr.String()
	}

	if dnsParams.PSK == "" {
		errorString := "DNS Request did not provide PSK, request ignored"
		core.LogData(errorString)
		returnErr := errors.New(errorString)
		ErrorHandler(returnErr)
		return "", returnErr
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
			core.RegisterSession(core.SessionID, "DNS", implant, raddr, 0, "", "", "", "")
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
				return "", err
			}

			core.BroadcastSession(strconv.Itoa(newSession))
			return string(jsonResp), nil
		}
	} else {
		errorString := "DNS Request Invalid PSK, request ignored"
		core.LogData(errorString)
		returnErr := errors.New(errorString)
		ErrorHandler(returnErr)
		return "", returnErr
	}

	// Try to load session by provided SessionID
	sVal, ok := core.Sessions.Load(dnsParams.SessionID)
	if !ok {
		if core.PersistenceMode {
			reconnectString := "Old implant with UUID: " + dnsParams.UUID.String() + " connected, attempting to reestablish session..."
			core.LogData(reconnectString)
			core.WarningColorBold.Println(reconnectString)

			implant := core.RegisterImplant(dnsParams.ImplantArch, dnsParams.Update, additionalFunctions, dnsParams.UUID.String())
			core.RegisterSession(core.SessionID, "DNS", implant, raddr, 0, "", "", "", "")
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
				return "", err
			}

			core.BroadcastSession(strconv.Itoa(newSession))
			return string(jsonResp), nil
		} else {
			errorString := "DNS Request Invalid UUID, request ignored"
			core.LogData(errorString)
			returnErr := errors.New(errorString)
			ErrorHandler(returnErr)
			return "", returnErr
		}
	}

	session := sVal.(core.Session)

	if session.Implant.ID != dnsParams.UUID || dnsParams.UUID == core.ZeroedUUID {
		if core.PersistenceMode {
			reconnectString := "Old implant with UUID: " + dnsParams.UUID.String() + " connected, attempting to reestablish session..."
			core.LogData(reconnectString)
			core.WarningColorBold.Println(reconnectString)

			implant := core.RegisterImplant(dnsParams.ImplantArch, dnsParams.Update, additionalFunctions, dnsParams.UUID.String())
			core.RegisterSession(core.SessionID, "DNS", implant, raddr, 0, "", "", "", "")
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
				return "", err
			}

			core.BroadcastSession(strconv.Itoa(newSession))
			return string(jsonResp), nil
		} else {
			errorString := "DNS Request Invalid UUID, request ignored"
			core.LogData(errorString)
			returnErr := errors.New(errorString)
			ErrorHandler(returnErr)
			return "", returnErr
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

	// Update implant state and check-in
	core.UpdateImplant(dnsParams.SessionID, dnsParams.Update, dnsParams.ImplantArch, additionalFunctions)
	core.SessionCheckIn(dnsParams.SessionID)

	// Prepare cmd/user response similar to DNSServerHandler
	var cmd string
	var user string

	sVal, _ = core.Sessions.Load(dnsParams.SessionID)
	if sVal != nil {
		sess := sVal.(core.Session)
		if sess.Implant.Commands != nil && len(sess.Implant.Commands) > 0 {
			cmd = sess.Implant.Commands[0].Command
			user = sess.Implant.Commands[0].Operator
		}
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
		return "", err
	}

	return string(jsonResp), nil
}

func (d *DNSListener) handleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	msg := new(dns.Msg)
	msg.SetReply(r)
	
	if len(r.Question) == 0 {
		return
	}

	question := r.Question[0]
	if question.Qtype != dns.TypeTXT {
		return
	}

	// Parse DNS query format: <chunk_index>.<total_chunks>.<encoded_psk>.<uuid>.<chunk>.domain
	parts := strings.Split(question.Name, ".")
	if len(parts) < 6 {
		log.Printf("error: Problem occurred while parsing input from a DNS based implant")
		return
	}

	// Extract and decode PSK (expect Raw URL encoding)
	encodedPSK := parts[2]
	decodedPSK, err := base64.RawURLEncoding.DecodeString(encodedPSK)
	if err != nil || string(decodedPSK) != PSK {
		log.Printf("DNS Request did not provide valid PSK, request ignored")
		return
	}

	// Parse chunk metadata
	chunkIndex := parseInt(parts[0])
	totalChunks := parseInt(parts[1])
	implantID := parts[3]
	encodedData := parts[4]

	// Process chunk
	complete, assembledData := d.processChunk(implantID, chunkIndex, totalChunks, encodedData)

	// If this was the last chunk, process the complete message and return JSON
	if complete {
		// Decode the complete base64 message (Raw URL encoding)
		decodedData, err := base64.RawURLEncoding.DecodeString(assembledData)
		if err != nil {
			log.Printf("Failed to decode assembled data: %v", err)
			// fallback ack
			txt := &dns.TXT{
				Hdr: dns.RR_Header{
					Name:   question.Name,
					Rrtype: dns.TypeTXT,
					Class:  dns.ClassINET,
					Ttl:    0,
				},
				Txt: []string{"ack"},
			}
			msg.Answer = append(msg.Answer, txt)
			w.WriteMsg(msg)
			return
		}

		// Process the complete message through existing handler and capture reply
		jsonReply, err := d.handleImplantMessage(string(decodedData), w.RemoteAddr())
		if err != nil || jsonReply == "" {
			// fallback ack
			txt := &dns.TXT{
				Hdr: dns.RR_Header{
					Name:   question.Name,
					Rrtype: dns.TypeTXT,
					Class:  dns.ClassINET,
					Ttl:    0,
				},
				Txt: []string{"ack"},
			}
			msg.Answer = append(msg.Answer, txt)
			w.WriteMsg(msg)
			return
		}

		// Return JSON reply as TXT to implant
		txt := &dns.TXT{
			Hdr: dns.RR_Header{
				Name:   question.Name,
				Rrtype: dns.TypeTXT,
				Class:  dns.ClassINET,
				Ttl:    0,
			},
			Txt: []string{jsonReply},
		}
		msg.Answer = append(msg.Answer, txt)
		w.WriteMsg(msg)
		return
	}

	// For partial chunks, send ack
	txt := &dns.TXT{
		Hdr: dns.RR_Header{
			Name:   question.Name,
			Rrtype: dns.TypeTXT,
			Class:  dns.ClassINET,
			Ttl:    0,
		},
		Txt: []string{"ack"},
	}
	msg.Answer = append(msg.Answer, txt)
	w.WriteMsg(msg)
}

func (d *DNSListener) processChunk(implantID string, index, total int, data string) (bool, string) {
	d.chunkMutex.Lock()
	defer d.chunkMutex.Unlock()

	// Initialize tracker if needed
	if _, exists := d.chunkMap[implantID]; !exists {
		d.chunkMap[implantID] = &ChunkTracker{
			Chunks:      make(map[int]string),
			TotalChunks: total,
			LastUpdate:  time.Now(),
		}
	}

	tracker := d.chunkMap[implantID]
	tracker.Chunks[index] = data
	tracker.LastUpdate = time.Now()

	// Check if we have all chunks
	if len(tracker.Chunks) == tracker.TotalChunks {
		assembled := d.assembleChunks(implantID)
		delete(d.chunkMap, implantID)
		return true, assembled
	}

	return false, ""
}

func (d *DNSListener) assembleChunks(implantID string) string {
	tracker := d.chunkMap[implantID]
	result := make([]string, tracker.TotalChunks)
	
	for i := 0; i < tracker.TotalChunks; i++ {
		if chunk, ok := tracker.Chunks[i]; ok {
			result[i] = chunk
		}
	}
	
	return strings.Join(result, "")
}

func (d *DNSListener) cleanupStaleChunks() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		d.chunkMutex.Lock()
		now := time.Now()
		for id, tracker := range d.chunkMap {
			if now.Sub(tracker.LastUpdate) > 5*time.Minute {
				delete(d.chunkMap, id)
			}
		}
		d.chunkMutex.Unlock()
	}
}

func NewDNSListener(address string, port int, psk string) *DNSListener {
	listener := &DNSListener{
		// ...existing initialization...
		chunkBuffers: make(map[string]*MessageChunks),
		chunkMap:     make(map[string]*ChunkTracker),
	}

	// Start cleanup routine
	go listener.cleanupStaleChunks()

	return listener
}
