package confl

import (
	"encoding/json"
	"fmt"

	"github.com/teambition/confl/sources"
	"golang.org/x/net/context"
)

// LoadJSONAndWatch load the config from client by the key and watch the key
// call reload when the key changed
func LoadJSONAndWatch(ctx context.Context, config interface{}, key string, client sources.Client, reload func() error) (err error) {
	if err = loadJSON(config, key, client); err != nil {
		return
	}
	for resp := range client.Watch(ctx, key) {
		if resp.Error != nil {
			fmt.Println(resp.Error)
			continue
		}
		if err = loadJSON(config, key, client); err != nil {
			fmt.Println(resp.Error)
			continue
		}
		if err = reload(); err != nil {
			fmt.Println(err)
		}
	}
	return
}

func loadJSON(config interface{}, key string, client sources.Client) (err error) {
	var value string
	if value, err = client.Key(context.Background(), key); err != nil {
		return
	}
	err = json.Unmarshal([]byte(value), config)
	return
}
