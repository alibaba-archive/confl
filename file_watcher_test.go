package confl_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teambition/confl"
	"gopkg.in/yaml.v2"
)

func TestFileWatcher(t *testing.T) {
	assert := assert.New(t)
	watcher, err := confl.NewFileWatcher(&Config{}, "./config_test.json")
	assert.Nil(err)
	config := watcher.Config().(Config)
	assert.Equal("username", config.Username)
	assert.Equal("666", config.In.Pass)
}

func TestFileWatcherWithYaml(t *testing.T) {
	assert := assert.New(t)
	watcher, err := confl.NewFileWatcher(&Config{}, "./config_test.yaml", yaml.Unmarshal)
	assert.Nil(err)
	config := watcher.Config().(Config)
	assert.Equal("username", config.Username)
	assert.Equal("666", config.In.Pass)
}
