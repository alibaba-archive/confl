package vault

import (
	"encoding/json"
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

	t.Run("init", func(t *testing.T) {
		assert := assert.New(t)
		// init twice
		changeCh := make(chan struct{})
		cfg := &Config{
			AuthType: Token,
			Token:    token,
			Address:  addr,
		}
		err := Init(cfg, changeCh)
		err = Init(cfg, changeCh)
		assert.NotNil(err)

		// no auth type
		defaultClient = nil
		changeCh = make(chan struct{})
		cfg = &Config{
			Token:   token,
			Address: addr,
		}
		err = Init(cfg, changeCh)
		assert.NotNil(err)

		// no change channel
		defaultClient = nil
		cfg = &Config{
			AuthType: Token,
			Token:    token,
			Address:  addr,
		}
		err = Init(cfg, nil)
		assert.NotNil(err)

		// error secure transport
		defaultClient = nil
		cfg = &Config{
			AuthType: Token,
			Token:    token,
			Address:  addr,
			CAcert:   "/path/to/noting",
		}
		err = Init(cfg, changeCh)
		assert.NotNil(err)

		// error vault address
		defaultClient = nil
		changeCh = make(chan struct{})
		cfg = &Config{
			AuthType: Token,
			Token:    token,
			Address:  "xxxxxx",
		}
		err = Init(cfg, changeCh)
		assert.NotNil(err)

		// unknown auth type
		defaultClient = nil
		cfg = &Config{
			AuthType: AuthType("hello"),
			Address:  addr,
		}
		err = Init(cfg, changeCh)
		assert.NotNil(err)

		// auth error
		defaultClient = nil
		cfg = &Config{
			AuthType: Token,
			Token:    "error token",
			Address:  addr,
		}
		err = Init(cfg, changeCh)
		assert.NotNil(err)

		// interval test success
		defaultClient = nil
		cfg = &Config{
			AuthType: Token,
			Token:    token,
			Address:  addr,
			Interval: "10s",
		}
		err = Init(cfg, changeCh)
		assert.Equal(10*time.Second, defaultClient.interval)

		// interval test fail
		defaultClient = nil
		cfg = &Config{
			AuthType: Token,
			Token:    token,
			Address:  addr,
			Interval: "10x",
		}
		err = Init(cfg, changeCh)
		assert.NotNil(err)
	})

	t.Run("defaultClient", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)
		changeCh := make(chan struct{})
		defaultClient = nil
		cfg := &Config{
			AuthType: Token,
			Token:    token,
			Address:  addr,
		}
		err := Init(cfg, changeCh)
		require.Nil(err)

		t.Run("key", func(t *testing.T) {
			// wrong fmt key
			_, err = defaultClient.key("unknown")
			assert.NotNil(err)

			// unknown key
			_, err = defaultClient.key("secret/unknown")
			assert.NotNil(err)

			// not a string
			_, err = defaultClient.Client.Logical().Write("secret/password", map[string]interface{}{"value": 3})
			require.Nil(err)
			_, err = defaultClient.key("secret/password")
			assert.NotNil(err)
		})

		defaultClient.interval = time.Second

		t.Run("watch", func(t *testing.T) {
			key := "secret/password"
			defaultClient.addKV(key, "test1")
			finishCh := make(chan struct{})
			go func() {
				defaultClient.watch()
				finishCh <- struct{}{}
			}()
			_, err = defaultClient.Client.Logical().Write(key, map[string]interface{}{"value": "test2"})
			assert.Equal(struct{}{}, <-changeCh)
			require.Nil(err)

			t.Run("close", func(t *testing.T) {
				Close()
				require.Equal(struct{}{}, <-finishCh)
			})
		})

		t.Run("secret", func(t *testing.T) {
			type config struct {
				Password Secret `json:"password"`
			}

			c := config{}
			password := "123456"

			_, err = defaultClient.Client.Logical().Write("secret/password", map[string]interface{}{"value": password})
			require.Nil(err)
			err = json.Unmarshal([]byte(`{"password": "VAULT(secret/password)"}`), &c)
			require.Nil(err)
			assert.Equal(password, c.Password.Value)

			err = json.Unmarshal([]byte(`{"password": "VAULT(xxxxx/password)"}`), &c)
			assert.NotNil(err)

			err = json.Unmarshal([]byte(`{"password": "VAULT(secret/unknown)"}`), &c)
			assert.NotNil(err)
		})

		t.Run("watchError", func(t *testing.T) {
			key := "secret/unknown"
			defaultClient.addKV(key, "test1")
			finishCh := make(chan struct{})

			defaultClient.stopCh = make(chan struct{})
			defaultClient.onError = func(err error) {
				assert.NotNil(err)
				finishCh <- struct{}{}
			}
			go func() {
				defaultClient.watch()
			}()
			<-finishCh
		})
	})
}

func TestUninit(t *testing.T) {
	defaultClient = nil
	assert := assert.New(t)
	type config struct {
		Password *Secret `json:"password"`
	}

	c := config{}
	err := json.Unmarshal([]byte(`{"password": "VAULT(secret/password)"}`), &c)
	assert.NotNil(err)
}
