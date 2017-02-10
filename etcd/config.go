package etcd

// Config etcd configuration
type Config struct {
	// cluseter addresses
	Clusters []string
	// security connection
	// the path of certificate file
	Cert string
	// the path of certificate'key file
	Key string
	// the path of CACert file
	CAcert string
	// auth user/pass
	Username string
	Password string
}
