# !!!MAKE SURE YOUR GOPATH ENVIRONMENT VARIABLE IS SET FIRST!!!

# Variables
DIR=builds
LUPO_SERVER=LUPO_SERVER
LUPO_CLIENT=LUPO_CLIENT
W=Windows-x64
W386=Windows-386
WARM64=Windows-arm64
L=Linux-x64
L386=Linux-386
A=Linux-arm7
A5=Linux-arm5
A6=Linux-arm6
A64=Linux-arm64
M=Linux-mips
MLE=Linux-mipsle
M64=Linux-mips64
M64LE=Linux-mips64le
LPPC64=Linux-ppc64
LPPC64LE=Linux-ppc64le
LS390X=Linux-s390x
LRISCV64=Linux-riscv64
LLOONG64=Linux-loong64
D=Darwin-arm64
DX=Darwin-x64
FBSD64=FreeBSD-x64
FBSD386=FreeBSD-386
FBSDARM=FreeBSD-arm
FBSDARM64=FreeBSD-arm64
OBSD64=OpenBSD-x64
OBSD386=OpenBSD-386
OBSDARM=OpenBSD-arm
OBSDARM64=OpenBSD-arm64
NBSD64=NetBSD-x64
NBSD386=NetBSD-386
NBSDARM=NetBSD-arm
NBSDARM64=NetBSD-arm64
SOL=Solaris-x64
DRAG=Dragonfly-x64

# Make Directory to store executables
$(shell mkdir -p ${DIR})

# Install go dependencies
$(shell go get ./lupo-server)
$(shell go get ./lupo-client)
$(shell go get ./sample)


# Change default to just make for the host OS and add MAKE ALL to do this
default: LUPO_SERVER-windows LUPO_SERVER-windows-386 LUPO_SERVER-windows-arm64 LUPO_SERVER-linux LUPO_SERVER-linux-386 LUPO_SERVER-darwin LUPO_SERVER-darwin-x64 LUPO_SERVER-arm LUPO_SERVER-arm5 LUPO_SERVER-arm6 LUPO_SERVER-arm64 LUPO_SERVER-mips LUPO_SERVER-mipsle LUPO_SERVER-mips64 LUPO_SERVER-mips64le LUPO_SERVER-linux-ppc64 LUPO_SERVER-linux-ppc64le LUPO_SERVER-linux-s390x LUPO_SERVER-linux-riscv64 LUPO_SERVER-linux-loong64 LUPO_SERVER-freebsd LUPO_SERVER-freebsd-386 LUPO_SERVER-freebsd-arm LUPO_SERVER-freebsd-arm64 LUPO_SERVER-openbsd LUPO_SERVER-openbsd-386 LUPO_SERVER-openbsd-arm LUPO_SERVER-openbsd-arm64 LUPO_SERVER-netbsd LUPO_SERVER-netbsd-386 LUPO_SERVER-netbsd-arm LUPO_SERVER-netbsd-arm64 LUPO_SERVER-solaris LUPO_SERVER-dragonfly LUPO_CLIENT-windows LUPO_CLIENT-windows-386 LUPO_CLIENT-windows-arm64 LUPO_CLIENT-linux LUPO_CLIENT-linux-386 LUPO_CLIENT-darwin LUPO_CLIENT-darwin-x64 LUPO_CLIENT-arm LUPO_CLIENT-arm5 LUPO_CLIENT-arm6 LUPO_CLIENT-arm64 LUPO_CLIENT-mips LUPO_CLIENT-mipsle LUPO_CLIENT-mips64 LUPO_CLIENT-mips64le LUPO_CLIENT-linux-ppc64 LUPO_CLIENT-linux-ppc64le LUPO_CLIENT-linux-s390x LUPO_CLIENT-linux-riscv64 LUPO_CLIENT-linux-loong64 LUPO_CLIENT-freebsd LUPO_CLIENT-freebsd-386 LUPO_CLIENT-freebsd-arm LUPO_CLIENT-freebsd-arm64 LUPO_CLIENT-openbsd LUPO_CLIENT-openbsd-386 LUPO_CLIENT-openbsd-arm LUPO_CLIENT-openbsd-arm64 LUPO_CLIENT-netbsd LUPO_CLIENT-netbsd-386 LUPO_CLIENT-netbsd-arm LUPO_CLIENT-netbsd-arm64 LUPO_CLIENT-solaris LUPO_CLIENT-dragonfly

