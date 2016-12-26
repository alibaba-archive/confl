package sources

type Config struct {
	Type         Type
	AuthToken    string
	AuthType     string
	BasicAuth    bool
	ClientCaKeys string
	ClientCert   string
	ClientKey    string
	Hosts        []string
	Password     string
	Scheme       string
	Table        string
	Username     string
	AppID        string
	UserID       string
}
