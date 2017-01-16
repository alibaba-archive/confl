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
	key, value string
}

func (s *Secret) Value() string {
	return s.value
}

// UnmarshalJSON implement the json.Unmarshaler interface
// This will be executed when call the json.Unmarshal(data []byte, v interface{}) on this type
func (s *Secret) UnmarshalJSON(b []byte) error {
	if defaultClient == nil {
		return errors.New("need initialize vault client")
	}

	if s == nil {
		s = &Secret{}
	}

	var err error
	// the json string like `"xxx"` so need remove the double quotes
	tmp := strings.Trim(string(b), `"`)
	s.key, err = secretKey(tmp)
	if err != nil {
		return err
	}

	s.value, err = defaultClient.key(s.key)
	if err != nil {
		return err
	}

	defaultClient.addKV(s.key, s.value)
	return nil
}
