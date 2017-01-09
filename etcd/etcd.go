package etcd

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/kelseyhightower/envconfig"
	"github.com/teambition/confl"
)

type Client struct {
	confPath string
	client   client.KeysAPI
}

func NewClientFromEnv() (*Client, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, err
	}
	return NewClient(cfg)
}

// NewClient return a *etcd.Client perhaps need auth or tls
func NewClient(cfg Config) (*Client, error) {
	var (
		c    client.Client
		kapi client.KeysAPI
		err  error
	)

	var (
		transport = &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 10 * time.Second,
		}
		tlsCfg = &tls.Config{
			InsecureSkipVerify: false,
		}
		ecfg = client.Config{
			Endpoints:               cfg.Clusters,
			HeaderTimeoutPerRequest: 3 * time.Second,
		}
	)

	if cfg.Username != "" && cfg.Password != "" {
		ecfg.Username = cfg.Username
		ecfg.Password = cfg.Password
	}

	if cfg.CAcert != "" {
		cert, err := ioutil.ReadFile(cfg.CAcert)
		if err != nil {
			return nil, err
		}

		certPool := x509.NewCertPool()
		ok := certPool.AppendCertsFromPEM(cert)

		if ok {
			tlsCfg.RootCAs = certPool
		}
	}

	if cfg.Cert != "" && cfg.Key != "" {
		certificate, err := tls.LoadX509KeyPair(cfg.Cert, cfg.Key)
		if err != nil {
			return nil, err
		}
		tlsCfg.Certificates = []tls.Certificate{certificate}
	}

	transport.TLSClientConfig = tlsCfg
	ecfg.Transport = transport

	c, err = client.New(ecfg)
	if err != nil {
		return nil, err
	}

	kapi = client.NewKeysAPI(c)
	return &Client{confPath: cfg.ConfPath, client: kapi}, nil

}

// watch the confPath changes from etcd
// error will be ignored
func (c *Client) watch(ctx context.Context, respChan chan<- *client.Response) error {
	for {
		// set AfterIndex to 0 means watcher watch events begin at newest index
		// set Recursive to false means that the key must be exsited and not be a dir
		watcher := c.client.Watcher(c.confPath, &client.WatcherOptions{Recursive: false, AfterIndex: 0})
		resp, err := watcher.Next(ctx)
		if err != nil {
			// perhaps some terrible error happened
			// caller need recall WatchConfig
			return err
		}

		if resp.Node.Dir {
			// do not care about directory
			return ErrorUnexpectedDir
		}

		select {
		// if context canced then stop watch
		case <-ctx.Done():
			return ctx.Err()
		case respChan <- resp:
		}
	}
	return nil
}

// Get the latest value of key by Quorum = true
func (c *Client) get(ctx context.Context) (*client.Response, error) {
	resp, err := c.client.Get(ctx, c.confPath, &client.GetOptions{
		Recursive: false,
		Quorum:    true,
	})

	if err != nil {
		return nil, err
	}

	if resp.Node.Dir {
		return nil, ErrorUnexpectedDir
	}
	return resp, nil
}

// LoadConfig get value from etcd backend by the confPath
func (c *Client) LoadConfig(ctx context.Context, config interface{}) error {
	resp, err := c.get(ctx)
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(resp.Node.Value), config)
}

// WatchConfig initialize the config from etcd with confPath firstly
// Then watch the changes of the confPath and reassign config
// Call reload when success
func (c *Client) WatchConfig(ctx context.Context, config interface{}, reload confl.Reloader, errChan chan<- error) error {
	err := c.LoadConfig(ctx, config)
	if err != nil {
		return err
	}

	respChan := make(chan *client.Response)

	noWaitErrChan := func(errChan chan<- error, err error) {
		if err != nil {
			select {
			case errChan <- err:
			default:
				// if errChan is full then miss this err
			}
		}
	}

	// watch the key changes
	go func() {
		err := c.watch(ctx, respChan)
		noWaitErrChan(errChan, err)
		close(respChan)
	}()

	for range respChan {
		err := c.LoadConfig(ctx, config)
		noWaitErrChan(errChan, err)
		if err == nil {
			err = reload.Reload()
			noWaitErrChan(errChan, err)
		}

	}
	return nil
}
