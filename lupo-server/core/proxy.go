package core

import (
	"bufio"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/InjectionSoftwareandSecurityLLC/lupo/lupo-server/embedded"
)

// ---------------------------------------------------------------------------
// Constants / paths
// ---------------------------------------------------------------------------

// lupoDir returns the base Lupo data directory (~/.lupo).
func lupoDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".lupo"
	}
	return filepath.Join(home, ".lupo")
}

func messengerSrcDir() string  { return filepath.Join(lupoDir(), "messenger-src") }
func messengerVenvDir() string { return filepath.Join(lupoDir(), "messenger-env") }
func shimPath() string         { return filepath.Join(lupoDir(), "lupo-proxy-shim.py") }
func versionStampPath() string { return filepath.Join(lupoDir(), ".messenger_version") }

// python3 binary inside the managed venv
func venvPython() string { return filepath.Join(messengerVenvDir(), "bin", "python3") }

// ---------------------------------------------------------------------------
// ProxyState — tracks one live Messenger shim per session
// ---------------------------------------------------------------------------

// ProxyForwarder carries details about a single active SOCKS5 listener or TCP
// port forwarder. IDs are Lupo-assigned and shown in 'proxy show'.
type ProxyForwarder struct {
	ID              int
	Type            string // "socks" | "forward"
	ListeningHost   string
	ListeningPort   string
	DestinationHost string // empty for SOCKS
	DestinationPort string // empty for SOCKS
	port            int    // raw port for shim stop_socks
	config          string // raw config string for shim stop_forward
}

// ProxyState holds all runtime information for one Messenger proxy server instance.
type ProxyState struct {
	ID          int
	Lhost       string
	Cmd         *exec.Cmd
	Stdin       io.WriteCloser
	stdout      *bufio.Scanner
	EncKey      string
	ServerURL   string
	ServerPort  int
	Status      string   // "starting" | "active" | "stopped"
	MessengerID string   // most recently connected client ID
	Forwarders  []ProxyForwarder
	mu          sync.Mutex
	subMu       sync.Mutex
	subs        map[string][]chan map[string]interface{}
}

// subscribe registers a one-shot channel that receives the next shim event of
// the given type. Used to synchronise command callers with the background event loop.
func (p *ProxyState) subscribe(event string) <-chan map[string]interface{} {
	ch := make(chan map[string]interface{}, 1)
	p.subMu.Lock()
	if p.subs == nil {
		p.subs = make(map[string][]chan map[string]interface{})
	}
	p.subs[event] = append(p.subs[event], ch)
	p.subMu.Unlock()
	return ch
}

// GetForwarders returns a thread-safe snapshot of the proxy's forwarder list.
func (p *ProxyState) GetForwarders() []ProxyForwarder {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]ProxyForwarder, len(p.Forwarders))
	copy(out, p.Forwarders)
	return out
}

// publish delivers ev to all one-shot subscribers waiting for that event type.
func (p *ProxyState) publish(ev map[string]interface{}) {
	event := getString(ev, "event")
	p.subMu.Lock()
	chans := p.subs[event]
	p.subs[event] = nil
	p.subMu.Unlock()
	for _, ch := range chans {
		select {
		case ch <- ev:
		default:
		}
	}
}

// sendCommand writes a JSON command to the shim's stdin.
func (p *ProxyState) sendCommand(payload map[string]interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(p.Stdin, "%s\n", data)
	return err
}

// readEvent reads the next JSON event line from the shim's stdout.
func (p *ProxyState) readEvent() (map[string]interface{}, error) {
	if !p.stdout.Scan() {
		if err := p.stdout.Err(); err != nil {
			return nil, err
		}
		return nil, io.EOF
	}
	var ev map[string]interface{}
	if err := json.Unmarshal([]byte(p.stdout.Text()), &ev); err != nil {
		return nil, fmt.Errorf("invalid shim event JSON: %w", err)
	}
	return ev, nil
}

// ---------------------------------------------------------------------------
// Global proxy registry — ID-keyed map supporting multiple concurrent servers
// ---------------------------------------------------------------------------