all: default

# Compile Windows binaries (all architectures)
windows: LUPO_SERVER-windows LUPO_SERVER-windows-386 LUPO_SERVER-windows-arm64 LUPO_CLIENT-windows LUPO_CLIENT-windows-386 LUPO_CLIENT-windows-arm64
windows-x64: LUPO_SERVER-windows LUPO_CLIENT-windows
windows-386: LUPO_SERVER-windows-386 LUPO_CLIENT-windows-386
windows-arm64: LUPO_SERVER-windows-arm64 LUPO_CLIENT-windows-arm64

# Compile Linux binaries (all architectures)
linux: LUPO_SERVER-linux LUPO_SERVER-linux-386 LUPO_SERVER-arm LUPO_SERVER-arm5 LUPO_SERVER-arm6 LUPO_SERVER-arm64 LUPO_SERVER-mips LUPO_SERVER-mipsle LUPO_SERVER-mips64 LUPO_SERVER-mips64le LUPO_SERVER-linux-ppc64 LUPO_SERVER-linux-ppc64le LUPO_SERVER-linux-s390x LUPO_SERVER-linux-riscv64 LUPO_SERVER-linux-loong64 LUPO_CLIENT-linux LUPO_CLIENT-linux-386 LUPO_CLIENT-arm LUPO_CLIENT-arm5 LUPO_CLIENT-arm6 LUPO_CLIENT-arm64 LUPO_CLIENT-mips LUPO_CLIENT-mipsle LUPO_CLIENT-mips64 LUPO_CLIENT-mips64le LUPO_CLIENT-linux-ppc64 LUPO_CLIENT-linux-ppc64le LUPO_CLIENT-linux-s390x LUPO_CLIENT-linux-riscv64 LUPO_CLIENT-linux-loong64
linux-x64: LUPO_SERVER-linux LUPO_CLIENT-linux
linux-386: LUPO_SERVER-linux-386 LUPO_CLIENT-linux-386
linux-arm: LUPO_SERVER-arm LUPO_CLIENT-arm
linux-arm5: LUPO_SERVER-arm5 LUPO_CLIENT-arm5
linux-arm6: LUPO_SERVER-arm6 LUPO_CLIENT-arm6
linux-arm64: LUPO_SERVER-arm64 LUPO_CLIENT-arm64
linux-mips: LUPO_SERVER-mips LUPO_CLIENT-mips
linux-mipsle: LUPO_SERVER-mipsle LUPO_CLIENT-mipsle
linux-mips64: LUPO_SERVER-mips64 LUPO_CLIENT-mips64
linux-mips64le: LUPO_SERVER-mips64le LUPO_CLIENT-mips64le
linux-ppc64: LUPO_SERVER-linux-ppc64 LUPO_CLIENT-linux-ppc64
linux-ppc64le: LUPO_SERVER-linux-ppc64le LUPO_CLIENT-linux-ppc64le
linux-s390x: LUPO_SERVER-linux-s390x LUPO_CLIENT-linux-s390x
linux-riscv64: LUPO_SERVER-linux-riscv64 LUPO_CLIENT-linux-riscv64
linux-loong64: LUPO_SERVER-linux-loong64 LUPO_CLIENT-linux-loong64

# Compile Darwin (macOS) binaries (all architectures)
darwin: LUPO_SERVER-darwin LUPO_SERVER-darwin-x64 LUPO_CLIENT-darwin LUPO_CLIENT-darwin-x64
darwin-arm64: LUPO_SERVER-darwin LUPO_CLIENT-darwin
darwin-x64: LUPO_SERVER-darwin-x64 LUPO_CLIENT-darwin-x64

# Compile FreeBSD binaries (all architectures)
freebsd: LUPO_SERVER-freebsd LUPO_SERVER-freebsd-386 LUPO_SERVER-freebsd-arm LUPO_SERVER-freebsd-arm64 LUPO_CLIENT-freebsd LUPO_CLIENT-freebsd-386 LUPO_CLIENT-freebsd-arm LUPO_CLIENT-freebsd-arm64
freebsd-x64: LUPO_SERVER-freebsd LUPO_CLIENT-freebsd
freebsd-386: LUPO_SERVER-freebsd-386 LUPO_CLIENT-freebsd-386
freebsd-arm: LUPO_SERVER-freebsd-arm LUPO_CLIENT-freebsd-arm
freebsd-arm64: LUPO_SERVER-freebsd-arm64 LUPO_CLIENT-freebsd-arm64

