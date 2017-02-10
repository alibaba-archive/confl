package etcd

import (
	"context"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/teambition/confl/util"
)

type Client struct {
	client    client.KeysAPI
	ctx       context.Context
	cancel    context.CancelFunc
	errHandle func(err error)
}

// NewClient return a *etcd.Client perhaps need auth or tls
func NewClient(cfg Config, errHandle ...func(err error)) (*Client, error) {
	var (
		c    client.Client
		kapi client.KeysAPI
		err  error
	)

	ecfg := client.Config{
		Endpoints:               cfg.Clusters,
		HeaderTimeoutPerRequest: 3 * time.Second,
	}

	if cfg.Username != "" && cfg.Password != "" {
		ecfg.Username = cfg.Username
		ecfg.Password = cfg.Password
	}

	ecfg.Transport, err = util.SecureTransport(cfg.CAcert, cfg.Cert, cfg.Key)
	if err != nil {
		return nil, err
	}

	c, err = client.New(ecfg)
	if err != nil {
		return nil, err
	}

	kapi = client.NewKeysAPI(c)
	ctx, cancel := context.WithCancel(context.Background())

	ec := &Client{
		client: kapi,
		ctx:    ctx,
		cancel: cancel,
	}

	if len(errHandle) == 1 {
		ec.errHandle = errHandle[0]
	}
	return ec, nil
}

// WatchKey the key changes from etcd until be stopped
func (c *Client) WatchKey(key string, changeCh chan<- struct{}) {
	for {
		// set AfterIndex to 0 means watcher watch events begin at newest index
		// set Recursive to false means that the key must be exsited and not be a dir
		watcher := c.client.Watcher(key, &client.WatcherOptions{Recursive: false, AfterIndex: 0})
		_, err := watcher.Next(c.ctx)
		if err != nil {
			if c.errHandle != nil {
				c.errHandle(err)
			}
			if c.ctx.Err() != nil {
				// means context has be canceled and stop watch
				return
			}
			// unexpected error happended
			time.Sleep(2 * time.Second)
			continue
		}

		changeCh <- struct{}{}
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
	c.cancel()
	return nil
}
