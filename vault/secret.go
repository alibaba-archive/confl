package vault

import (
	"errors"
	"fmt"
	"strings"
)

// the value of vault type in json string must be like "VAULT(secret/xxx)"
// "xxx" means the key name in vault

const (
	vaultPrefix  = "VAULT("
	secretPrefix = "VAULT(secret/"
	secretSuffix = ")"
)

// secretKey checks the format of input value whether like "VAULT(secret/xxx)" or not
// return the content in parentheses like "secret/password"
func secretKey(in string) (string, error) {
	if !strings.HasPrefix(in, secretPrefix) {
		return "", fmt.Errorf("vault secret key(%s) has no prefix(%s)", in, secretPrefix)
	}
	tmp := strings.TrimPrefix(in, vaultPrefix)
	return strings.TrimSuffix(tmp, secretSuffix), nil
}

// Secret is a vault secret type. Some else like auth type and audit type
// It is used to store the key that needs to be encrypted
type Secret struct {
	key, Value string
}

// UnmarshalText implement the json.Unmarshaler interface
// This will be executed when call the json.Unmarshal(data []byte, v interface{}) on this type
func (s *Secret) UnmarshalText(b []byte) error {
	if defaultClient == nil {
		return errors.New("need initialize vault client")
	}

	var (
		tmp Secret
		err error
	)

	tmp.key, err = secretKey(string(b))
	if err != nil {
		return err
	}

	tmp.Value, err = defaultClient.key(tmp.key)
	if err != nil {
		return err
	}

	defaultClient.addKV(tmp.key, tmp.Value)
	*s = tmp
	return nil
}
