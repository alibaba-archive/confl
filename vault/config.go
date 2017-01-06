package vault

type AuthType string

const (
	AuthNone   AuthType = ""
	AuthAppID  AuthType = "app-id"
	AuthToken  AuthType = "token"
	AuthGithub AuthType = "github"
	AuthPass   AuthType = "userpass"
)

type Config struct {
	AuthType AuthType `envconfig:"CONFL_VAULT_AUTH_TYPE" required:"true"`
	Address  string   `envconfig:"CONFL_VAULT_ADDRESS" required:"true"`
	AppID    string   `envconfig:"CONFL_VAULT_APP_ID"`
	UserID   string   `envconfig:"CONFL_VAULT_USER_ID"`
	Username string   `envconfig:"CONFL_VAULT_USERNAME"`
	Password string   `envconfig:"CONFL_VAULT_PASSWORD"`
	Token    string   `envconfig:"CONFL_VAULT_TOKEN"`
	// x509 key pairs
	Cert string `envconfig:"CONFL_VAULT_CERT"`
	Key  string `envconfig:"CONFL_VAULT_KEY"`
	// CAcert pem
	CAcert string `envconfig:"CONFL_VAULT_CACERT"`
}