# Compile OpenBSD binaries (all architectures)
openbsd: LUPO_SERVER-openbsd LUPO_SERVER-openbsd-386 LUPO_SERVER-openbsd-arm LUPO_SERVER-openbsd-arm64 LUPO_CLIENT-openbsd LUPO_CLIENT-openbsd-386 LUPO_CLIENT-openbsd-arm LUPO_CLIENT-openbsd-arm64
openbsd-x64: LUPO_SERVER-openbsd LUPO_CLIENT-openbsd
openbsd-386: LUPO_SERVER-openbsd-386 LUPO_CLIENT-openbsd-386
openbsd-arm: LUPO_SERVER-openbsd-arm LUPO_CLIENT-openbsd-arm
openbsd-arm64: LUPO_SERVER-openbsd-arm64 LUPO_CLIENT-openbsd-arm64

# Compile NetBSD binaries (all architectures)
netbsd: LUPO_SERVER-netbsd LUPO_SERVER-netbsd-386 LUPO_SERVER-netbsd-arm LUPO_SERVER-netbsd-arm64 LUPO_CLIENT-netbsd LUPO_CLIENT-netbsd-386 LUPO_CLIENT-netbsd-arm LUPO_CLIENT-netbsd-arm64
netbsd-x64: LUPO_SERVER-netbsd LUPO_CLIENT-netbsd
netbsd-386: LUPO_SERVER-netbsd-386 LUPO_CLIENT-netbsd-386
netbsd-arm: LUPO_SERVER-netbsd-arm LUPO_CLIENT-netbsd-arm
netbsd-arm64: LUPO_SERVER-netbsd-arm64 LUPO_CLIENT-netbsd-arm64

# Compile Solaris x64 binaries
solaris: LUPO_SERVER-solaris LUPO_CLIENT-solaris

# Compile Dragonfly x64 binaries
dragonfly: LUPO_SERVER-dragonfly LUPO_CLIENT-dragonfly

# Compile LUPO_SERVER - Windows x64
LUPO_SERVER-windows:
	export GOOS=windows GOARCH=amd64;go build -o ${DIR}/${LUPO_SERVER}-${W}.exe lupo-server/main.go

# Compile LUPO_SERVER - Linux x64
LUPO_SERVER-linux:
	export GOOS=linux;export GOARCH=amd64;go build -o ${DIR}/${LUPO_SERVER}-${L} lupo-server/main.go

# Compile LUPO_SERVER - Darwin arm64 (Apple Silicon)
LUPO_SERVER-darwin:
	export GOOS=darwin;export GOARCH=arm64;go build -o ${DIR}/${LUPO_SERVER}-${D} lupo-server/main.go

# Compile LUPO_SERVER - Darwin x64 (Intel macOS)
LUPO_SERVER-darwin-x64:
	export GOOS=darwin;export GOARCH=amd64;go build -o ${DIR}/${LUPO_SERVER}-${DX} lupo-server/main.go

# Compile LUPO_SERVER - Linux mips
LUPO_SERVER-mips:
	export CGO_ENABLED=0;export GOOS=linux;export GOARCH=mips;go build -o ${DIR}/${LUPO_SERVER}-${M} lupo-server/main.go

# Compile LUPO_SERVER - Linux arm (ARMv7 / GOARM=7)
LUPO_SERVER-arm:
	export CGO_ENABLED=0;export GOOS=linux;export GOARCH=arm;export GOARM=7;go build -o ${DIR}/${LUPO_SERVER}-${A} lupo-server/main.go

# Compile LUPO_SERVER - Linux arm5 (ARMv5 / GOARM=5, software float)
LUPO_SERVER-arm5:
	export CGO_ENABLED=0;export GOOS=linux;export GOARCH=arm;export GOARM=5;go build -o ${DIR}/${LUPO_SERVER}-${A5} lupo-server/main.go

# Compile LUPO_SERVER - Linux arm6 (ARMv6 / GOARM=6, VFPv1)
LUPO_SERVER-arm6:
	export CGO_ENABLED=0;export GOOS=linux;export GOARCH=arm;export GOARM=6;go build -o ${DIR}/${LUPO_SERVER}-${A6} lupo-server/main.go

