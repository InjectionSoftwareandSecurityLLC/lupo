package server

import (
	"fmt"
	"strings"
	"encoding/base64"
	"log"
	"os"
	"text/tabwriter"

	"github.com/miekg/dns"
)


func DNSServerHandler(w dns.ResponseWriter, r *dns.Msg) {
	remote := w.RemoteAddr().String()
	fmt.Println("\n📥 Received from", remote)

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.Debug)
	fmt.Fprintln(tw, "Question Name\tType\tClass")
	for _, q := range r.Question {
		fmt.Fprintf(tw, "%s\t%s\t%s\n", q.Name, dns.TypeToString[q.Qtype], dns.ClassToString[q.Qclass])
	}
	tw.Flush()

	msg := new(dns.Msg)
	msg.SetReply(r)

	for _, q := range r.Question {
		if q.Qtype == dns.TypeTXT {
			subdomainRaw := extractSubdomainPrefix(q.Name)
			subdomainDecoded := decodeBase64(subdomainRaw)
			clientPayload := extractClientPayload(r)

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
		}
	}

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