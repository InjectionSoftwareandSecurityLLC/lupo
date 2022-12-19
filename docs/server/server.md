# Server Usage

This document will describe the various commands and sub commands available in that are unique to the Lupo server. All commands here are only available on the server which means they either have not or will not be implemented in the client. The documentation will reflect this functionality as consistently as possible.

Please note, this document will not be comprehensive as CLI is "self documenting", meaning that you can get detailed usage information about any command within Lupo by simple running `help <command>` or `help <command> <subcommand>` etc...

For general usage on both client and server see the the [General Usage (Server/Client)
](./general_cli.md) documentation. For usage specific to the client, see the [Client](../client/client.md) documentation.


## Server Flags
The Lupo Server client is a CLI program that serves as a front end CLI and back end C2 server to run all of Lupo C2's functionalities. Currently the client can take a flag for a resource file to automate lupo commands.
- `-r` (Flag): this is the "resource file" flag. passing this flag the name of a saved resource file containing lupo commands. this will load in the file and execute each lupo command in sequence for automation.

## Wolfpack
The Wolfpack server is a team server used to allow multi-operator collaboration when interacting with the C2. The Wolfpack server can only be managed by the core Lupo "server" user. It is used to generate/manage operators/configurations to allow connections to a given Wolfpack instance. The actual `wolfpack` sub command can only be managed by the server preventing clients from arbitrarily adding/removing operators.
- wolfpack: base command for managing, starting, and stopping Wolfpack operators/wolfpack team server instances.
    - (sub command) deregister: removes an operator from the wolfpack server based on their username.
    - (sub command) register: add an operator to the wolfpack server. generates a configuration with their PSK and connection information so that the client binary can seamlessly authenticate them. requires a TLS cert to be generated for the Wolfpack instance.
    - (sub command) show: shows any active running Wolfpack server instance. currently only one instance at a time is supported, starting another instance will replace the old instance without properly shutting it down.
    - (sub command) start: starts a Wolfpack server instance. requires a TLS cert to be generated as all Wolfpack server communications are expected to be over HTTPS.
    - (sub command) stop: stops an actively running Wolfpack server instance.


