package etcd

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/coreos/etcd/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	cfg := &Config{
		Clusters: []string{"http://127.0.0.1:2379"},
	}
	// local etcd server without tls and basic auth
	cl, err := NewClient(cfg)
	require.Nil(err)

	t.Run("Key", func(t *testing.T) {
		key := "/confl/test1/key"
		values := []string{
			"test1",
			"test2",
			"test3",
		}
		for _, value := range values {
			_, err := cl.client.Set(context.Background(), key, value, &client.SetOptions{})
			require.Nil(err)
			v, err := cl.Key(key)
			require.Nil(err)
			assert.Equal(value, v)
		}
	})

	t.Run("WatchKey", func(t *testing.T) {
		type config struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}
		key := "/confl/test1/watchkey"
		values := []config{
			config{Name: "confl1", Age: 1},
			config{Name: "confl2", Age: 2},
			config{Name: "confl3", Age: 3},
		}

		changeCh := make(chan struct{})
		valueCh := make(chan string)
		doneCh := make(chan struct{})
		finishCh := make(chan struct{})
		go func() {
			cl.WatchKey(key, changeCh)
			finishCh <- struct{}{}
		}()

		go func() {
			for range changeCh {
				v := <-valueCh
				value, err := cl.Key(key)
				require.Nil(err)
				assert.Equal(v, value)
				doneCh <- struct{}{}
			}
		}()

		for _, value := range values {
			data, err := json.Marshal(value)
			require.Nil(err)
			v := string(data)
			_, err = cl.client.Set(context.Background(), key, v, &client.SetOptions{})
			require.Nil(err)
			valueCh <- v
			<-doneCh
		}

		t.Run("close", func(t *testing.T) {
			cl.Close()
			assert.Equal(struct{}{}, <-finishCh)
		})
	})

}
