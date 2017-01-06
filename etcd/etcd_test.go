package etcd

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/stretchr/testify/assert"
	"github.com/teambition/confl"
)

func TestInterface(t *testing.T) {
	var cl1 = &Confl{}
	var cl2 confl.Confl
	assert.NotPanics(t, func() {
		cl2 = cl1
	})
}

func TestEtcdConfl(t *testing.T) {
	cfg := Config{
		Clusters: []string{"http://localhost:2379"},
	}
	// local etcd server without tls and basic auth
	cl, err := NewConfl(cfg)
	assert.Nil(t, err)

	t.Run("watch", func(t *testing.T) {
		assert := assert.New(t)
		key := "/confl/watch/test"
		values := []string{
			"test1",
			"test2",
			"test3",
		}
		valueChan := make(chan string)
		respChan := make(chan *client.Response)
		go func() {
			err := cl.watch(context.Background(), key, respChan)
			assert.Nil(err)
		}()
		go func() {
			for resp := range respChan {
				value := <-valueChan
				assert.Equal(value, resp.Node.Value)
			}
		}()
		for _, value := range values {
			_, err := cl.client.Set(context.Background(), key, value, &client.SetOptions{TTL: 10 * time.Second})
			assert.Nil(err)
			valueChan <- value
		}
	})

	t.Run("get", func(t *testing.T) {
		assert := assert.New(t)
		kvs := map[string]string{
			"/confl/get/test1": "test1",
			"/confl/get/test2": "test2",
			"/confl/get/test3": "test2",
		}
		for key, value := range kvs {
			_, err := cl.client.Set(context.Background(), key, value, &client.SetOptions{TTL: 10 * time.Second})
			assert.Nil(err)
			resp, err := cl.get(context.Background(), key)
			assert.Nil(err)
			assert.Equal(value, resp.Node.Value)
		}
	})

	t.Run("interface", func(t *testing.T) {
		type config struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}
		key := "/confl/load/test"

		t.Run("LoadConfig", func(t *testing.T) {
			assert := assert.New(t)
			value := `{"name": "confl", "age": 18}`
			_, err := cl.client.Set(context.Background(), key, value, &client.SetOptions{TTL: 10 * time.Second})
			assert.Nil(err)
			c := config{}
			err = cl.LoadConfig(context.Background(), &c, key)
			assert.Nil(err)
			assert.Equal("confl", c.Name)
			assert.Equal(18, c.Age)
		})

		t.Run("WatchConfig", func(t *testing.T) {
			assert := assert.New(t)
			values := []config{
				config{Name: "confl1", Age: 1},
				config{Name: "confl2", Age: 2},
				config{Name: "confl3", Age: 3},
			}
			c := config{}
			valueChan := make(chan config)
			cl.WatchConfig(context.Background(), &c, key, func() error {
				value := <-valueChan
				assert.Equal(value.Name, c.Name)
				assert.Equal(value.Age, c.Age)
				return nil
			})

			go func() {
				for _, value := range values {
					data, err := json.Marshal(value)
					assert.Nil(err)
					_, err = cl.client.Set(context.Background(), key, string(data), &client.SetOptions{TTL: 10 * time.Second})
					assert.Nil(err)
					valueChan <- value
				}
			}()
		})
	})
}
