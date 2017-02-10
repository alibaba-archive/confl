package common

// your configuration struct
// now just support json unmarshal
type Config struct {
	Username string `json:"username"`
	// tag vault:"secret" is a secure type for store password, secret, token etc
	// it will load value from vault which is a tool for managing secrets
	Password string `json:"password" vault:"secret/password"`
	Token    string `json:"token" vault:"secret/token"`
}
