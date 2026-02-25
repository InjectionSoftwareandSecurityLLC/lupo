# General Usage (Server/Client)

This document will describe the various commands and sub commands available in Lupo. All commands here are universal across the server and client. This means that, outside of a few custom error message use cases, you can operate Lupo with any of these commands regardless of whether you operate it from the server CLI or from a client CLI in conjunction with the Wolfpack Server.

Please note, this document will not be comprehensive as CLI is "self documenting", meaning that you can get detailed usage information about any command within Lupo by simple running `help <command>` or `help <command> <subcommand>` etc...

Instead this will show case the most common commands and how to use them then split [Server](./server.md) and [Client](../client/client.md) specific commands to their own documentation respectively.

On start lupo's server can take two optional arguments:
1. `--tls-ip` which is the IP address you'd like to associate with Lupo's default self-signed TLS cert. By default this will attempt to assign it to a non-loopback interface automatically which is usually the correct one if you're only using a single primary network interface. If not an you'd like to assign a different one - that's what this flag is for! If an interface can not be identified it will fallback to loopback on `127.0.0.1`.
2. `-r` which is for specifying a lupo resource file. This is similar to the Metasploit framework's `.rc` files in that they will play back lupo commands line by line to give you the ability to automate configuration of callbacks, wolfpack services, and other lupo server settings.

## Lupo Commands
- exec: executes a local system command so the user can interact with their local session without leaving the Lupo CLI.
- connector: base command for establishing bind based connections over HTTP/HTTPS. Good for bind based implants or web shells.
    - (sub command) start: takes several flags to start a bind connection over HTTP/HTTPS and establish a permanent session that is managed by the interact/session commands. Starting a basic connection to a PHP based web shell might work like this:
        - Start connector: `connector start -d shell.php -r example.com -p 8080 -x HTTP` (example shown running on localhost):
        ![connector gif](../assets/connector.gif)
        - A session should open allowing execution to "http://example.com:8080/shell.php?cmd=(CMD)" with GET based queries via HTTP. The placeholder (CMD) will be populated when executing the `cmd` sub command within the session shell.
