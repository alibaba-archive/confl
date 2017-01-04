package json

import (
	"fmt"
	"strings"

	"github.com/teambition/confl/sources"
	"golang.org/x/net/context"
)

const (
	VaultPrefix = "VAULT:"
)

// Vault parse encrypted secret from vault service
type Vault struct {
	client     sources.Client
	key, value string
}

// NewVault need a vault.Client
func NewVault(client sources.Client) *Vault {
	return &Vault{
		client: client,
	}
}

func (v *Vault) UnmarshalJSON(b []byte) (err error) {
	tmp := strings.Trim(string(b), `"`)
	if !strings.HasPrefix(tmp, VaultPrefix) {
		err = fmt.Errorf("value(%s) has no prefix(%s)", tmp, VaultPrefix)
		return
	}
	v.key = strings.TrimSpace(strings.TrimLeft(tmp, VaultPrefix))
	v.value, err = v.client.Key(context.Background(), v.key)
	return
}

func (v *Vault) Value() string {
	return v.value
}
