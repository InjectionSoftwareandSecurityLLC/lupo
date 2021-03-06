# Client Usage

This document will describe the various commands and sub commands available in that are unique to the Lupo client. All commands here are only available on the client which means they either have not or will not be implemented in the server. The documentation will reflect this functionality as consistently as possible.

Please note, this document will not be comprehensive as CLI is "self documenting", meaning that you can get detailed usage information about any command within Lupo by simple running `help <command>` or `help <command> <subcommand>` etc...

For general usage on both client and server see the the[General Usage (Server/Client)
](../server/general_cli.md) documentation. For usage specific to the client, see the [Server](../server/server.md) documentation.


## Client Flags
The Wolfpack client is a CLI program that serves as a front end to interface with a Wolfpack server instance. These clients allow for multi-operator use of the Wolfpack server and currently implements all commands of the server except for the `wolfpack` command to prevent non-privileged users from arbitrarily changing Wolfpack configurations without having access to the Lupo server's host. Currently the client only takes one additional configuration flag.
- `-c` (Flag): this is the "configuration" flag. passing this flag the name of a client configuration JSON file will tell it to use that configuration for authentication. by default it looks for the file `wolfpack.json` in the same directory as the binary. **NOTE** a known bug exists that prevents the flags from parsing full directories, so it's best to put the JSON file in the same directory as you execute the client binary from, but this file can have any name the user desires.


