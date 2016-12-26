package sources

import (
	"errors"

	"golang.org/x/net/context"
)

type Response struct {
	Key       string
	NextIndex uint64
	Error     error
}

type Type int

const (
	None Type = iota
	Etcd
	Vault
	Consul
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
	Watch(ctx context.Context, key string) <-chan *Response
	Key(ctx context.Context, key string) (string, error)
	WithClient(Client)
	Type() Type
}

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
		return NewEtcdClient(hosts, config.ClientCert, config.ClientKey, config.ClientCaKeys, config.BasicAuth, config.Username, config.Password)
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
		return NewVaultClient(hosts[0], config.AuthType, vaultConfig)
	}
	return nil, errors.New("Invalid sources")
}