var (
	proxyMap    = map[int]*ProxyState{}
	proxyMapMu  sync.Mutex
	nextProxyID = 1

	// Global forwarder ID counter — unique across all proxies so users can
	// reference a forwarder by ID without also specifying a proxy ID.
	nextFwdID   = 1
	nextFwdIDMu sync.Mutex
)

func allocFwdID() int {
	nextFwdIDMu.Lock()
	defer nextFwdIDMu.Unlock()
	id := nextFwdID
	nextFwdID++
	return id
}

// GetProxy returns the ProxyState for the given ID.
func GetProxy(id int) (*ProxyState, bool) {
	proxyMapMu.Lock()
	defer proxyMapMu.Unlock()
	s, ok := proxyMap[id]
	return s, ok
}

// GetAllProxies returns a snapshot of all running proxy states sorted by ID.
func GetAllProxies() []*ProxyState {
	proxyMapMu.Lock()
	defer proxyMapMu.Unlock()
	out := make([]*ProxyState, 0, len(proxyMap))
	for _, s := range proxyMap {
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// ---------------------------------------------------------------------------
// Messenger installation
// ---------------------------------------------------------------------------

// embeddedHash returns a short hash derived from the embedded shim bytes so
// that a version change in the binary triggers a reinstall.
func embeddedHash() string {
	sum := sha256.Sum256(embedded.ProxyShim)
	return base64.StdEncoding.EncodeToString(sum[:8])
}

// EnsureMessengerInstalled verifies the managed venv is present and up to
// date. It installs or reinstalls automatically when needed. Returns the path
// to the venv's python3 interpreter.
func EnsureMessengerInstalled() (string, error) {
	stamp := versionStampPath()
	current := embeddedHash()

	// If the venv exists and the version stamp matches we're good.
	if _, err := os.Stat(venvPython()); err == nil {
		if data, err := os.ReadFile(stamp); err == nil {
			if strings.TrimSpace(string(data)) == current {
				return venvPython(), nil
			}
		}
	}

	SuccessColorBold.Println("[*] Messenger not found or outdated — performing setup (this runs once)...")

	if err := extractMessengerAssets(); err != nil {
		return "", fmt.Errorf("failed to extract Messenger assets: %w", err)
	}

	if err := createVenv(); err != nil {
		return "", fmt.Errorf("failed to create Python venv: %w", err)
	}

	if err := pipInstall(); err != nil {
		return "", fmt.Errorf("failed to install Messenger into venv: %w", err)
	}

	// Write version stamp
	if err := os.WriteFile(stamp, []byte(current), 0644); err != nil {
		return "", fmt.Errorf("failed to write version stamp: %w", err)
	}

	SuccessColorBold.Println("[+] Messenger setup complete.")
	return venvPython(), nil
}

// extractMessengerAssets writes embedded.MessengerFS and embedded.ProxyShim
// to disk under ~/.lupo/.
func extractMessengerAssets() error {
	srcDir := messengerSrcDir()
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		return err
	}

	// Walk the embedded FS and recreate the tree under srcDir.
	// The embed root is "messenger/", so we strip that prefix when writing.
	err := fs.WalkDir(embedded.MessengerFS, "messenger", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Strip leading "messenger/" prefix
		rel := strings.TrimPrefix(path, "messenger")
		rel = strings.TrimPrefix(rel, string(filepath.Separator))
		if rel == "" {
			return nil
		}
		dest := filepath.Join(srcDir, rel)
		if d.IsDir() {
			return os.MkdirAll(dest, 0755)
		}
		data, err := embedded.MessengerFS.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dest, data, 0644)
	})
	if err != nil {
		return err
	}

	// Write the shim script
	if err := os.MkdirAll(lupoDir(), 0755); err != nil {
		return err
	}
	return os.WriteFile(shimPath(), embedded.ProxyShim, 0755)
}

