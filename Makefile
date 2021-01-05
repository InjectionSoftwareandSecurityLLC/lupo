# !!!MAKE SURE YOUR GOPATH ENVIRONMENT VARIABLE IS SET FIRST!!!

# Variables
DIR=builds/
LUPO_SERVER=LUPO_SERVER
WIN_LDFLAGS=-ldflags "-H=windowsgui"
W=Windows-x64
L=Linux-x64
A=Linux-arm
M=Linux-mips
D=Darwin-x64

# Make Directory to store executables
$(shell mkdir -p ${DIR})

# Change default to just make for the host OS and add MAKE ALL to do this
default: LUPO_SERVER-windows LUPO_SERVER-linux LUPO_SERVER-darwin

all: default

# Compile Windows binaries
windows: LUPO_SERVER-windows

# Compile Linux binaries
linux: LUPO_SERVER-linux

# Compile Arm binaries
arm: LUPO_SERVER-arm

# Compile mips binaries
mips: LUPO_SERVER-mips

# Compile Darwin binaries
darwin: LUPO_SERVER-darwin

# Compile LUPO_SERVER - Windows x64
LUPO_SERVER-windows:
	export GOOS=windows GOARCH=amd64;go build ${WIN_LDFLAGS} -o ${DIR}/${LUPO_SERVER}-${W}.exe lupo-server/main.go

# Compile LUPO_SERVER - Linux mips
LUPO_SERVER-mips:
	export GOOS=linux;export GOARCH=mips;go build -o ${DIR}/${LUPO_SERVER}-${M} lupo-server/main.go

# Compile LUPO_SERVER - Linux arm
LUPO_SERVER-arm:
	export GOOS=linux;export GOARCH=arm;export GOARM=7;go build -o ${DIR}/${LUPO_SERVER}-${A} lupo-server/main.go

# Compile LUPO_SERVER - Linux x64
LUPO_SERVER-linux:
	export GOOS=linux;export GOARCH=amd64;go build -o ${DIR}/${LUPO_SERVER}-${L} lupo-server/main.go

# Compile LUPO_SERVER - Darwin x64
LUPO_SERVER-darwin:
	export GOOS=darwin;export GOARCH=amd64;go build -o ${DIR}/${LUPO_SERVER}-${D} lupo-server/main.go

clean:
	rm -rf ${DIR}*