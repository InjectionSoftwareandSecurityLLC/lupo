# Building

To build Lupo C2 server or client simply insure the latest version of golang is installed for your system.

From there use the provided `Makefile` to generate client and server binaries. These binaries will be written to a `builds` directory. By default `make all` will build cross platform binaries for Windows, Linux, MacOS(ARM & x64), ARMv7, and MIPS. Platforms like ARMv7, MIPS, and any other golang compatible platform are all technically supported, but only Windows, MacOS, and Linux are currently **officially** supported by Lupo.

To generate platform specific binaries simply specify the binary type, and the platform in your make command.

Server build example:
1. `make LUPO_SERVER-linux`

Client build example:
1. `make LUPO_CLIENT-linux`

Alternatively you may make use of the releases at:
https://github.com/InjectionSoftwareandSecurityLLC/lupo/releases


Encryption Note:
- TLS certs are dynamically generated each run if one does not exist already. These certs are stored in the `tls-certs` directory and are used by the default configurations for HTTPS and Wolfpack listeners. All services that require TLS also provide optional arguments to specify your own custom cert per service if you'd prefer.
