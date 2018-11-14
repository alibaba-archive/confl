package confl_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teambition/confl"
)

func TestMockWatcher(t *testing.T) {
	assert := assert.New(t)
	watcher, err := confl.NewMockWatcher(&Config{})
	assert.Nil(err)
	config := watcher.Config().(Config)
	assert.Equal("", config.Username)
	assert.Equal("", config.In.Pass)
}
