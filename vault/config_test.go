package vault

import (
	"os"
	"testing"

	"github.com/kelseyhightower/envconfig"
	"github.com/stretchr/testify/assert"
)

func TestConfigEnv(t *testing.T) {
	assert := assert.New(t)
	var cfg Config
	authType := "token"
	address := "http://localhost:8200"
	appID := "1"
	userID := "2"
	username := "3"
	password := "4"
	token := "5"
	cert := "6"
	key := "7"
	cacert := "8"
	os.Clearenv()
	os.Setenv("CONFL_VAULT_AUTH_TYPE", authType)
	os.Setenv("CONFL_VAULT_ADDRESS", address)
	os.Setenv("CONFL_VAULT_APP_ID", appID)
	os.Setenv("CONFL_VAULT_USER_ID", userID)
	os.Setenv("CONFL_VAULT_USERNAME", username)
	os.Setenv("CONFL_VAULT_PASSWORD", password)
	os.Setenv("CONFL_VAULT_TOKEN", token)
	os.Setenv("CONFL_VAULT_CERT", cert)
	os.Setenv("CONFL_VAULT_KEY", key)
	os.Setenv("CONFL_VAULT_CACERT", cacert)
	err := envconfig.Process("", &cfg)
	assert.Nil(err)
	assert.Equal(AuthToken, AuthType(authType))
	assert.Equal(AuthToken, cfg.AuthType)
	assert.Equal(address, cfg.Address)
	assert.Equal(appID, cfg.AppID)
	assert.Equal(userID, cfg.UserID)
	assert.Equal(username, cfg.Username)
	assert.Equal(password, cfg.Password)
	assert.Equal(token, cfg.Token)
	assert.Equal(cert, cfg.Cert)
	assert.Equal(key, cfg.Key)
	assert.Equal(cacert, cfg.CAcert)
}
