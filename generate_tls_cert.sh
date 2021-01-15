#!/bin/bash
echo "This script is just a helper to quickly generate self-signed TLS certs for Lupo C2"
echo "This script depends on openssl to be installed, so if it is not please install it now..."
read -n 1 -s -r -p "press enter to continue..."
DAYS=$1
echo "Generating private key..."
echo ""
openssl genrsa -out lupo-server.key 2048
openssl ecparam -genkey -name secp384r1 -out lupo-server.key
echo ""
echo ""
if [ ! -z "$DAYS" ];
then
echo "Generating a cert with the following parameters:"
echo ""
echo "openssl req -new -x509 -sha256 -key lupo-server.key -out lupo-server.crt -days $DAYS -subj \"/C=US/ST=Lupo/L=Lupo/O=Lupo/OU=Lupo/CN=localhost\"  -addext \"subjectAltName = DNS:localhost\""
echo ""
openssl req -new -x509 -sha256 -key lupo-server.key -out lupo-server.crt -days $DAYS -subj "/C=US/ST=Lupo/L=Lupo/O=Lupo/OU=Lupo/CN=localhost"  -addext "subjectAltName = DNS:localhost" 
else
echo "Generating a cert with the following parameters:"
echo ""
echo "openssl req -new -x509 -sha256 -key lupo-server.key -out lupo-server.crt -days 3650 -subj \"/C=US/ST=Lupo/L=Lupo/O=Lupo/OU=Lupo/CN=localhost\"  -addext \"subjectAltName = DNS:localhost\""
echo ""
echo "To change the number of days pass in a parameter to the script and re-run it"
echo "example: generate_tls_cert.sh <days>"
echo ""
openssl req -new -x509 -sha256 -key lupo-server.key -out lupo-server.crt -days 3650 -subj "/C=US/ST=Lupo/L=Lupo/O=Lupo/OU=Lupo/CN=localhost" -addext "subjectAltName = DNS:localhost" 
fi
echo ""
echo "Generating PEM file to use with Lupo implants/clients..."
cat lupo-server.key > lupo-server.pem
cat lupo-server.crt >> lupo-server.pem
echo ""
echo "Place the key and crt files in the same directory as the lupo server binary."
echo "Alternatively, specify them with the appropriate arguments when starting an HTTPS listener."
echo "By default Lupo will look for a key and cert named lupo-server.key and lupo-server.crt in its current directory."
echo "You may also specify the key and cert locations via custom arguments in the listener."