// createVenv runs `python3 -m venv` to create the managed virtualenv.
func createVenv() error {
	python3, err := exec.LookPath("python3")
	if err != nil {
		return errors.New("python3 not found in PATH — Messenger requires Python 3 (https://python.org)")
	}
	cmd := exec.Command(python3, "-m", "venv", messengerVenvDir())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// pipInstall installs the extracted Messenger source into the venv.
func pipInstall() error {
	pip := filepath.Join(messengerVenvDir(), "bin", "pip")
	// Install with --recurse-submodules-equivalent: the source tree is already
	// fully extracted so a plain pip install from the directory works.
	cmd := exec.Command(pip, "install", messengerSrcDir())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ---------------------------------------------------------------------------
// Proxy lifecycle
// ---------------------------------------------------------------------------

// StartProxy starts a new Messenger shim and waits for its "ready" event.
// Multiple proxy servers can run simultaneously; each gets a unique ID.
func StartProxy(lhost string, serverPort int) (*ProxyState, error) {
	pyBin, err := EnsureMessengerInstalled()
	if err != nil {
		return nil, err
	}

	args := []string{
		shimPath(),
		"--lhost", lhost,
		"--server-port", strconv.Itoa(serverPort),
	}

	cmd := exec.Command(pyBin, args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("could not open shim stdin pipe: %w", err)
	}
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("could not open shim stdout pipe: %w", err)
	}
	cmd.Stderr = os.Stderr // shim stderr → lupo stderr, kept separate from events

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start proxy shim: %w", err)
	}

	state := &ProxyState{
		Lhost:      lhost,
		Cmd:        cmd,
		Stdin:      stdin,
		stdout:     bufio.NewScanner(stdoutPipe),
		ServerPort: serverPort,
		Status:     "starting",
	}

	// Wait for the "ready" event (with timeout)
	readyCh := make(chan error, 1)
	go func() {
		ev, err := state.readEvent()
		if err != nil {
			readyCh <- err
			return
		}
		if ev["event"] != "ready" {
			readyCh <- fmt.Errorf("unexpected first event from shim: %v", ev)
			return
		}
		state.EncKey = getString(ev, "enc_key")
		state.ServerURL = getString(ev, "server_url")
		state.Status = "active"
		readyCh <- nil
	}()

	select {
	case err := <-readyCh:
		if err != nil {
			_ = cmd.Process.Kill()
			return nil, fmt.Errorf("proxy shim failed to start: %w", err)
		}
	case <-time.After(15 * time.Second):
		_ = cmd.Process.Kill()
		return nil, errors.New("proxy shim did not become ready within 15s")
	}

	// Background goroutine: read remaining events, update state, and fan out to
	// any command callers waiting on a specific event type via subscribe().
	go func() {
		for {
			ev, err := state.readEvent()
			if err != nil {
				state.mu.Lock()
				state.Status = "stopped"
				state.mu.Unlock()
				return
			}
			applyEvent(state, ev)
			state.publish(ev)
		}
	}()

	proxyMapMu.Lock()
	id := nextProxyID
	nextProxyID++
	state.ID = id
	proxyMap[id] = state
	proxyMapMu.Unlock()

	return state, nil
}

// StopProxy sends a stop command to the shim for the given proxy ID and cleans up.
func StopProxy(id int) error {
	proxyMapMu.Lock()
	state := proxyMap[id]
	delete(proxyMap, id)
	proxyMapMu.Unlock()

	if state == nil {
		return fmt.Errorf("no proxy with id %d", id)
	}

	_ = state.sendCommand(map[string]interface{}{"cmd": "stop"})
	time.Sleep(300 * time.Millisecond)
	if state.Cmd.ProcessState == nil {
		_ = state.Cmd.Process.Kill()
	}
	state.Status = "stopped"
	return nil
}

// ProxySOCKS instructs the shim for proxy id to open a SOCKS5 listener.
func ProxySOCKS(id int, port int) error {
	state, ok := GetProxy(id)
	if !ok {
		return fmt.Errorf("no proxy with id %d", id)
	}
	if state.MessengerID == "" {
		return errors.New("no Messenger client connected yet — wait for callback")
	}
	return state.sendCommand(map[string]interface{}{"cmd": "socks", "port": port})
}

// ProxyForward instructs the shim for proxy id to open a local port forward.
func ProxyForward(id int, config string) error {
	state, ok := GetProxy(id)
	if !ok {
		return fmt.Errorf("no proxy with id %d", id)
	}
	if state.MessengerID == "" {
		return errors.New("no Messenger client connected yet — wait for callback")
	}
	return state.sendCommand(map[string]interface{}{"cmd": "forward", "config": config})
}

