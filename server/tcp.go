package server

import (
	"bufio"
	"fmt"
	"net"

	"github.com/InjectionSoftwareandSecurityLLC/lupo/core"
)

// TCPServerHandler - handles incoming TCP connections
func TCPServerHandler(conn net.Conn) {
	fmt.Printf("Serving %s\n", conn.RemoteAddr().String())

	var cmd string

	if core.Sessions[0].Implant.Commands != nil {
		cmd = core.Sessions[0].Implant.Commands[0]
	}
	/* No JSON for now til we figure this out
	response := map[string]interface{}{
		"cmd": cmd,
	}

	json.NewEncoder(w).Encode(response)
	*/

	// hardcode for testing for now
	cmd = "id"

	core.UpdateImplant(0, 0, nil)

	core.SessionCheckIn(0)
	conn.Write([]byte(cmd))

	netData, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(netData)

}
