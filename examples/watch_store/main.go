package main

import (
	"fmt"
	"time"

	"github.com/teambition/confl"
	"github.com/teambition/confl/examples/common"
)

// How it work?
//
// First:
// write `{"username": "username"}` to etcd's key /confl/test
// cli: etcdctl set /confl/test '{"username": "username"}'
//
// Second:
// write `123456` to vault's key secret/password and
// cli: vault write secret/password value=123456
// write `123456` to vault's key secret/token
// cli: vault write secret/token value=123456
//
// Then:
// watcher will get {"username": "username", "password": "123456", token:"123456"}

func main() {
	watcher, err := confl.NewStoreWatcher(&common.Config{}, "/confl/test", confl.Options{
		Etcd: confl.Etcd{
			Clusters: []string{"http://localhost:2379"},
		},
		Vault: confl.Vault{
			AuthType: "token",
			Address:  "http://localhost:8200",
			Token:    "teambition",
			// set interval to 10 seconds just for test
			// you need set it a little bigger in production
			// perhaps DefaultInterval 5 minutes just ok
			Interval: "10s",
		},
	})
	if err != nil {
		panic(err)
	}
	defer watcher.Close()

	watcher.OnError(func(err error) {
		fmt.Println("your error handler start")
		fmt.Println(err)
	})

	// add hook for update events
	// perhaps you need reload something that depends the configuration
	watcher.AddHook(func(oc, nc interface{}) {
		ocfg := oc.(common.Config)
		ncfg := nc.(common.Config)
		// use cfg
		fmt.Printf("change old username: %s\n", ocfg.Username)
		fmt.Printf("change old password: %s\n", ocfg.Password)
		fmt.Printf("change old token: %s\n", ocfg.Token)
		fmt.Printf("change new username: %s\n", ncfg.Username)
		fmt.Printf("change new password: %s\n", ncfg.Password)
		fmt.Printf("change new token: %s\n", ncfg.Token)
	})

	// start watch
	// it is a blocking method
	go watcher.Watch()

	// get configuration from watcher
	cfg := watcher.Config().(common.Config)
	// use cfg
	fmt.Printf("load username: %s\n", cfg.Username)
	fmt.Printf("load password: %s\n", cfg.Password)
	fmt.Printf("load password: %s\n", cfg.Token)
	time.Sleep(time.Hour)
}
