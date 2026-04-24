# Building

To build Lupo C2 server or client simply insure the latest version of golang is installed for your system.

From there use the provided `Makefile` to generate client and server binaries. These binaries will be written to a `builds` directory. By default `make all` compiles binaries for *almost* every supported platform and architecture, including Windows, Linux, macOS (ARM & x64), ARMv7, ARM64, MIPS variants, ppc64, s390x, riscv64, loong64, FreeBSD, OpenBSD, NetBSD, Solaris, and Dragonfly BSD. Pre-built binaries for all targets are also available in the [releases](https://github.com/InjectionSoftwareandSecurityLLC/lupo/releases).

**Officially supported platforms** (tested and maintained): Windows x64, macOS (ARM64/Apple Silicon and x64/Intel), and Linux x64. All other Go-compatible targets are compiled and provided as-is — they should work but are not actively tested.

> **Note:** Several Go-supported targets are explicitly excluded due to the `github.com/desertbit/readline` dependency not implementing the required terminal syscalls for those platforms. Excluded targets: **Plan 9**, **AIX**, **Illumos**, **WASM** (both wasip1 and js), and **Android**.

To generate platform specific binaries simply specify the binary type, and the platform in your make command.

Server build example:
1. `make LUPO_SERVER-linux`

Client build example:
1. `make LUPO_CLIENT-linux`

You can also use the shorthand group targets to build both server and client for a platform at once (e.g. `make linux`, `make windows`).

Alternatively you may make use of the releases at:
https://github.com/InjectionSoftwareandSecurityLLC/lupo/releases

---

## Available Build Targets

Each target below can be called individually as `make <target>` to build both the server and client for that platform. For server or client only, use `make LUPO_SERVER-<platform>` or `make LUPO_CLIENT-<platform>`.

### Group Targets

Each OS has an umbrella target that builds all supported architectures for that OS, plus individual arch subtargets.

**Windows** (`all: x64, 386, arm64`)

| Target | Description |
|---|---|
| `windows` | All Windows architectures |
| `windows-x64` | Windows x64 |
| `windows-386` | Windows 386 |
| `windows-arm64` | Windows ARM64 |

**Linux** (`all: x64, 386, arm, arm64, mips, mipsle, mips64, mips64le, ppc64, ppc64le, s390x, riscv64, loong64`)

| Target | Description |
|---|---|
| `linux` | All Linux architectures |
| `linux-x64` | Linux x64 |
| `linux-386` | Linux 386 |
| `linux-arm` | Linux ARMv7 (GOARM=7, VFPv3 — Cortex-A) |
| `linux-arm6` | Linux ARMv6 (GOARM=6, VFPv1 — ARM11 / Raspberry Pi 1) |
| `linux-arm5` | Linux ARMv5 (GOARM=5, software float — old/embedded ARM) |
| `linux-arm64` | Linux ARM64 |
| `linux-mips` | Linux MIPS |
| `linux-mipsle` | Linux MIPS little-endian |
| `linux-mips64` | Linux MIPS64 |
| `linux-mips64le` | Linux MIPS64 little-endian |
| `linux-ppc64` | Linux ppc64 |
| `linux-ppc64le` | Linux ppc64 little-endian |
| `linux-s390x` | Linux s390x |
| `linux-riscv64` | Linux RISC-V 64 |
| `linux-loong64` | Linux LoongArch 64 |

**macOS/Darwin** (`all: arm64, x64`)

| Target | Description |
|---|---|
| `darwin` | All macOS architectures |
| `darwin-arm64` | macOS ARM64 (Apple Silicon) |
| `darwin-x64` | macOS x64 (Intel) |

**FreeBSD (`all: x64, 386, arm, arm64`)

| Target | Description |
|---|---|
| `freebsd` | All FreeBSD architectures |
| `freebsd-x64` | FreeBSD x64 |
| `freebsd-386` | FreeBSD 386 |
| `freebsd-arm` | FreeBSD ARM |
| `freebsd-arm64` | FreeBSD ARM64 |

**OpenBSD** (`all: x64, 386, arm, arm64`)

| Target | Description |
|---|---|
| `openbsd` | All OpenBSD architectures |
| `openbsd-x64` | OpenBSD x64 |
| `openbsd-386` | OpenBSD 386 |
| `openbsd-arm` | OpenBSD ARM |
| `openbsd-arm64` | OpenBSD ARM64 |

**NetBSD** (`all: x64, 386, arm, arm64`)

| Target | Description |
|---|---|
| `netbsd` | All NetBSD architectures |
| `netbsd-x64` | NetBSD x64 |
| `netbsd-386` | NetBSD 386 |
| `netbsd-arm` | NetBSD ARM |
| `netbsd-arm64` | NetBSD ARM64 |

**Android** (`all: x64, 386, arm, arm64`)

| Target | Description |
|---|---|
| `android` | All Android architectures |
| `android-x64` | Android x64 |
| `android-386` | Android 386 |
| `android-arm` | Android ARM |
| `android-arm64` | Android ARM64 |

**Single-arch targets**

| Target | Description |
|---|---|
| `solaris` | Solaris x64 |
| `dragonfly` | Dragonfly BSD x64 |

### Individual Binary Targets

Each group target above resolves to individual `LUPO_SERVER-<platform>` and `LUPO_CLIENT-<platform>` targets. The full list of individual platform suffixes:

`windows`, `windows-x64`, `windows-386`, `windows-arm64`, `linux`, `linux-x64`, `linux-386`, `linux-arm`, `linux-arm6`, `linux-arm5`, `linux-arm64`, `linux-mips`, `linux-mipsle`, `linux-mips64`, `linux-mips64le`, `linux-ppc64`, `linux-ppc64le`, `linux-s390x`, `linux-riscv64`, `linux-loong64`, `darwin`, `darwin-arm64`, `darwin-x64`, `freebsd`, `freebsd-x64`, `freebsd-386`, `freebsd-arm`, `freebsd-arm64`, `openbsd`, `openbsd-x64`, `openbsd-386`, `openbsd-arm`, `openbsd-arm64`, `netbsd`, `netbsd-x64`, `netbsd-386`, `netbsd-arm`, `netbsd-arm64`, `solaris`, `dragonfly`


Encryption Note:
- TLS certs are dynamically generated each run if one does not exist already. These certs are stored in the `tls-certs` directory and are used by the default configurations for HTTPS and Wolfpack listeners. All services that require TLS also provide optional arguments to specify your own custom cert per service if you'd prefer.
