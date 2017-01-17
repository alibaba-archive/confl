package main

import (
	"fmt"
	"os"
	"time"

	"github.com/teambition/confl"
	"github.com/teambition/confl/vault"
)

// How it work?
//
// First, write `{"username": "username", "password": "VAULT(secret/password)"}` to etcd's key /confl/test
// cli: etcdctl set /confl/test '{"username": "username", "password": "VAULT(secret/password)"}'
//
// Second, write `123456` to vault's key secret/password
// cli: vault write secret/password value=123456
//
// Then, watcher will get {"username": "username", "password": "123456"}

func main() {
	// all supported enviroment variables see README
	// perfect fit the docker, k8s and swarm etc
	os.Clearenv()
	os.Setenv("CONFL_CONF_PATH", "/confl/test")
	os.Setenv("CONFL_ETCD_CLUSTERS", "http://localhost:2379")
	os.Setenv("CONFL_VAULT_AUTH_TYPE", "token")
	os.Setenv("CONFL_VAULT_ADDRESS", "http://localhost:8200")
	os.Setenv("CONFL_VAULT_TOKEN", "f1103197-d8a4-9ea7-026f-04fae02561af")

	// set interval to 10 seconds just for test
	// you need set it a little bigger in production
	// perhaps DefaultInterval 5 minutes just ok
	vault.DefaultInterval = 10 * time.Second

	// you configuration struct
	// now just support json unmarshal
	type Config struct {
		Username string `json:"username"`
		// *vault.Secret is a secure type for store password, secret, token etc
		// it will load value from vault which is a tool for managing secrets
		Password vault.Secret `json:"password"`
	}

	watcher, err := confl.NewFromEnv(&Config{}, nil)
	if err != nil {
		panic(err)
	}

	// add hook for update events
	// perhaps you need reload something that depends the configuration
	watcher.AddHook(func(oc, nc interface{}) {
		ocfg := oc.(Config)
		ncfg := nc.(Config)
		// use cfg
		fmt.Printf("change old username: %s\n", ocfg.Username)
		fmt.Printf("change old password: %s\n", ocfg.Password.Value)
		fmt.Printf("change new username: %s\n", ncfg.Username)
		fmt.Printf("change new password: %s\n", ncfg.Password.Value)
	})

	// start watch
	// it is a blocking method
	go watcher.GoWatch()

	// get configuration from watcher
	cfg := watcher.Config().(Config)
	// use cfg
	fmt.Printf("load username: %s\n", cfg.Username)
	fmt.Printf("load password: %s\n", cfg.Password.Value)

	time.Sleep(time.Hour)
}
