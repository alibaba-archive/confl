package vault

import (
	"errors"
	"fmt"
	"sync"
	"time"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/teambition/confl/util"
)

var (
	DefaultInterval = 5 * time.Minute
)

var (
	defaultClient *client
)

type client struct {
	*vaultapi.Client
	c      *Config
	mu     sync.RWMutex
	kvs    map[string]string
	stopCh chan struct{}
}

// Init initialize the vault defaultClient for r/w sercrets
func Init(cfg *Config) error {
	if cfg.AuthType == None {
		return errors.New("you have to set the auth type when using the vault backend")
	}

	if cfg.Interval == 0 {
		cfg.Interval = DefaultInterval
	}

	if cfg.ChangeCh == nil {
		return errors.New("need change channel for watch changes")
	}

	var (
		err  error
		vcfg = vaultapi.DefaultConfig()
	)

	vcfg.Address = cfg.Address
	vcfg.HttpClient.Transport, err = util.SecureTransport(cfg.CAcert, cfg.Cert, cfg.Key)
	if err != nil {
		return err
	}

	cl, err := vaultapi.NewClient(vcfg)
	if err != nil {
		return err
	}

	// auth type
	var secret *vaultapi.Secret

	// check the auth type and authenticate the vault service
	switch cfg.AuthType {
	case AppID:
		secret, err = cl.Logical().Write("/auth/app-id/login", map[string]interface{}{
			"app_id":  cfg.AppID,
			"user_id": cfg.UserID,
		})
	case Github:
		secret, err = cl.Logical().Write("/auth/github/login", map[string]interface{}{
			"token": cfg.Token,
		})
	case Token:
		cl.SetToken(cfg.Token)
		secret, err = cl.Logical().Read("/auth/token/lookup-self")
	case Pass:
		secret, err = cl.Logical().Write(fmt.Sprintf("/auth/userpass/login/%s", cfg.Username), map[string]interface{}{
			"password": cfg.Password,
		})
	default:
		return fmt.Errorf("unsupported auth type(%s)", cfg.AuthType)
	}

	if err != nil {
		return err
	}

	if cl.Token() == "" {
		cl.SetToken(secret.Auth.ClientToken)
	}

	defaultClient = &client{
		Client: cl,
		c:      cfg,
		kvs:    map[string]string{},
		stopCh: make(chan struct{}),
	}

	go defaultClient.watch()
	return nil
}

// addKV when config struct contains *vault.Secret type
// then add it's Key and Value to kvs for watch
func (c *client) addKV(key, value string) {
	c.mu.Lock()
	c.kvs[key] = value
	c.mu.Unlock()
}

// key get the value by given key
// value only support string type
func (c *client) key(key string) (string, error) {
	resp, err := c.Logical().Read(key)
	if err != nil {
		return "", err
	}

	if resp == nil || resp.Data == nil {
		return "", fmt.Errorf("vault secret key(%s) is not existed", key)
	}

	if value, ok := resp.Data["value"]; ok {
		// value just can only be string type
		if text, ok := value.(string); ok {
			return text, nil
		}
	}
	return "", fmt.Errorf("vault secret key(%s) value needs a string type", key)
}

// watch the key changes from kvs
// it is triggered every c.c.Interval
func (c *client) watch() {
	t := time.Tick(c.c.Interval)
	for {
		select {
		case <-c.stopCh:
			return
		case <-t:
			c.mu.RLock()
			for key, value := range c.kvs {
				v, err := c.key(key)
				if err != nil {
					if c.c.OnError != nil {
						c.c.OnError(err)
					}
					continue
				}
				if value != v {
					c.c.ChangeCh <- struct{}{}
					break
				}
			}
			c.mu.RUnlock()
		}
	}
}

func (c *client) close() error {
	close(c.stopCh)
	return nil
}

// Close close the defaultClient
func Close() error {
	if defaultClient != nil {
		return defaultClient.close()
	}
	return nil
}
