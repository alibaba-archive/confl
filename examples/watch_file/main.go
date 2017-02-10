package main

import (
	"fmt"
	"time"

	"github.com/teambition/confl"
	"github.com/teambition/confl/examples/common"
)

func main() {
	watcher, err := confl.NewFileWatcher(&common.Config{}, "../common/config.json")
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
	fmt.Printf("load token: %s\n", cfg.Token)

	time.Sleep(time.Hour)
}
