package main

import (
	"fmt"

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
		fmt.Printf("old config: %#v\n", ocfg)
		fmt.Printf("new config: %#v\n", ncfg)
	})

	// get configuration from watcher
	cfg := watcher.Config().(common.Config)
	// use cfg
	fmt.Printf("load config: %#v\n", cfg)

	// start watch
	// it is a blocking method choose run with `go` by situation
	watcher.Watch()
}
