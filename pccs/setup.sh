#!/bin/bash

set -e

# generate certificate and private key
mkdir -p ./ssl_key
cd ./ssl_key
openssl genrsa -out private.pem 2048
openssl req -new -key private.pem -out csr.pem -subj '/CN=localhost'
openssl x509 -req -days 365 -in csr.pem -signkey private.pem -out file.crt
rm -rf csr.pem && chmod 644 ./*


# set values in config
cd ../config/
if [ -n "$APIKEY" ]; then
    apikey="$APIKEY"
    sed -i "s/\"ApiKey\"[ ]*:[ ]*\"\",/\"ApiKey\": \"$apikey\",/" default.json
else
    echo "ERROR: You need to submit an APIKEY, otherwise your PCCS can not connect to the PCS"
    exit 1
fi

if [ -n "$USERPASS" ]; then
    userpass="$USERPASS"
else
    userpass=$(openssl rand -hex 64)
fi
userhash=$(echo -n "$userpass" | sha512sum | tr -d '[:space:]-')
sed -i "s/\"UserTokenHash\"[ ]*:[ ]*\"\",/\"UserTokenHash\": \"$userhash\",/" default.json

if [ -n "$ADMINPASS" ]; then
    adminpass="$ADMINPASS"
else
    adminpass=$(openssl rand -hex 64)
fi
adminhash=$(echo -n "$adminpass" | sha512sum | tr -d '[:space:]-')
sed -i "s/\"AdminTokenHash\"[ ]*:[ ]*\"\",/\"AdminTokenHash\": \"$adminhash\",/" default.json

sed -i 's/\"hosts\"[ ]*:[ ]*\"127.0.0.1\",/\"hosts\": \"0.0.0.0\",/' default.json

cd ..
/usr/bin/node /opt/intel/pccs/pccs_server.js
