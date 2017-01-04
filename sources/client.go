package sources

import (
	"errors"

	"github.com/teambition/confl/sources/common"
	"github.com/teambition/confl/sources/etcd"
	"github.com/teambition/confl/sources/vault"
	"golang.org/x/net/context"
)

// The Client interface needed impled by a k/v store like etcd, consul, redis etc.
type Client interface {
	// Watch the changes of given keys at the root node.
	// the waitIndex is the last index get changes.
	// example:
	/*
		for resp := range client.Watch(ctx, "/hello") {
			if resp.Error != nil {
				continue
			}
			someValue, err := client.Key(context.Background(), "/hello")
			if err != nil {
				continue
			}
			reload()
		}
	*/
	Watch(ctx context.Context, key string) <-chan *common.Response
	Key(ctx context.Context, key string) (string, error)
}

type Type int

const (
	None Type = iota
	Consul
	Etcd
	Vault
)

// New is used to create a storage client based on our configuration.
func New(config Config) (Client, error) {
	if config.Type == None {
		return nil, errors.New("need set client type")
	}
	hosts := config.Hosts
	switch config.Type {
	case Consul:
		// TODO
	case Etcd:
		// Create the etcd client upfront and use it for the life of the process.
		// The etcdClient is an http.Client and designed to be reused.
		return etcd.NewClient(hosts, config.ClientCert, config.ClientKey, config.ClientCaKeys, config.BasicAuth, config.Username, config.Password)
	case Vault:
		vaultConfig := map[string]string{
			"app-id":   config.AppID,
			"user-id":  config.UserID,
			"username": config.Username,
			"password": config.Password,
			"token":    config.AuthToken,
			"cert":     config.ClientCert,
			"key":      config.ClientKey,
			"caCert":   config.ClientCaKeys,
		}
		return vault.NewClient(hosts[0], config.AuthType, vaultConfig)
	}
	return nil, errors.New("Invalid sources")
}
