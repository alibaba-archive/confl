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
	// type of auth one in (app-id, token, github, userpass)
	AuthType string
	// vault service address
	Address string
	// AuthType = app-id
	// "app id auth backend"(See https://www.vaultproject.io/docs/auth/app-id.html)
	// this is more useful for micro services
	// every micro service can be given a app_id to distinguish between identities
	AppID  string
	UserID string
	// AuthType = userpass
	// "userpass auth backend"(see https://www.vaultproject.io/docs/auth/userpass.html)
	Username string
	Password string
	// AuthType = token or github
	// auth token (See https://www.vaultproject.io/docs/auth/token.html)
	Token string

	// security connection
	// Cert and Key are the pair of x509
	// the path of certificate file
	Cert string
	// the path of certificate'key file
	Key string
	// the path of CACert file
	CAcert string
	// loop interval
	// vaule likes `10s` `1m` `1h`
	Interval string
}
