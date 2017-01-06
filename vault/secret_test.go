package vault

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSecretKey(t *testing.T) {
	assert := assert.New(t)
	keys := []string{
		"xxx",
		"VAULT(secret/hello)",
		"VAULT(auth/hello)",
		" VAULT(auth/hello) ",
	}
	key, err := secretKey(keys[0])
	assert.NotNil(err)
	key, err = secretKey(keys[1])
	assert.Nil(err)
	assert.Equal("secret/hello", key)
	key, err = secretKey(keys[2])
	assert.NotNil(err)
	key, err = secretKey(keys[3])
	assert.NotNil(err)
}
