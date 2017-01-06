package etcd

import (
	"os"
	"testing"

	"github.com/kelseyhightower/envconfig"
	"github.com/stretchr/testify/assert"
)

func TestConfigEnv(t *testing.T) {
	assert := assert.New(t)
	var cfg Config
	clusters := []string{"http://node1.example.com:2379", "http://node2.example.com:2379", "http://node3.example.com:2379"}
	clustersStr := "http://node1.example.com:2379,http://node2.example.com:2379,http://node3.example.com:2379"
	cert := "1"
	key := "2"
	cacert := "3"
	username := "4"
	password := "5"
	os.Clearenv()
	os.Setenv("CONFL_ETCD_CLUSTERS", clustersStr)
	os.Setenv("CONFL_ETCD_CERT", cert)
	os.Setenv("CONFL_ETCD_KEY", key)
	os.Setenv("CONFL_ETCD_CACERT", cacert)
	os.Setenv("CONFL_ETCD_USERNAME", username)
	os.Setenv("CONFL_ETCD_PASSWORD", password)
	err := envconfig.Process("", &cfg)
	assert.Nil(err)
	assert.Equal(clusters, cfg.Clusters)
	assert.Equal(cert, cfg.Cert)
	assert.Equal(key, cfg.Key)
	assert.Equal(cacert, cfg.CAcert)
	assert.Equal(username, cfg.Username)
	assert.Equal(password, cfg.Password)
}
