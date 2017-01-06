package vault

import (
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
	Key, Value string
	vault      *Vault
}

// UnmarshalJSON implement the json.Unmarshaler interface
// This will be executed when call the json.Unmarshal(data []byte, v interface{}) on this type
func (s *Secret) UnmarshalJSON(b []byte) error {
	var err error
	// the json string like `"xxx"` so need remove the double quotes
	tmp := strings.Trim(string(b), `"`)
	s.Key, err = secretKey(tmp)
	if err != nil {
		return err
	}

	resp, err := s.vault.Logical().Read(s.Key)
	if err != nil {
		return err
	}

	if resp == nil || resp.Data == nil {
		return fmt.Errorf("vault secret key(%s) is not existed", s.Key)
	}

	if value, ok := resp.Data["value"]; ok {
		// value just can only be string type
		if text, ok := value.(string); ok {
			s.Value = text
			return nil
		}
	}

	return fmt.Errorf("vault secret key(%s) value needs a string type", s.Key)
}
