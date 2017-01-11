package etcd

import (
	"context"
	"testing"

	"encoding/json"

	"github.com/coreos/etcd/client"
	"github.com/stretchr/testify/assert"
)

func TestClientStruct(t *testing.T) {
	cfg := &Config{
		Clusters: []string{"http://localhost:2379"},
	}
	// local etcd server without tls and basic auth
	cl, err := NewClient(cfg)
	assert.Nil(t, err)

	t.Run("Key", func(t *testing.T) {
		assert := assert.New(t)
		key := "/confl/test1/key"
		values := []string{
			"test1",
			"test2",
			"test3",
		}
		for _, value := range values {
			_, err := cl.client.Set(context.Background(), key, value, &client.SetOptions{})
			assert.Nil(err)
			v, err := cl.Key(key)
			assert.Nil(err)
			assert.Equal(value, v)
		}
	})

	t.Run("watchNext", func(t *testing.T) {
		assert := assert.New(t)
		key := "/confl/test1/watchnext"
		values := []string{
			"test1",
			"test2",
			"test3",
		}
		valueCh := make(chan string)
		respCh := make(chan *client.Response)
		go func() {
			for {
				go func() {
					resp, err := cl.watchNext(key)
					if assert.Nil(err) {
						respCh <- resp
					}
				}()
				value := <-valueCh
				resp := <-respCh
				assert.Nil(err)
				assert.Equal(value, resp.Node.Value)
			}
		}()
		for _, value := range values {
			_, err := cl.client.Set(context.Background(), key, value, &client.SetOptions{})
			assert.Nil(err)
			valueCh <- value
		}
	})

	t.Run("WatchKey", func(t *testing.T) {
		assert := assert.New(t)
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
		go cl.WatchKey(key, changeCh)
		go func() {
			for range changeCh {
				v := <-valueCh
				value, err := cl.Key(key)
				assert.Nil(err)
				assert.Equal(v, value)
				doneCh <- struct{}{}
			}
		}()
		for _, value := range values {
			data, err := json.Marshal(value)
			assert.Nil(err)
			v := string(data)
			_, err = cl.client.Set(context.Background(), key, v, &client.SetOptions{})
			valueCh <- v
			assert.Nil(err)
			<-doneCh
		}
	})
}
