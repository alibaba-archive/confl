package main

import (
	"fmt"
	"os"
	"time"

	"github.com/teambition/confl"
	"github.com/teambition/confl/vault"
)

// just for test
func setEnv() {
	os.Clearenv()
	os.Setenv("CONFL_CONF_PATH", "/confl/test")
	os.Setenv("CONFL_ETCD_CLUSTERS", "http://localhost:2379")
	os.Setenv("CONFL_VAULT_AUTH_TYPE", "token")
	os.Setenv("CONFL_VAULT_ADDRESS", "http://localhost:8200")
	os.Setenv("CONFL_VAULT_TOKEN", "06900225-b34b-69de-7872-21a2c8b52306")
}

type Config struct {
	Username string        `json:"username"`
	Password *vault.Secret `json:"password"`
}

func main() {
	config := &Config{}
	setEnv()
	// set interval to 10 seconds just for test
	// you need set it a little bigger
	vault.DefaultInterval = 10 * time.Second

	watch, err := confl.NewFromEnv(config, nil)
	if err != nil {
		panic(err)
	}

	watch.AddHook(func(c interface{}) {
		if config, ok := c.(Config); ok {
			fmt.Printf("change username: %s\n", config.Username)
			fmt.Printf("change password: %s\n", config.Password.Value)
		}
	})
	go watch.GoWatch()
	fmt.Printf("load username: %s\n", config.Username)
	fmt.Printf("load password: %s\n", config.Password.Value)
	time.Sleep(time.Hour)
}
