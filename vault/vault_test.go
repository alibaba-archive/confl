package vault

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVault(t *testing.T) {
	assert := assert.New(t)
	cfg := Config{
		AuthType: AuthToken,
		Token:    "273e8292-e45d-7da1-2560-3118adbe01c0",
		Address:  "http://localhost:8200",
	}
	v, err := New(cfg)
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
