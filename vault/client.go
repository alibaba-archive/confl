package vault

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/kelseyhightower/envconfig"
)

var (
	DefaultInterval = 5 * time.Minute
	defaultClient   *Client
)

type Client struct {
	*vaultapi.Client
	kvs map[string]string
}

// NewFromEnv initialize the Client with environment variables
func NewClientFromEnv() (*Client, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, err
	}

	return NewClient(cfg)
}

func NewClient(cfg Config) (*Client, error) {
	if defaultClient != nil {
		return defaultClient, nil
	}

	if cfg.AuthType == None {
		return nil, errors.New("you have to set the auth type when using the vault backend")
	}

	var (
		vcfg   = vaultapi.DefaultConfig()
		tlsCfg = &tls.Config{}
	)

	vcfg.Address = cfg.Address

	if cfg.Cert != "" && cfg.Key != "" {
		certificate, err := tls.LoadX509KeyPair(cfg.Cert, cfg.Key)
		if err != nil {
			return nil, err
		}
		tlsCfg.Certificates = []tls.Certificate{certificate}
		tlsCfg.BuildNameToCertificate()
	}

	if cfg.CAcert != "" {
		cacert, err := ioutil.ReadFile(cfg.CAcert)
		if err != nil {
			return nil, err
		}
		certPool := x509.NewCertPool()
		certPool.AppendCertsFromPEM(cacert)
		tlsCfg.RootCAs = certPool
	}

	vcfg.HttpClient.Transport = &http.Transport{TLSClientConfig: tlsCfg}

	client, err := vaultapi.NewClient(vcfg)
	if err != nil {
		return nil, err
	}

	// auth typep
	var secret *vaultapi.Secret

	// check the auth type and authenticate the vault service
	switch cfg.AuthType {
	case AppID:
		secret, err = client.Logical().Write("/auth/app-id/login", map[string]interface{}{
			"app_id":  cfg.AppID,
			"user_id": cfg.UserID,
		})
	case Github:
		secret, err = client.Logical().Write("/auth/github/login", map[string]interface{}{
			"token": cfg.Token,
		})
	case Token:
		client.SetToken(cfg.Token)
		secret, err = client.Logical().Read("/auth/token/lookup-self")
	case Pass:
		secret, err = client.Logical().Write(fmt.Sprintf("/auth/userpass/login/%s", cfg.Username), map[string]interface{}{
			"password": cfg.Password,
		})
	}

	if err != nil {
		return nil, err
	}

	if client.Token() == "" {
		client.SetToken(secret.Auth.ClientToken)
	}

	defaultClient = &Client{client, map[string]string{}}
	return defaultClient, nil
}

func (c *Client) addKV(key, value string) {
	c.kvs[key] = value
}

func (c *Client) Key(key string) (string, error) {
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

// WatchKey the key changes from etcd until be stopped
func (c *Client) WatchKey(key string, reloadCh chan<- struct{}, stopCh <-chan struct{}, errCh chan<- error) {
	t := time.Tick(DefaultInterval)
	for {
		select {
		case <-stopCh:
			return
		case <-t:
			for key, value := range c.kvs {
				v, err := c.Key(key)
				if err != nil {
					errCh <- err
					continue
				}
				if value != v {
					reloadCh <- struct{}{}
					break
				}
			}
		}
	}
}

func (c *Client) Close() error {
	return c.Close()
}
