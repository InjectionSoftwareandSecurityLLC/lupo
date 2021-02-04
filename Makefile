# !!!MAKE SURE YOUR GOPATH ENVIRONMENT VARIABLE IS SET FIRST!!!

# Variables
DIR=builds/
LUPO_SERVER=LUPO_SERVER
LUPO_CLIENT=LUPO_CLIENT
WIN_LDFLAGS=-ldflags "-H=windowsgui"
W=Windows-x64
L=Linux-x64
A=Linux-arm
M=Linux-mips
D=Darwin-x64

# Make Directory to store executables
$(shell mkdir -p ${DIR})

# Install go dependencies
$(shell go get ./lupo-server)
$(shell go get ./lupo-client)
$(shell go get ./sample)


# Change default to just make for the host OS and add MAKE ALL to do this
default: LUPO_SERVER-windows LUPO_SERVER-linux LUPO_SERVER-darwin LUPO_CLIENT-windows LUPO_CLIENT-linux LUPO_CLIENT-darwin

all: default

# Compile Windows binaries
windows: LUPO_SERVER-windows LUPO_CLIENT-windows

# Compile Linux binaries
linux: LUPO_SERVER-linux LUPO_CLIENT-linux

# Compile Darwin binaries
darwin: LUPO_SERVER-darwin LUPO_CLIENT-darwin

# Compile Arm binaries
arm: LUPO_SERVER-arm LUPO_CLIENT-arm

# Compile mips binaries
mips: LUPO_SERVER-mips LUPO_CLIENT-mips

# Compile LUPO_SERVER - Windows x64
LUPO_SERVER-windows:
	export GOOS=windows GOARCH=amd64;go build ${WIN_LDFLAGS} -o ${DIR}/${LUPO_SERVER}-${W}.exe lupo-server/main.go

# Compile LUPO_SERVER - Linux x64
LUPO_SERVER-linux:
	export GOOS=linux;export GOARCH=amd64;go build -o ${DIR}/${LUPO_SERVER}-${L} lupo-server/main.go

# Compile LUPO_SERVER - Darwin x64
LUPO_SERVER-darwin:
	export GOOS=darwin;export GOARCH=amd64;go build -o ${DIR}/${LUPO_SERVER}-${D} lupo-server/main.go

# Compile LUPO_SERVER - Linux mips
LUPO_SERVER-mips:
	export GOOS=linux;export GOARCH=mips;go build -o ${DIR}/${LUPO_SERVER}-${M} lupo-server/main.go

# Compile LUPO_SERVER - Linux arm
LUPO_SERVER-arm:
	export GOOS=linux;export GOARCH=arm;export GOARM=7;go build -o ${DIR}/${LUPO_SERVER}-${A} lupo-server/main.go

# Compile LUPO_CLIENT - Windows x64
LUPO_CLIENT-windows:
	export GOOS=windows GOARCH=amd64;go build ${WIN_LDFLAGS} -o ${DIR}/${LUPO_CLIENT}-${W}.exe lupo-client/main.go

# Compile LUPO_CLIENT - Linux x64
LUPO_CLIENT-linux:
	export GOOS=linux;export GOARCH=amd64;go build -o ${DIR}/${LUPO_CLIENT}-${L} lupo-client/main.go

# Compile LUPO_CLIENT - Darwin x64
LUPO_CLIENT-darwin:
	export GOOS=darwin;export GOARCH=amd64;go build -o ${DIR}/${LUPO_CLIENT}-${D} lupo-client/main.go

# Compile LUPO_CLIENT - Linux mips
LUPO_CLIENT-mips:
	export GOOS=linux;export GOARCH=mips;go build -o ${DIR}/${LUPO_CLIENT}-${M} lupo-client/main.go

# Compile LUPO_CLIENT - Linux arm
LUPO_CLIENT-arm:
	export GOOS=linux;export GOARCH=arm;export GOARM=7;go build -o ${DIR}/${LUPO_CLIENT}-${A} lupo-client/main.go

clean:
	rm -rf ${DIR}*