- interact: base command for interacting with sessions. using the interact command and passing it a valid session ID will drop the user into a `session` cli where they can begin interaction with that session.
    - (sub command) clean: cleans all sessions from Lupo where status is `DEAD`. this will not stop the session from trying to reconnect if it ever comes back online, but depending on how the implant was implemented, it may not be able to re-stablish a valid session (a re-authentication routine would need to be added by that implant's developer). it does however, remove the session from the active sessions preventing the same authentication identifiers from being reused, and of course removes the ability to interact with it.
    - (sub command) kill: removes a specific session. works exactly like `clean` does but only on one specific session regardless if the session is `ALIVE` or `DEAD`.
    - (sub command) show: shows all registered sessions and relevant information such as hosts and statuses.
- listener: base command for managing listeners. current supported listeners are HTTP, HTTPS, and TCP.
    - (sub command) kill: removes a specified listener by ID and kills the corresponding server routine running it.
    - (sub command) manage: manage global listener attributes such as the PSK (random by default).
    - (sub command) show: shows all running listeners and their meta information.
    - (sub command) cmd: sends/posts a command that will be collected or received by a given connection/implant and executed as a system command - supports running a single command across multiple sessions with filtering options.
    - (sub command) start: configure and start a new listener. A sample command for a basic HTTPS server might look like this:
        - Start listener: `listener start -l 0.0.0.0 -p 8443`
        - That command will start an HTTPS listener using the default TLS keys stored in the runtime generate `tls-certs` folder. This keypair is also useful as the default for cert pinning on implants (although use of a redirector is the reccommended set up). All of these options can of course be modified using the relevant flags/arguments available in the CLI if you'd like to set up unique certs per listener.
        - Sample showing setting a custom PSK, starting a listener, showing active listeners, and killing a listener:
        ![listener gif](../assets/listener.gif)
- stager: base command for managing file stagers. stagers are simple HTTP/HTTPS static file servers used to serve payloads and other files to targets.
    - (sub command) start: configure and start a new file stager. A sample command for a basic HTTP stager might look like this:
        - Start stager: `stager start -l 0.0.0.0 -p 8080 -d /opt/payloads`
        - That command will start an HTTP stager serving files from the `/opt/payloads` directory (which will be created if it does not exist). For HTTPS, supply `-x HTTPS` along with `-k <key path>` and `-c <cert path>` to specify your TLS key and certificate. The default TLS key and cert paths point to the same runtime-generated keypair used by listeners.
        - Available flags: `-l`/`--lhost` (default `127.0.0.1`), `-p`/`--lport` (default `8080`), `-x`/`--protocol` (`HTTP` or `HTTPS`, default `HTTP`), `-d`/`--dir` (directory to serve, default `stager`), `-k`/`--key` (TLS key path, HTTPS only), `-c`/`--cert` (TLS cert path, HTTPS only).
    - (sub command) show: shows all running stagers and their configuration (ID, host, port, protocol, and directory).
    - (sub command) kill: removes a specific stager by ID and shuts down the corresponding file server.
- session: base command, a sub shell that allows interaction with a session specified from the `interact` command.
    - (sub command) back: returns to the core Lupo CLI shell.
    - (sub command) cmd: sends/posts a command that will be collected or received by a given connection/implant and executed as a system command.
    - (sub command) download: downloads a file from the target session _if_ they have a download handler implemented.
    - (sub command) upload: uploads a file to the target session _if_ they have a upload handler implemented.
    - (sub command) updateinterval: updates the implant's check-in delay _if_ they have the updateinterval handler implemented.
    - (sub command) kill: kills a specified session, requires an argument so you don't accidentally kill your current interacted with session. this works the same way as `interact kill` does.
    - (sub command) load: loads extended functions if they are available to a given session/implant
    - (sub command) session: swaps between sessions within the session sub shell so the user doesn't need to go back and use `interact` to switch sessions.
    - (sub command) upload: uploads a file to the target session _if_ they have an upload handler implemented.
    - (sub command) mem_inject: provides an interface to upload shellcode to be injected into memory _if_ the target session has a memory injection handler implemented.
    - (sub command) pid_inject: provides an interface to upload shellcode to be injected into a specific process identifier _if_ the target session has a process injection handler implemented.
    - (sub command) bof_loader: delivers a BOF/COFF object file payload to the target session for execution _if_ the target session has a BOF loader handler implemented. Accepts a positional `payload` argument (local path to the COFF file) and an optional `-a`/`--arguments` flag for passing typed arguments to the BOF entry point. Arguments use a type-prefix syntax: `wstring:<value>` (UTF-16LE wide string, default), `string:<value>` (ASCII string), `int:<value>` (32-bit integer), or `short:<value>` (16-bit integer). Multiple arguments are space-separated within the flag value. Execution is synchronous — output is returned on the implant's next check-in.
    - (sub command) bof_loader_async: identical to `bof_loader` but executes the BOF in a background goroutine so the implant continues checking in without blocking. Accepts the same `payload` argument and `-a`/`--arguments` flag with the same type-prefix syntax. Results are queued on the implant and returned to all connected operators on the next check-in after the BOF completes. Requires the target session to have an async BOF loader handler implemented.
    - (sub command) pe_loader: delivers a PE (portable executable) payload to the target session for in-memory execution _if_ the target session has a PE loader handler implemented. Accepts a positional `payload` argument (local path to the PE file) and an optional `-a`/`--arguments` flag for passing arguments to the PE's entry point.


## Logging
Lupo implements some sophisticated yet incredibly simple logging mechanisms. Lupo was designed with "Red Teaming" in mind. Operators should live, breathe, and die by their logs to replicate events, know where they messed up, but most importantly to verify and keep track of what they've done. 
- History: Lupo uses Grumble CLI's built in history file to keep a local history of all commands executed by an operator. This works like a bash history and is more for convenience than it is for logging, but is helpful nonetheless. Since this is local the server and client binaries' logs are local to their respective systems so retrieving a client's history would require having that physical machine. The history file is stored in `.lupo.history` by default.
- Operator Logs: Lupo implements a custom logger that is implemented throughout every relevant traceable event within Lupo and is easily added to other events as time passes. These are generated on the server and stored on the server meaning only those with access to the Lupo C2 server can read the logs. These are stored with a format of `YYYY/MM/DD 00:00:00 (24 Hour Format) <message>`. The log file is stored in `.lupo.log` by default. This log tracks everything from executed commands, configuration changes, responses between server/implants including files, command output, and check in information. Log sample:
![log sample png](../assets/log_sample.png)
- Chat Log: Lupo uses the same custom logger implemented by the Operator Logs function to log a chat history. The history is only stored on the Lupo server and can only be written to and accessed by the Lupo server. When a user enters the chat the log is read and sent to the chat instance being interacted with. When new messages are sent the message is written to the chat and the most recent line is then broadcasts to every Wolfpack user that is also in the chat CLI on their Lupo client. These are stored with a format of `YYYY/MM/DD 00:00:00 (24 Hour Format) <Username>: <message>`.

## Helpful Tools
Included with the Lupo C2 repo there are a handful of sample programs to help folks get started.

- The [samples](../../samples) directory in the repository includes sample payloads that can be used with Lupo. These are meant to be feature complete examples that fully implement each API. These are purposefully not distributed as binaries to promote longevity, but the whole point of Lupo is to "bring your own implant", so feel free to use the provided samples as a base, but it is _not_ recommended to use them in real world engagements without prior obfuscation/modification. Remember, they are samples, and while they are feature complete, they are meant to serve as examples only. Read more in the [implants](../implants/implants.md) documentation. Use at your own risk. Currently implemented samples:
    - HTTP/HTTPS Golang Based Implant
    - Bind Connector PHP Based Web Shell