# Compile LUPO_SERVER - Linux arm64
LUPO_SERVER-arm64:
	export CGO_ENABLED=0;export GOOS=linux;export GOARCH=arm64;go build -o ${DIR}/${LUPO_SERVER}-${A64} lupo-server/main.go

# Compile LUPO_SERVER - Linux mips64
LUPO_SERVER-mips64:
	export CGO_ENABLED=0;export GOOS=linux;export GOARCH=mips64;go build -o ${DIR}/${LUPO_SERVER}-${M64} lupo-server/main.go

# Compile LUPO_SERVER - Linux mips64le
LUPO_SERVER-mips64le:
	export CGO_ENABLED=0;export GOOS=linux;export GOARCH=mips64le;go build -o ${DIR}/${LUPO_SERVER}-${M64LE} lupo-server/main.go

# Compile LUPO_SERVER - Linux mipsle
LUPO_SERVER-mipsle:
	export CGO_ENABLED=0;export GOOS=linux;export GOARCH=mipsle;go build -o ${DIR}/${LUPO_SERVER}-${MLE} lupo-server/main.go

# Compile LUPO_SERVER - Linux 386
LUPO_SERVER-linux-386:
	export CGO_ENABLED=0;export GOOS=linux;export GOARCH=386;go build -o ${DIR}/${LUPO_SERVER}-${L386} lupo-server/main.go

# Compile LUPO_SERVER - Linux ppc64
LUPO_SERVER-linux-ppc64:
	export CGO_ENABLED=0;export GOOS=linux;export GOARCH=ppc64;go build -o ${DIR}/${LUPO_SERVER}-${LPPC64} lupo-server/main.go

# Compile LUPO_SERVER - Linux ppc64le
LUPO_SERVER-linux-ppc64le:
	export CGO_ENABLED=0;export GOOS=linux;export GOARCH=ppc64le;go build -o ${DIR}/${LUPO_SERVER}-${LPPC64LE} lupo-server/main.go

# Compile LUPO_SERVER - Linux s390x
LUPO_SERVER-linux-s390x:
	export CGO_ENABLED=0;export GOOS=linux;export GOARCH=s390x;go build -o ${DIR}/${LUPO_SERVER}-${LS390X} lupo-server/main.go

# Compile LUPO_SERVER - Linux riscv64
LUPO_SERVER-linux-riscv64:
	export CGO_ENABLED=0;export GOOS=linux;export GOARCH=riscv64;go build -o ${DIR}/${LUPO_SERVER}-${LRISCV64} lupo-server/main.go

# Compile LUPO_SERVER - Linux loong64
LUPO_SERVER-linux-loong64:
	export CGO_ENABLED=0;export GOOS=linux;export GOARCH=loong64;go build -o ${DIR}/${LUPO_SERVER}-${LLOONG64} lupo-server/main.go

# Compile LUPO_SERVER - Windows 386
LUPO_SERVER-windows-386:
	export CGO_ENABLED=0;export GOOS=windows;export GOARCH=386;go build -o ${DIR}/${LUPO_SERVER}-${W386}.exe lupo-server/main.go

# Compile LUPO_SERVER - Windows arm64
LUPO_SERVER-windows-arm64:
	export CGO_ENABLED=0;export GOOS=windows;export GOARCH=arm64;go build -o ${DIR}/${LUPO_SERVER}-${WARM64}.exe lupo-server/main.go

# Compile LUPO_SERVER - FreeBSD x64
LUPO_SERVER-freebsd:
	export CGO_ENABLED=0;export GOOS=freebsd;export GOARCH=amd64;go build -o ${DIR}/${LUPO_SERVER}-${FBSD64} lupo-server/main.go

# Compile LUPO_SERVER - FreeBSD 386
LUPO_SERVER-freebsd-386:
	export CGO_ENABLED=0;export GOOS=freebsd;export GOARCH=386;go build -o ${DIR}/${LUPO_SERVER}-${FBSD386} lupo-server/main.go

# Compile LUPO_SERVER - FreeBSD arm
LUPO_SERVER-freebsd-arm:
	export CGO_ENABLED=0;export GOOS=freebsd;export GOARCH=arm;go build -o ${DIR}/${LUPO_SERVER}-${FBSDARM} lupo-server/main.go

