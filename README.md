# confl

Watch a distributed store and reload configurate.


## Features

* Used as a library
* Auto-Reloading

## Getting Started

```shell
go get github.com/teambition/confl
```


#### from etcd and vault

```go
package main

import (
	"context"
	"fmt"

	"github.com/teambition/confl"
	"github.com/teambition/confl/encoding/json"
	"github.com/teambition/confl/sources"
)

func main() {
	etcdConfig := sources.Config{
		Type:  sources.Etcd,
		Hosts: []string{"http://localhost:2379"},
	}

	etcdClient, err := sources.New(etcdConfig)
	if err != nil {
		panic(err)
	}

	vaultConfig := sources.Config{
		Type:      sources.Vault,
		AuthType:  "token",
		AuthToken: "273e8292-e45d-7da1-2560-3118adbe01c0",
		Hosts:     []string{"http://localhost:8200"},
	}

	vaultClient, err := sources.New(vaultConfig)
	if err != nil {
		panic(err)
	}

	type config struct {
		Username string      `json:"username"`
		Password *json.Vault `json:"password"`
		Phone    *json.Vault `json:"phone"`
	}

	c := &config{
		Password: json.NewVault(vaultClient),
		Phone:    json.NewVault(vaultClient),
	}

	reload := func() error {
		fmt.Printf("%#v\n", c)
		fmt.Printf("%#v\n", c.Password.Value())
		fmt.Printf("%#v\n", c.Phone.Value())
		return nil
	}

	// write {"username": "xushuai", "password": "VAULT: secret/password", "phone": "VAULT: secret/phone"} to etcd

	err = confl.LoadJSONAndWatch(context.Background(), c, "/teambition/auth-production", etcdClient, reload)
	if err != nil {
		panic(err)
	}

}
```