// ProxyStatus requests a status snapshot from the shim for proxy id.
// It subscribes for the "status" event before sending the command so the
// background goroutine's fan-out delivers the response correctly.
func ProxyStatus(id int) (*ProxyState, error) {
	state, ok := GetProxy(id)
	if !ok {
		return nil, fmt.Errorf("no proxy with id %d", id)
	}
	ch := state.subscribe("status")
	if err := state.sendCommand(map[string]interface{}{"cmd": "status"}); err != nil {
		return state, err
	}
	select {
	case ev := <-ch:
		applyEvent(state, ev)
	case <-time.After(5 * time.Second):
		return state, fmt.Errorf("timed out waiting for status from proxy %d", id)
	}
	return state, nil
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

// applyEvent updates ProxyState fields from an incoming shim event.
// Forwarder state is maintained incrementally via socks_started/forward_started
// and socks_stopped/forward_stopped; the status event is not used for forwarders.
func applyEvent(state *ProxyState, ev map[string]interface{}) {
	state.mu.Lock()
	defer state.mu.Unlock()

	switch getString(ev, "event") {
	case "messenger_connected":
		state.MessengerID = getString(ev, "messenger_id")
	case "socks_started":
		port := 0
		if p, ok := ev["port"].(float64); ok {
			port = int(p)
		}
		state.Forwarders = append(state.Forwarders, ProxyForwarder{
			ID:            allocFwdID(),
			Type:          "socks",
			ListeningHost: getString(ev, "host"),
			ListeningPort: strconv.Itoa(port),
			port:          port,
		})
	case "forward_started":
		cfg := getString(ev, "config")
		parts := strings.SplitN(cfg, ":", 4)
		fwd := ProxyForwarder{
			ID:     allocFwdID(),
			Type:   "forward",
			config: cfg,
		}
		if len(parts) == 4 {
			fwd.ListeningHost = parts[0]
			fwd.ListeningPort = parts[1]
			fwd.DestinationHost = parts[2]
			fwd.DestinationPort = parts[3]
		}
		state.Forwarders = append(state.Forwarders, fwd)
	case "socks_stopped":
		port := 0
		if p, ok := ev["port"].(float64); ok {
			port = int(p)
		}
		updated := state.Forwarders[:0]
		for _, f := range state.Forwarders {
			if !(f.Type == "socks" && f.port == port) {
				updated = append(updated, f)
			}
		}
		state.Forwarders = updated
	case "forward_stopped":
		cfg := getString(ev, "config")
		updated := state.Forwarders[:0]
		for _, f := range state.Forwarders {
			if !(f.Type == "forward" && f.config == cfg) {
				updated = append(updated, f)
			}
		}
		state.Forwarders = updated
	}
}

// ProxyKillForwarder finds the forwarder with the given global ID across all
// running proxies, sends the appropriate stop command to the shim, and waits
// for the shim to confirm before returning.
func ProxyKillForwarder(fwdID int) error {
	proxyMapMu.Lock()
	var owner *ProxyState
	var fwd ProxyForwarder
	for _, s := range proxyMap {
		s.mu.Lock()
		for _, f := range s.Forwarders {
			if f.ID == fwdID {
				owner = s
				fwd = f
				break
			}
		}
		s.mu.Unlock()
		if owner != nil {
			break
		}
	}
	proxyMapMu.Unlock()

	if owner == nil {
		return fmt.Errorf("no forwarder with id %d", fwdID)
	}

	var shimCmd map[string]interface{}
	var wantEvent string
	if fwd.Type == "socks" {
		shimCmd = map[string]interface{}{"cmd": "stop_socks", "port": fwd.port}
		wantEvent = "socks_stopped"
	} else {
		shimCmd = map[string]interface{}{"cmd": "stop_forward", "config": fwd.config}
		wantEvent = "forward_stopped"
	}

	ch := owner.subscribe(wantEvent)
	if err := owner.sendCommand(shimCmd); err != nil {
		return err
	}
	select {
	case <-ch:
	case <-time.After(5 * time.Second):
		return fmt.Errorf("timed out waiting for forwarder %d to stop", fwdID)
	}
	return nil
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
