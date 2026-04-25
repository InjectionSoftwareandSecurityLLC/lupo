// Package embedded provides the Go embed declarations for bundling Messenger
// source and the Lupo proxy shim into the compiled binary.
package embedded

import "embed"

// MessengerFS holds the entire Messenger source tree, extracted at runtime
// to ~/.lupo/messenger-src/ and installed into a dedicated Python venv.
//
// The "all:" prefix is required so that Python files beginning with "_"
// (e.g. __init__.py) are included — Go's default embed behaviour excludes them.
//
//go:embed all:messenger
var MessengerFS embed.FS

// ProxyShim holds the lupo-proxy-shim.py script that bridges Lupo's Go process
// management layer with the Messenger Python library via JSON on stdin/stdout.
//
//go:embed shim/lupo-proxy-shim.py
var ProxyShim []byte
