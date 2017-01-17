Confl
=====
Configuration reload with etcd, security storage with vault!

[![Build Status](http://img.shields.io/travis/teambition/confl.svg?style=flat-square)](https://travis-ci.org/teambition/confl)
[![Coverage Status](http://img.shields.io/coveralls/teambition/confl.svg?style=flat-square)](https://coveralls.io/r/teambition/confl)
[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/teambition/confl/master/LICENSE)
[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](http://godoc.org/github.com/teambition/confl)

## Features

* Simple API for use
* Used as a library
* Care about updates of configuration
* Support Auto-Reload

## Getting Started

### Install

#### Install confl

```shell
go get -u -v github.com/teambition/confl
```

#### Start vault with etcd backend

```shell
cat vault.hcl
```

```hcl
backend "etcd" {
  # etcd listen address
  address = "http://127.0.0.1:2379"
  # root path in etcd
  path = "vault/"
}

listener "tcp" {
  # listen address
  address = "localhost:8200"
  # secure config
  tls_disable = 1
  # tls_cert_file = /path/to/cert/file
  # tls_key_file = /path/to/key/file
  # tls_min_version = tls12
}
```

```shell
vault server -config=vault.hcl
```

### Usage

Use with environment variables:
```shell
# env var
# the configuration path
export CONFL_CONF_PATH=/path/to/configuration
# etcd cluster
# etcd cluster addresses
export CONFL_ETCD_CLUSTERS=http://node1.example.com:2379,http://node2.example.com:2379,http://node3.example.com:2379
# security connection
export CONFL_ETCD_CERT=/path/to/cert
export CONFL_ETCD_KEY=/path/to/key
export CONFL_ETCD_CACERT=/path/to/cacert
# etcd username/password for auth
export CONFL_ETCD_USERNAME=username
export CONFL_ETCD_PASSWORD=password

# vault var
# type of auth one in (app-id, token, github, userpass)
export CONFL_VAULT_AUTH_TYPE=token
export CONFL_VAULT_ADDRESS=http://localhost:8200
# case app-id
# this is more useful for micro services
# every micro service can be given a app_id to distinguish between identities
export CONFL_VAULT_APP_ID=app_id
export CONFL_VAULT_USER_ID=user_id
# case userpass
# auth with username/password
export CONFL_VAULT_USERNAME=username
export CONFL_VAULT_PASSWORD=password
# case token or github
export CONFL_VAULT_TOKEN=some token
# security connection
export CONFL_VAULT_CERT=/path/to/cert
export CONFL_VAULT_KEY=/path/to/key
export CONFL_VAULT_CACERT=/path/to/cacert
```

More [examples](examples/)

## Development

### Test

1. start a local etcd service:
```shell
etcd
```

2. start local valut service:
```shell
vault server -dev
```

3. test with cover:
```shell
make cover
```