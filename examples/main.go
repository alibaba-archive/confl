package main

import (
	"context"
	"fmt"

	"github.com/teambition/confl"
	"github.com/teambition/confl/etcd"
	"github.com/teambition/confl/vault"
)

func main() {
	etcdConfig := etcd.Config{
		Clusters: []string{"http://localhost:2379"},
	}

	etcdConfig.ConfPath = "/teambition/auth-production"

	cl, err := etcd.NewClient(etcdConfig)
	if err != nil {
		panic(err)
	}

	vaultConfig := vault.Config{
		AuthType: vault.AuthToken,
		Token:    "273e8292-e45d-7da1-2560-3118adbe01c0",
		Address:  "http://localhost:8200",
	}

	v, err := vault.New(vaultConfig)
	if err != nil {
		panic(err)
	}

	type config struct {
		Username string        `json:"username"`
		Password *vault.Secret `json:"password"`
		Phone    *vault.Secret `json:"phone"`
	}

	c := &config{
		Password: v.Secret(),
		Phone:    v.Secret(),
	}

	reload := func() error {
		fmt.Printf("%#v\n", c.Password.Value)
		fmt.Printf("%#v\n", c.Phone.Value)
		return nil
	}

	// write to etcd
	_ = `{"username": "xushuai", "password": "VAULT(secret/password)", "phone": "VAULT(secret/phone)"}`

	errChan := make(chan error)
	go func() {
		err := cl.WatchConfig(context.Background(), c, confl.ReloadFunc(reload), errChan)
		_ = err
	}()

	for range errChan {
	}
}
