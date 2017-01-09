package confl

// BaseConf each backend config needs inherit this struct
// Example:
// type Config struct {
// 	   confl.BaseConf
// 	   Clusters []string `envconfig:"CONFL_ETCD_CLUSTERS" required:"true"`
// 	   Cert     string   `envconfig:"CONFL_ETCD_CERT"`
// 	   Key      string   `envconfig:"CONFL_ETCD_KEY"`
// 	   CAcert   string   `envconfig:"CONFL_ETCD_CACERT"`
// 	   Username string   `envconfig:"CONFL_ETCD_USERNAME"`
// 	   Password string   `envconfig:"CONFL_ETCD_PASSWORD"`
// }
type BaseConf struct {
	// the path of config file
	ConfPath string `envconfig:"CONFL_CONF_PATH" required:"true"`
}
