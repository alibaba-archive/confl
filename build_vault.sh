#!/bin/sh
https://releases.hashicorp.com/vault/0.6.4/vault_0.6.4_linux_amd64.zip
VERSION="0.6.4"
FILE_ROOT="vault_${VERSION}_linux_amd64"
FILE_NAME="${FILE_ROOT}.zip"

curl -L "https://releases.hashicorp.com/vault/${VERSION}/${FILE_NAME}" -o "${FILE_NAME}"
unzip "${FILE_NAME}" -d $FILE_ROOT
cd $FILE_ROOT

echo "Starting vault"
./vault server -dev &
export VAULT_ADDR='http://127.0.0.1:8200'
sleep 1
