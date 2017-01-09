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
	var cl1 = &Client{}
	var cl2 confl.Confl
	assert.NotPanics(t, func() {
		cl2 = cl1
	})
}

func TestClientStruct(t *testing.T) {
	cfg := Config{
		Clusters: []string{"http://localhost:2379"},
	}
	cfg.ConfPath = "/confl/test"
	// local etcd server without tls and basic auth
	cl, err := NewClient(cfg)
	assert.Nil(t, err)

	t.Run("watch", func(t *testing.T) {
		assert := assert.New(t)
		values := []string{
			"test1",
			"test2",
			"test3",
		}
		valueChan := make(chan string)
		respChan := make(chan *client.Response)
		go func() {
			err := cl.watch(context.Background(), respChan)
			assert.Nil(err)
		}()
		go func() {
			for resp := range respChan {
				value := <-valueChan
				assert.Equal(value, resp.Node.Value)
			}
		}()
		// wait until watch get ready
		time.Sleep(time.Second)
		for _, value := range values {
			_, err := cl.client.Set(context.Background(), cfg.ConfPath, value, &client.SetOptions{TTL: 10 * time.Second})
			assert.Nil(err)
			valueChan <- value
		}
	})

	t.Run("get", func(t *testing.T) {
		assert := assert.New(t)
		values := []string{
			"test1",
			"test2",
			"test3",
		}
		for _, value := range values {
			_, err := cl.client.Set(context.Background(), cfg.ConfPath, value, &client.SetOptions{TTL: 10 * time.Second})
			assert.Nil(err)
			resp, err := cl.get(context.Background())
			assert.Nil(err)
			assert.Equal(value, resp.Node.Value)
		}
	})

	t.Run("interface", func(t *testing.T) {
		type config struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}

		t.Run("LoadConfig", func(t *testing.T) {
			assert := assert.New(t)
			value := `{"name": "confl", "age": 18}`
			_, err := cl.client.Set(context.Background(), cfg.ConfPath, value, &client.SetOptions{TTL: 10 * time.Second})
			assert.Nil(err)
			c := config{}
			err = cl.LoadConfig(context.Background(), &c)
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
			errChan := make(chan error)

			go func() {
				err := cl.WatchConfig(context.Background(), &c, confl.ReloadFunc(func() error {
					value := <-valueChan
					assert.Equal(value.Name, c.Name)
					assert.Equal(value.Age, c.Age)
					return nil
				}), errChan)
				assert.Nil(err)
			}()

			// wait until watch get ready
			time.Sleep(time.Second)
			for _, value := range values {
				data, err := json.Marshal(value)
				assert.Nil(err)
				_, err = cl.client.Set(context.Background(), cfg.ConfPath, string(data), &client.SetOptions{TTL: 10 * time.Second})
				assert.Nil(err)
				valueChan <- value
			}

		})
	})
}
