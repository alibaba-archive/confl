package vault

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/vault/http"
	"github.com/hashicorp/vault/vault"
	"github.com/stretchr/testify/assert"
)

func TestClient(t *testing.T) {
	core, _, token := vault.TestCoreUnsealed(t)
	ln, addr := http.TestServer(t, core)
	defer ln.Close()

	assert := assert.New(t)
	cfg := Config{
		AuthType: AuthToken,
		Token:    token,
		Address:  addr,
	}
	v, err := NewClient(cfg)
	assert.Nil(err)

	type config struct {
		Password *Secret `json:"password"`
	}

	c := config{
		Password: v.Secret(),
	}
	password := "123456"

	_, err = v.Client.Logical().Write("secret/password", map[string]interface{}{"value": password})
	assert.Nil(err)
	err = json.Unmarshal([]byte(`{"password": "VAULT(secret/password)"}`), &c)
	assert.Nil(err)
	assert.Equal(password, c.Password.Value)
}
