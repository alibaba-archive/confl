#!/bin/sh

VERSION="v2.2.2"
FILE_ROOT="etcd-${VERSION}-linux-amd64"
FILE_NAME="${FILE_ROOT}.tar.gz"

curl -L "https://github.com/coreos/etcd/releases/download/${VERSION}/${FILE_NAME}" -o "${FILE_NAME}"
tar xzvf "${FILE_NAME}"
cd $FILE_ROOT

echo "Starting etcd"
./etcd &
sleep 3
