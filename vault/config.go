package vault

type AuthType int

const (
	AuthNone AuthType = iota
	AuthAppID
	AuthToken
	AuthGithub
	AuthPass
)

type Config struct {
	AuthType AuthType
	Address  string
	AppID    string
	UserID   string
	Username string
	Password string
	Token    string
	// x509 key pairs
	Cert string
	Key  string
	// CAcert pem
	CAcert string
}
