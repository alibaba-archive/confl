package sources

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"fmt"

	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

type EtcdClient struct {
	client client.KeysAPI
}

func NewEtcdClient(hosts []string, cert, key, caCert string, basicAuth bool, username string, password string) (*EtcdClient, error) {
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
	return &EtcdClient{kapi}, nil

}

func (c *EtcdClient) WithClient(client Client) {}

func (c *EtcdClient) Key(ctx context.Context, key string) (string, error) {
	resp, err := c.client.Get(ctx, key, &client.GetOptions{
		Recursive: true,
		Sort:      true,
		Quorum:    true,
	})
	if err != nil {
		return "", err
	}
	if resp.Node.Dir {
		return "", fmt.Errorf("key(%s) is a directory", key)
	}
	return resp.Node.Value, nil
}

func (c *EtcdClient) Watch(ctx context.Context, key string) <-chan *Response {
	respChan := make(chan *Response)
	go func() {
		defer close(respChan)
		// set AfterIndex to 0 means watcher watch events begin at newest index
		watcher := c.client.Watcher(key, &client.WatcherOptions{Recursive: true})
		for {
			if ctx.Err() != nil {
				// means context has canceled
				break
			}
			resp, err := watcher.Next(ctx)
			if err != nil {
				respChan <- &Response{
					Error: err,
				}
				switch err.(type) {
				case *client.Error:
				default:
					// if network error happened need sleep before contact
					time.Sleep(1 * time.Second)
				}
				continue
			}

			if !resp.Node.Dir {
				respChan <- &Response{
					Key:       resp.Node.Key,
					NextIndex: resp.Node.ModifiedIndex,
				}
			}
		}
	}()

	return respChan
}

func (c *EtcdClient) Type() Type {
	return Etcd
}
