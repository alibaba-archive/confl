package etcd

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/teambition/confl/sources/common"
	"golang.org/x/net/context"
)

type Client struct {
	client client.KeysAPI
}

// NewClient return a *etcd.Client perhaps need auth or tls
func NewClient(hosts []string, cert, key, caCert string, basicAuth bool, username string, password string) (*Client, error) {
	var (
		c    client.Client
		kapi client.KeysAPI
		err  error
	)

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
	}
	cfg := client.Config{
		Endpoints:               hosts,
		HeaderTimeoutPerRequest: 3 * time.Second,
	}

	if basicAuth {
		cfg.Username = username
		cfg.Password = password
	}

	if caCert != "" {
		certBytes, err := ioutil.ReadFile(caCert)
		if err != nil {
			return nil, err
		}

		caCertPool := x509.NewCertPool()
		ok := caCertPool.AppendCertsFromPEM(certBytes)

		if ok {
			tlsConfig.RootCAs = caCertPool
		}
	}

	if cert != "" && key != "" {
		tlsCert, err := tls.LoadX509KeyPair(cert, key)
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{tlsCert}
	}

	transport.TLSClientConfig = tlsConfig
	cfg.Transport = transport

	c, err = client.New(cfg)
	if err != nil {
		return nil, err
	}

	kapi = client.NewKeysAPI(c)
	return &Client{kapi}, nil

}

// Key the latest value of key by Quorum = true
func (c *Client) Key(ctx context.Context, key string) (string, error) {
	resp, err := c.client.Get(ctx, key, &client.GetOptions{
		Recursive: false,
		Quorum:    true,
	})
	if err != nil {
		return "", err
	}
	// must be a key not a directory
	if resp.Node.Dir {
		return "", common.ErrorUnexpectDir
	}
	return resp.Node.Value, nil
}

func (c *Client) Watch(ctx context.Context, key string) <-chan *common.Response {
	respChan := make(chan *common.Response)
	go func() {
		defer close(respChan)
		for {
			select {
			case <-ctx.Done():
				respChan <- &common.Response{
					Error: ctx.Err(),
				}
				return
			default:
			}
			// set AfterIndex to 0 means watcher watch events begin at newest index
			// set Recursive to false means that the key must be exsited and not be a dir
			watcher := c.client.Watcher(key, &client.WatcherOptions{Recursive: false, AfterIndex: 0})
			resp, err := watcher.Next(ctx)
			if err != nil {
				respChan <- &common.Response{
					Error: err,
				}
				// if error happened need sleep before continue
				time.Sleep(time.Second)
				continue
			}

			if resp.Node.Dir {
				respChan <- &common.Response{
					Error: common.ErrorUnexpectDir,
				}
				return
			}

			respChan <- &common.Response{
				Key:       resp.Node.Key,
				Value:     resp.Node.Value,
				NextIndex: resp.Node.ModifiedIndex,
			}
		}
	}()
	return respChan
}
