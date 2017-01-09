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
)

type Confl struct {
	confPath string
	client   client.KeysAPI
}

func NewConflFromEnv() (*Confl, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, err
	}
	return NewConfl(cfg)
}

// NewConfl return a *etcd.Client perhaps need auth or tls
func NewConfl(cfg Config) (*Confl, error) {
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
	return &Confl{confPath: cfg.ConfPath, client: kapi}, nil

}

// watch the confPath changes from etcd
// error will be ignored
func (c *Confl) watch(ctx context.Context, respChan chan<- *client.Response) error {
	defer close(respChan)
	for {
		// if context canced then stop watch
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// set AfterIndex to 0 means watcher watch events begin at newest index
		// set Recursive to false means that the key must be exsited and not be a dir
		watcher := c.client.Watcher(c.confPath, &client.WatcherOptions{Recursive: false, AfterIndex: 0})
		resp, err := watcher.Next(ctx)
		if err != nil {
			// if error happened need sleep before continue
			time.Sleep(time.Second)
			continue
		}

		if resp.Node.Dir {
			return ErrorUnexpectedDir
		}
		respChan <- resp
	}
	return nil
}

// Get the latest value of key by Quorum = true
func (c *Confl) get(ctx context.Context) (*client.Response, error) {
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
func (c *Confl) LoadConfig(ctx context.Context, config interface{}) error {
	resp, err := c.get(ctx)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(resp.Node.Value), config)
}

// WatchConfig initialize the config from etcd with confPath firstly
// Then watch the changes of the confPath and reassign config
// Call reload when success
func (c *Confl) WatchConfig(ctx context.Context, config interface{}, reload func() error) <-chan error {
	respChan := make(chan *client.Response)
	errChan := make(chan error)
	err := c.LoadConfig(ctx, config)
	if err != nil {
		errChan <- err
	}

	// watch the key changes
	go func() {
		err := c.watch(ctx, respChan)
		if err != nil {
			errChan <- err
		}
	}()

	go func() {
		for range respChan {
			err := c.LoadConfig(ctx, config)
			if err != nil {
				errChan <- err
				continue
			}
			if err = reload(); err != nil {
				errChan <- err
			}
		}
	}()
	return errChan
}