# Compile LUPO_SERVER - FreeBSD arm64
LUPO_SERVER-freebsd-arm64:
	export CGO_ENABLED=0;export GOOS=freebsd;export GOARCH=arm64;go build -o ${DIR}/${LUPO_SERVER}-${FBSDARM64} lupo-server/main.go

# Compile LUPO_SERVER - OpenBSD x64
LUPO_SERVER-openbsd:
	export CGO_ENABLED=0;export GOOS=openbsd;export GOARCH=amd64;go build -o ${DIR}/${LUPO_SERVER}-${OBSD64} lupo-server/main.go

# Compile LUPO_SERVER - OpenBSD 386
LUPO_SERVER-openbsd-386:
	export CGO_ENABLED=0;export GOOS=openbsd;export GOARCH=386;go build -o ${DIR}/${LUPO_SERVER}-${OBSD386} lupo-server/main.go

# Compile LUPO_SERVER - OpenBSD arm
LUPO_SERVER-openbsd-arm:
	export CGO_ENABLED=0;export GOOS=openbsd;export GOARCH=arm;go build -o ${DIR}/${LUPO_SERVER}-${OBSDARM} lupo-server/main.go

# Compile LUPO_SERVER - OpenBSD arm64
LUPO_SERVER-openbsd-arm64:
	export CGO_ENABLED=0;export GOOS=openbsd;export GOARCH=arm64;go build -o ${DIR}/${LUPO_SERVER}-${OBSDARM64} lupo-server/main.go

# Compile LUPO_SERVER - NetBSD x64
LUPO_SERVER-netbsd:
	export CGO_ENABLED=0;export GOOS=netbsd;export GOARCH=amd64;go build -o ${DIR}/${LUPO_SERVER}-${NBSD64} lupo-server/main.go

# Compile LUPO_SERVER - NetBSD 386
LUPO_SERVER-netbsd-386:
	export CGO_ENABLED=0;export GOOS=netbsd;export GOARCH=386;go build -o ${DIR}/${LUPO_SERVER}-${NBSD386} lupo-server/main.go

# Compile LUPO_SERVER - NetBSD arm
LUPO_SERVER-netbsd-arm:
	export CGO_ENABLED=0;export GOOS=netbsd;export GOARCH=arm;go build -o ${DIR}/${LUPO_SERVER}-${NBSDARM} lupo-server/main.go

# Compile LUPO_SERVER - NetBSD arm64
LUPO_SERVER-netbsd-arm64:
	export CGO_ENABLED=0;export GOOS=netbsd;export GOARCH=arm64;go build -o ${DIR}/${LUPO_SERVER}-${NBSDARM64} lupo-server/main.go

# Compile LUPO_SERVER - Solaris x64
LUPO_SERVER-solaris:
	export CGO_ENABLED=0;export GOOS=solaris;export GOARCH=amd64;go build -o ${DIR}/${LUPO_SERVER}-${SOL} lupo-server/main.go

# Compile LUPO_SERVER - Dragonfly x64
LUPO_SERVER-dragonfly:
	export CGO_ENABLED=0;export GOOS=dragonfly;export GOARCH=amd64;go build -o ${DIR}/${LUPO_SERVER}-${DRAG} lupo-server/main.go

# Compile LUPO_CLIENT - Windows x64
LUPO_CLIENT-windows:
	export GOOS=windows GOARCH=amd64;go build -o ${DIR}/${LUPO_CLIENT}-${W}.exe lupo-client/main.go

# Compile LUPO_CLIENT - Linux x64
LUPO_CLIENT-linux:
	export GOOS=linux;export GOARCH=amd64;go build -o ${DIR}/${LUPO_CLIENT}-${L} lupo-client/main.go

# Compile LUPO_CLIENT - Darwin arm64 (Apple Silicon)
LUPO_CLIENT-darwin:
	export GOOS=darwin;export GOARCH=arm64;go build -o ${DIR}/${LUPO_CLIENT}-${D} lupo-client/main.go

# Compile LUPO_CLIENT - Darwin x64 (Intel macOS)
LUPO_CLIENT-darwin-x64:
	export GOOS=darwin;export GOARCH=amd64;go build -o ${DIR}/${LUPO_CLIENT}-${DX} lupo-client/main.go

# Compile LUPO_CLIENT - Linux mips
LUPO_CLIENT-mips:
	export CGO_ENABLED=0;export GOOS=linux;export GOARCH=mips;go build -o ${DIR}/${LUPO_CLIENT}-${M} lupo-client/main.go

