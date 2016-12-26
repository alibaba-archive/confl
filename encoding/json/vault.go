package json

import (
	"errors"
	"fmt"
	"strings"

	"github.com/teambition/confl/sources"
	"golang.org/x/net/context"
)

const (
	VaultPrefix = "VAULT:"
)

type Vault struct {
	prefix     string
	client     sources.Client
	cancel     context.CancelFunc
	reload     func() error
	key, value string
}

func NewVault(prefix string, client sources.Client, reload func() error) (*Vault, error) {
	if client.Type() != sources.Vault {
		return nil, errors.New("error client type")
	}
	return &Vault{
		prefix: prefix,
		client: client,
		reload: reload,
	}, nil
}

func (v *Vault) Copy() *Vault {
	copyed := *v
	return &copyed
}

func (v *Vault) UnmarshalJSON(b []byte) (err error) {
	if v.cancel != nil {
		v.cancel()
	}
	tmp := strings.Trim(string(b), `"`)
	if !strings.HasPrefix(tmp, VaultPrefix) {
		err = fmt.Errorf("value(%s) has no prefix(%s)", tmp, VaultPrefix)
		return
	}
	v.key = strings.TrimSpace(strings.TrimLeft(tmp, VaultPrefix))
	splitKey := strings.Split(v.key, "/")
	finalKey := splitKey[len(splitKey)-1]
	if err = v.update(); err != nil {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	v.cancel = cancel
	go func() {
		var err error
		for resp := range v.client.Watch(ctx, v.prefix) {
			if resp.Error != nil {
				fmt.Println(resp.Error)
				continue
			}
			if !strings.HasSuffix(resp.Key, finalKey) {
				continue
			}
			if err = v.update(); err != nil {
				fmt.Println(resp.Error)
				continue
			}
			if err = v.reload(); err != nil {
				fmt.Println(err)
			}
		}
	}()
	return
}

func (v *Vault) update() (err error) {
	v.value, err = v.client.Key(context.Background(), v.key)
	return
}

func (v *Vault) Value() string {
	return v.value
}
