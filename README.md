# confl

[![Build Status](https://travis-ci.org/teambition/confl.svg?branch=master)](https://travis-ci.org/teambition/confl)

Configuration reload with etcd, security storage with vault!

## Features

* Simple API for use
* Used as a library
* Care about updates of configuration
* Support Auto-Reload

## Getting Started

### install

#### install confl 

```shell
go get -u -v github.com/teambition/confl
```

#### vault with etcd backend

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

#### usage

used with environment variables

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

examples [examples](examples/)
