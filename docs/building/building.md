# Building

To build Lupo C2 server or client simply insure the latest version of golang is installed for your system.

From there use the provided `Makefile` to generate client and server binaries. These binaries will be written to a `builds` directory. By default `make all` will build all cross platform binaries supported by Lupo for Windows, MacOS, and Linux. Technically ARM, MIPS, and any other golang platform is also supported, with options for ARM and MIPS included in the Makefile, however only Windows, MacOS, and Linux are currently officially supported.

To generate platform specific binaries simply specify the binary type, and the platform in your make command.

Server build example:
1. `make LUPO_SERVER-linux`

Client build example:
1. `make LUPO_CLIENT-linux`

Alternatively you may make use of the releases at:
https://github.com/InjectionSoftwareandSecurityLLC/lupo/releases