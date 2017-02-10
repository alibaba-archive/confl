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

#### Watch a configuration file

More [example](examples/watch_file)


#### Watch a backend storage

```go
// Options

// vault
// use https://github.com/kelseyhightower/envconfig to parse the environment variables for config
// it's container-friendly like docker, rocket
type Config struct {
	// type of auth one in (app-id, token, github, userpass)
	AuthType string
	// vault service address
	Address string
	// AuthType = app-id
	// "app id auth backend"(See https://www.vaultproject.io/docs/auth/app-id.html)
	// this is more useful for micro services
	// every micro service can be given a app_id to distinguish between identities
	AppID  string
	UserID string
	// AuthType = userpass
	// "userpass auth backend"(see https://www.vaultproject.io/docs/auth/userpass.html)
	Username string
	Password string
	// AuthType = token or github
	// auth token (See https://www.vaultproject.io/docs/auth/token.html)
	Token string
	// security connection
	// Cert and Key are the pair of x509
	// the path of certificate file
	Cert string
	// the path of certificate'key file
	Key string
	// the path of CACert file
	CAcert string
	// loop interval
	// vaule likes `10s` `1m` `1h`
	Interval string
}

// etcd
// Config etcd configuration
type Config struct {
	// cluseter addresses
	Clusters []string
	// security connection
	// the path of certificate file
	Cert   string 
	// the path of certificate'key file
	Key    string 
	// the path of CACert file
	CAcert string 
	// auth user/pass
	Username string
	Password string
}
```

More [example](examples/watch_store)

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
