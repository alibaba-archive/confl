package vault

import (
	"testing"
	"time"

	"github.com/hashicorp/vault/http"
	"github.com/hashicorp/vault/vault"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVault(t *testing.T) {
	core, _, token := vault.TestCoreUnsealed(t)
	ln, addr := http.TestServer(t, core)
	defer ln.Close()

	authType := "token"
	t.Run("New", func(t *testing.T) {
		assert := assert.New(t)
		changeCh := make(chan struct{})

		// no change channel
		cfg := Config{
			AuthType: authType,
			Token:    token,
			Address:  addr,
		}
		_, err := New(cfg, nil)
		assert.NotNil(err)

		// no auth type
		cfg = Config{
			Token:   token,
			Address: addr,
		}
		_, err = New(cfg, changeCh)
		assert.NotNil(err)

		// error secure transport
		cfg = Config{
			AuthType: authType,
			Token:    token,
			Address:  addr,
			CAcert:   "/path/to/noting",
		}
		_, err = New(cfg, changeCh)
		assert.NotNil(err)

		// error vault address
		cfg = Config{
			AuthType: authType,
			Token:    token,
			Address:  "xxxxxx",
		}
		_, err = New(cfg, changeCh)
		assert.NotNil(err)

		// unknown auth type
		cfg = Config{
			AuthType: "hello",
			Address:  addr,
		}
		_, err = New(cfg, changeCh)
		assert.NotNil(err)

		// auth error
		cfg = Config{
			AuthType: authType,
			Token:    "error token",
			Address:  addr,
		}
		_, err = New(cfg, changeCh)
		assert.NotNil(err)

		// interval test success
		cfg = Config{
			AuthType: authType,
			Token:    token,
			Address:  addr,
			Interval: "10s",
		}
		cl, err := New(cfg, changeCh)
		assert.Equal(10*time.Second, cl.interval)

		// interval test fail
		cfg = Config{
			AuthType: authType,
			Token:    token,
			Address:  addr,
			Interval: "10x",
		}
		_, err = New(cfg, changeCh)
		assert.NotNil(err)
	})

	t.Run("methods", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)
		changeCh := make(chan struct{})
		cfg := Config{
			AuthType: authType,
			Token:    token,
			Address:  addr,
			Interval: "1s",
		}
		client, err := New(cfg, changeCh)
		require.Nil(err)

		t.Run("key", func(t *testing.T) {
			// wrong fmt key
			_, err = client.key("unknown")
			assert.NotNil(err)

			// unknown key
			_, err = client.key("secret/unknown")
			assert.NotNil(err)

			// not a string
			_, err = client.Client.Logical().Write("secret/password", map[string]interface{}{"value": 3})
			require.Nil(err)
			_, err = client.key("secret/password")
			assert.NotNil(err)

			// success
			_, err = client.Client.Logical().Write("secret/password", map[string]interface{}{"value": "hello"})
			require.Nil(err)
			value, _ := client.key("secret/password")
			assert.Equal("hello", value)
		})

		t.Run("watch", func(t *testing.T) {
			key := "secret/password"
			client.addKV(key, "test1")
			_, err = client.Client.Logical().Write(key, map[string]interface{}{"value": "test2"})
			require.Nil(err)
			assert.Equal(struct{}{}, <-changeCh)

			// unknown k/v
			client.errHandle = func(err error) {
				assert.NotNil(err)
			}
			key2 := "secret/unknown"
			client.addKV(key2, "test1")
		})

		t.Run("scan", func(t *testing.T) {
			type config struct {
				Password string `json:"password" vault:"secret/password"`
			}

			key, value := "secret/password", "123456"
			c := &config{Password: "xxxxxx"}

			_, err = client.Client.Logical().Write(key, map[string]interface{}{"value": value})
			require.Nil(err)
			// scan strcut for vault tag
			err = client.Scan(c)
			assert.Nil(err)
			assert.Equal(value, c.Password)
		})

		t.Run("close", func(t *testing.T) {
			client.Close()
			assert.Panics(func() {
				close(client.stopCh)
			})
		})

	})
}
