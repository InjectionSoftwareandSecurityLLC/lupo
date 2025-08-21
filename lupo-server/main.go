package main

import (
	"flag"
	"net"
	"os"

	"github.com/InjectionSoftwareandSecurityLLC/lupo/lupo-server/cmd"
	"github.com/InjectionSoftwareandSecurityLLC/lupo/lupo-server/core"
	"github.com/desertbit/grumble"
)

func main() {
	// Optional TLS IP override flag
	tlsIP := flag.String("tls_ip", "", "Optional: override interface IP used in TLS cert SAN")
	lupo_rc := flag.String("r", "", "Optional: Specify a resource file for lupo automation")
	flag.Parse()
	os.Args = append([]string{os.Args[0]}, flag.Args()...)

	var ip net.IP
	if *tlsIP != "" {
		ip = net.ParseIP(*tlsIP)
		if ip == nil {
			core.LogData("❌ Invalid IP provided to -tls_ip: " + *tlsIP)
			ip = detectLocalIP()
		} else {
			core.LogData("📦 Using user-specified IP for TLS SAN: " + ip.String())
		}
	} else {
		ip = detectLocalIP()
		core.LogData("📡 Auto-detected local IP for TLS SAN: " + ip.String())
	}

	err := core.GenerateSelfSignedCert(ip.String(), "localhost", 3650)
	if err != nil {
		core.LogData("❌ TLS cert generation failed: " + err.Error())
	}
	
	if *lupo_rc != ""{
		go cmd.ExecuteResourceFile(*lupo_rc)
	}

	grumble.Main(cmd.App)
	core.LogData("Lupo C2 stopped!")
}

// detectLocalIP returns the first non-loopback IPv4 address or falls back to 127.0.0.1
func detectLocalIP() net.IP {
	ifaces, err := net.Interfaces()
	if err != nil {
		core.LogData("⚠️ Could not get network interfaces, falling back to 127.0.0.1")
		return net.ParseIP("127.0.0.1")
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue // skip down or loopback
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip != nil && !ip.IsLoopback() {
				if ipv4 := ip.To4(); ipv4 != nil {
					return ipv4
				}
			}
		}
	}
	core.LogData("⚠️ No non-loopback interface found, falling back to 127.0.0.1")
	return net.ParseIP("127.0.0.1")
}
