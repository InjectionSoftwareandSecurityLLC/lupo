package server

import "net/http"

// StagerServerHandler - returns an http.HandlerFunc that serves static files from the given directory.
// This handler does nothing except serve files — no PSK validation, no session handling, no implant routing.
func StagerServerHandler(dir string) http.HandlerFunc {
	return http.FileServer(http.Dir(dir)).ServeHTTP
}
