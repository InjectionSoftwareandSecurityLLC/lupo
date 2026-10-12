// server - the "server" package is used to organize all server handler functionality of the application.
//
// Any functionality related to handling connections, data, and server configurations are parsed and organized by this package.
package server

import (
	"encoding/json"
	"strings"
)

// hasPSKField distinguishes an omitted PSK from an explicitly empty one.
func hasPSKField(data []byte) bool {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return false
	}

	for field := range fields {
		if strings.EqualFold(field, "PSK") {
			return true
		}
	}

	return false
}

// PSK - Global variable for the Pre-shared key used in implant authentication
var PSK string
