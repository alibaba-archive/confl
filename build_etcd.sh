#!/bin/sh

VERSION="v3.1.0-rc.1"
FILE_ROOT="etcd-${VERSION}-linux-amd64"
FILE_NAME="${FILE_ROOT}.tar.gz"

curl -L "https://github.com/coreos/etcd/releases/download/${VERSION}/${FILE_NAME}" -o "${FILE_NAME}"
tar xzvf "${FILE_NAME}"
cd $FILE_ROOT

echo "Starting etcd"
./etcd &
cd -
sleep 1
