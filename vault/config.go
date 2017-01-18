package vault

type AuthType string

const (
	None   AuthType = ""
	AppID  AuthType = "app-id"
	Token  AuthType = "token"
	Github AuthType = "github"
	Pass   AuthType = "userpass"
)

// use https://github.com/kelseyhightower/envconfig to parse the environment variables for config
// it's container-friendly like docker, rocket
type Config struct {
	AuthType AuthType `envconfig:"CONFL_VAULT_AUTH_TYPE" required:"true"`
	Address  string   `envconfig:"CONFL_VAULT_ADDRESS" required:"true"`
	// "app id auth backend"(See https://www.vaultproject.io/docs/auth/app-id.html)
	AppID  string `envconfig:"CONFL_VAULT_APP_ID"`
	UserID string `envconfig:"CONFL_VAULT_USER_ID"`
	// "userpass auth backend"(see https://www.vaultproject.io/docs/auth/userpass.html)
	Username string `envconfig:"CONFL_VAULT_USERNAME"`
	Password string `envconfig:"CONFL_VAULT_PASSWORD"`
	// (See https://www.vaultproject.io/docs/auth/token.html)
	Token string `envconfig:"CONFL_VAULT_TOKEN"`
	// x509 key pairs
	Cert string `envconfig:"CONFL_VAULT_CERT"`
	Key  string `envconfig:"CONFL_VAULT_KEY"`
	// CAcert pem
	CAcert string `envconfig:"CONFL_VAULT_CACERT"`
	// watch interval
	// env vaule like `10s` `1m` `1h`
	Interval string `envconfig:"CONFL_VAULT_INTERVAL"`
}
