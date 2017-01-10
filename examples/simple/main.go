package main

import (
	"encoding/json"
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

func (c *Config) Path() string {
	path, _ := os.LookupEnv("CONFL_CONF_PATH")
	return path
}

func (c *Config) Unmarshal(data []byte) error {
	// vault.Secret only supports json.Unmarshal
	return json.Unmarshal(data, c)
}

func main() {
	config := &Config{}
	doneCh := make(chan struct{})
	stopCh := make(chan struct{})
	errCh := make(chan error, 10)
	setEnv()
	// set interval to 10 seconds just for test
	// you need set it a little bigger
	vault.DefaultInterval = 10 * time.Second

	watch, err := confl.NewWatcherFromEnv(config, doneCh, stopCh, errCh)
	if err != nil {
		panic(err)
	}
	watch.AddReloaders(confl.ReloadFunc(func() error {
		fmt.Printf("reload: %s\n", config.Username)
		fmt.Printf("reload: %s\n", config.Password.Value)
		return nil
	}))
	go watch.GoWatch()

	fmt.Printf("load: %s\n", config.Username)
	fmt.Printf("load: %s\n", config.Password)
	for err := range errCh {
		fmt.Println(err)
	}

}
