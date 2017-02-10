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
	defaultInterval = 5 * time.Minute
)

type Client struct {
	*vaultapi.Client
	mu     sync.RWMutex
	kvs    map[string]string
	stopCh chan struct{}
	//  watch interval
	interval  time.Duration
	changeCh  chan struct{}
	errHandle func(err error)
}

func New(cfg Config, changeCh chan struct{}, errHandle ...func(err error)) (*Client, error) {
	if changeCh == nil {
		return nil, errors.New("need change channel for watch changes")
	}

	if cfg.AuthType == "" {
		return nil, errors.New("you have to set the auth type when using the vault backend")
	}

	var (
		err      error
		interval time.Duration
	)

	if cfg.Interval != "" {
		if interval, err = time.ParseDuration(cfg.Interval); err != nil {
			return nil, err
		}
	} else {
		interval = defaultInterval
	}

	vcfg := vaultapi.DefaultConfig()
	vcfg.Address = cfg.Address
	vcfg.HttpClient.Transport, err = util.SecureTransport(cfg.CAcert, cfg.Cert, cfg.Key)
	if err != nil {
		return nil, err
	}

	cl, err := vaultapi.NewClient(vcfg)
	if err != nil {
		return nil, err
	}

	// auth type
	var secret *vaultapi.Secret

	// check the auth type and authenticate the vault service
	switch AuthType(cfg.AuthType) {
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
		return nil, fmt.Errorf("unsupported auth type(%s)", cfg.AuthType)
	}

	if err != nil {
		return nil, err
	}

	if cl.Token() == "" {
		cl.SetToken(secret.Auth.ClientToken)
	}

	c := &Client{
		Client:   cl,
		kvs:      map[string]string{},
		stopCh:   make(chan struct{}),
		interval: interval,
		changeCh: changeCh,
	}

	if len(errHandle) == 1 {
		c.errHandle = errHandle[0]
	}
	go c.watch()
	return c, nil
}

// addKV when config struct contains *vault.Secret type
// then add it's Key and Value to kvs for watch
func (c *Client) addKV(key, value string) {
	c.mu.Lock()
	c.kvs[key] = value
	c.mu.Unlock()
}

// key get the value by given key
// value only support string type
func (c *Client) key(key string) (string, error) {
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
func (c *Client) watch() {
	t := time.Tick(c.interval)
	for {
		select {
		case <-c.stopCh:
			return
		case <-t:
			c.mu.RLock()
			for key, value := range c.kvs {
				v, err := c.key(key)
				if err != nil {
					if c.errHandle != nil {
						c.errHandle(err)
					}
					continue
				}
				if value != v {
					c.changeCh <- struct{}{}
					break
				}
			}
			c.mu.RUnlock()
		}
	}
}

func (c *Client) Close() error {
	close(c.stopCh)
	return nil
}
