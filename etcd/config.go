package etcd

import (
	"github.com/teambition/confl"
)

type Config struct {
	confl.BaseConf
	Clusters []string `envconfig:"CONFL_ETCD_CLUSTERS" required:"true"`
	Cert     string   `envconfig:"CONFL_ETCD_CERT"`
	Key      string   `envconfig:"CONFL_ETCD_KEY"`
	CAcert   string   `envconfig:"CONFL_ETCD_CACERT"`
	Username string   `envconfig:"CONFL_ETCD_USERNAME"`
	Password string   `envconfig:"CONFL_ETCD_PASSWORD"`
}
