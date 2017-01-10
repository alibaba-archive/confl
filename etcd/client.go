package etcd

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/kelseyhightower/envconfig"
)

type Client struct {
	client client.KeysAPI
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
	return &Client{client: kapi}, nil

}

func (c *Client) watchNext(ctx context.Context, key string) (*client.Response, error) {
	// set AfterIndex to 0 means watcher watch events begin at newest index
	// set Recursive to false means that the key must be exsited and not be a dir
	watcher := c.client.Watcher(key, &client.WatcherOptions{Recursive: false, AfterIndex: 0})

	resp, err := watcher.Next(ctx)
	if err != nil {
		// perhaps some terrible error happened
		// caller need recall WatchConfig
		return nil, err
	}

	if resp.Node.Dir {
		// do not care about directory
		return nil, ErrorUnexpectedDir
	}
	return resp, nil
}

// WatchKey the key changes from etcd until be stopped
func (c *Client) WatchKey(key string, reloadCh chan<- struct{}, stopCh <-chan struct{}, errCh chan<- error) {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		for {
			_, err := c.watchNext(ctx, key)
			if err != nil {
				errCh <- err
				if ctx.Err() != nil {
					// means context has canceled and stop watch
					return
				}

				time.Sleep(2 * time.Second)
				continue
			}

			reloadCh <- struct{}{}
		}
	}()

	select {
	// canceled when stop
	case <-stopCh:
		cancel()
		return
	}
}

// Key the latest value of key by Quorum = true
func (c *Client) Key(key string) (string, error) {
	resp, err := c.client.Get(context.Background(), key, &client.GetOptions{
		Recursive: false,
		Quorum:    true,
	})

	if err != nil {
		return "", err
	}

	if resp.Node.Dir {
		return "", ErrorUnexpectedDir
	}

	return resp.Node.Value, nil
}

func (c *Client) Close() error {
	return nil
}