# Compile LUPO_CLIENT - Linux arm (ARMv7 / GOARM=7)
LUPO_CLIENT-arm:
	export CGO_ENABLED=0;export GOOS=linux;export GOARCH=arm;export GOARM=7;go build -o ${DIR}/${LUPO_CLIENT}-${A} lupo-client/main.go

# Compile LUPO_CLIENT - Linux arm5 (ARMv5 / GOARM=5, software float)
LUPO_CLIENT-arm5:
	export CGO_ENABLED=0;export GOOS=linux;export GOARCH=arm;export GOARM=5;go build -o ${DIR}/${LUPO_CLIENT}-${A5} lupo-client/main.go

# Compile LUPO_CLIENT - Linux arm6 (ARMv6 / GOARM=6, VFPv1)
LUPO_CLIENT-arm6:
	export CGO_ENABLED=0;export GOOS=linux;export GOARCH=arm;export GOARM=6;go build -o ${DIR}/${LUPO_CLIENT}-${A6} lupo-client/main.go

# Compile LUPO_CLIENT - Linux arm64
LUPO_CLIENT-arm64:
	export CGO_ENABLED=0;export GOOS=linux;export GOARCH=arm64;go build -o ${DIR}/${LUPO_CLIENT}-${A64} lupo-client/main.go

# Compile LUPO_CLIENT - Linux mips64
LUPO_CLIENT-mips64:
	export CGO_ENABLED=0;export GOOS=linux;export GOARCH=mips64;go build -o ${DIR}/${LUPO_CLIENT}-${M64} lupo-client/main.go

# Compile LUPO_CLIENT - Linux mips64le
LUPO_CLIENT-mips64le:
	export CGO_ENABLED=0;export GOOS=linux;export GOARCH=mips64le;go build -o ${DIR}/${LUPO_CLIENT}-${M64LE} lupo-client/main.go

# Compile LUPO_CLIENT - Linux mipsle
LUPO_CLIENT-mipsle:
	export CGO_ENABLED=0;export GOOS=linux;export GOARCH=mipsle;go build -o ${DIR}/${LUPO_CLIENT}-${MLE} lupo-client/main.go

# Compile LUPO_CLIENT - Linux 386
LUPO_CLIENT-linux-386:
	export CGO_ENABLED=0;export GOOS=linux;export GOARCH=386;go build -o ${DIR}/${LUPO_CLIENT}-${L386} lupo-client/main.go

# Compile LUPO_CLIENT - Linux ppc64
LUPO_CLIENT-linux-ppc64:
	export CGO_ENABLED=0;export GOOS=linux;export GOARCH=ppc64;go build -o ${DIR}/${LUPO_CLIENT}-${LPPC64} lupo-client/main.go

# Compile LUPO_CLIENT - Linux ppc64le
LUPO_CLIENT-linux-ppc64le:
	export CGO_ENABLED=0;export GOOS=linux;export GOARCH=ppc64le;go build -o ${DIR}/${LUPO_CLIENT}-${LPPC64LE} lupo-client/main.go

# Compile LUPO_CLIENT - Linux s390x
LUPO_CLIENT-linux-s390x:
	export CGO_ENABLED=0;export GOOS=linux;export GOARCH=s390x;go build -o ${DIR}/${LUPO_CLIENT}-${LS390X} lupo-client/main.go

# Compile LUPO_CLIENT - Linux riscv64
LUPO_CLIENT-linux-riscv64:
	export CGO_ENABLED=0;export GOOS=linux;export GOARCH=riscv64;go build -o ${DIR}/${LUPO_CLIENT}-${LRISCV64} lupo-client/main.go

# Compile LUPO_CLIENT - Linux loong64
LUPO_CLIENT-linux-loong64:
	export CGO_ENABLED=0;export GOOS=linux;export GOARCH=loong64;go build -o ${DIR}/${LUPO_CLIENT}-${LLOONG64} lupo-client/main.go

# Compile LUPO_CLIENT - Windows 386
LUPO_CLIENT-windows-386:
	export CGO_ENABLED=0;export GOOS=windows;export GOARCH=386;go build -o ${DIR}/${LUPO_CLIENT}-${W386}.exe lupo-client/main.go

# Compile LUPO_CLIENT - Windows arm64
LUPO_CLIENT-windows-arm64:
	export CGO_ENABLED=0;export GOOS=windows;export GOARCH=arm64;go build -o ${DIR}/${LUPO_CLIENT}-${WARM64}.exe lupo-client/main.go

