package etcd

type Config struct {
	Clusters []string `shot:"ecs" long:"etcd-clusters" description:"etcd clusters"`
	Cert     string   `shot:"ecert" long:"etcd-cert" description:"etcd cert"`
	Key      string   `shot:"ekey" long:"etcd-key" description:"etcd key"`
	CAcert   string   `shot:"eca" long:"etcd-cacert" description:"etcd cacert"`
	Username string   `shot:"ename" long:"etcd-username" description:"etcd username"`
	Password string   `shot:"epass" long:"etcd-password" description:"etcd password"`
}
