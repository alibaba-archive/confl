package main

import (
	j "encoding/json"
	"fmt"
	"time"

	"github.com/teambition/confl/encoding/json"
	"github.com/teambition/confl/sources"
)

type config struct {
	Username string      `json:"username"`
	Password *json.Vault `json:"password"`
	Phone    *json.Vault `json:"phone"`
}

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

	vaultClient.WithClient(etcdClient)

	reload := func() error {
		fmt.Println("reloading")
		return nil
	}

	v, err := json.NewVault("/vault", vaultClient, reload)
	if err != nil {
		panic(err)
	}

	c := &config{
		Password: v,
		Phone:    v.Copy(),
	}

	raw := `{"username": "xushuai", "password": "VAULT: secret/password", "phone": "VAULT: secret/phone"}`

	j.Unmarshal([]byte(raw), c)
	fmt.Printf("%#v\n", c.Password.Value())
	time.Sleep(time.Hour)
}