# Compile LUPO_CLIENT - FreeBSD x64
LUPO_CLIENT-freebsd:
	export CGO_ENABLED=0;export GOOS=freebsd;export GOARCH=amd64;go build -o ${DIR}/${LUPO_CLIENT}-${FBSD64} lupo-client/main.go

# Compile LUPO_CLIENT - FreeBSD 386
LUPO_CLIENT-freebsd-386:
	export CGO_ENABLED=0;export GOOS=freebsd;export GOARCH=386;go build -o ${DIR}/${LUPO_CLIENT}-${FBSD386} lupo-client/main.go

# Compile LUPO_CLIENT - FreeBSD arm
LUPO_CLIENT-freebsd-arm:
	export CGO_ENABLED=0;export GOOS=freebsd;export GOARCH=arm;go build -o ${DIR}/${LUPO_CLIENT}-${FBSDARM} lupo-client/main.go

# Compile LUPO_CLIENT - FreeBSD arm64
LUPO_CLIENT-freebsd-arm64:
	export CGO_ENABLED=0;export GOOS=freebsd;export GOARCH=arm64;go build -o ${DIR}/${LUPO_CLIENT}-${FBSDARM64} lupo-client/main.go

# Compile LUPO_CLIENT - OpenBSD x64
LUPO_CLIENT-openbsd:
	export CGO_ENABLED=0;export GOOS=openbsd;export GOARCH=amd64;go build -o ${DIR}/${LUPO_CLIENT}-${OBSD64} lupo-client/main.go

# Compile LUPO_CLIENT - OpenBSD 386
LUPO_CLIENT-openbsd-386:
	export CGO_ENABLED=0;export GOOS=openbsd;export GOARCH=386;go build -o ${DIR}/${LUPO_CLIENT}-${OBSD386} lupo-client/main.go

# Compile LUPO_CLIENT - OpenBSD arm
LUPO_CLIENT-openbsd-arm:
	export CGO_ENABLED=0;export GOOS=openbsd;export GOARCH=arm;go build -o ${DIR}/${LUPO_CLIENT}-${OBSDARM} lupo-client/main.go

# Compile LUPO_CLIENT - OpenBSD arm64
LUPO_CLIENT-openbsd-arm64:
	export CGO_ENABLED=0;export GOOS=openbsd;export GOARCH=arm64;go build -o ${DIR}/${LUPO_CLIENT}-${OBSDARM64} lupo-client/main.go

# Compile LUPO_CLIENT - NetBSD x64
LUPO_CLIENT-netbsd:
	export CGO_ENABLED=0;export GOOS=netbsd;export GOARCH=amd64;go build -o ${DIR}/${LUPO_CLIENT}-${NBSD64} lupo-client/main.go

# Compile LUPO_CLIENT - NetBSD 386
LUPO_CLIENT-netbsd-386:
	export CGO_ENABLED=0;export GOOS=netbsd;export GOARCH=386;go build -o ${DIR}/${LUPO_CLIENT}-${NBSD386} lupo-client/main.go

# Compile LUPO_CLIENT - NetBSD arm
LUPO_CLIENT-netbsd-arm:
	export CGO_ENABLED=0;export GOOS=netbsd;export GOARCH=arm;go build -o ${DIR}/${LUPO_CLIENT}-${NBSDARM} lupo-client/main.go

# Compile LUPO_CLIENT - NetBSD arm64
LUPO_CLIENT-netbsd-arm64:
	export CGO_ENABLED=0;export GOOS=netbsd;export GOARCH=arm64;go build -o ${DIR}/${LUPO_CLIENT}-${NBSDARM64} lupo-client/main.go

# Compile LUPO_CLIENT - Solaris x64
LUPO_CLIENT-solaris:
	export CGO_ENABLED=0;export GOOS=solaris;export GOARCH=amd64;go build -o ${DIR}/${LUPO_CLIENT}-${SOL} lupo-client/main.go

# Compile LUPO_CLIENT - Dragonfly x64
LUPO_CLIENT-dragonfly:
	export CGO_ENABLED=0;export GOOS=dragonfly;export GOARCH=amd64;go build -o ${DIR}/${LUPO_CLIENT}-${DRAG} lupo-client/main.go

clean:
	rm -rf ${DIR}*
