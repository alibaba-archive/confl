package vault

import (
	"fmt"
	"strings"
)

const (
	vaultPrefix  = "VAULT("
	secretPrefix = "VAULT(secret/"
	secretSuffix = ")"
)

func secretKey(in string) (string, error) {
	if !strings.HasPrefix(in, secretPrefix) {
		return "", fmt.Errorf("vault secret key(%s) has no prefix(%s)", in, secretPrefix)
	}
	tmp := strings.TrimPrefix(in, vaultPrefix)
	return strings.TrimSuffix(tmp, secretSuffix), nil
}

// Secret is a vault secret type. Perhaps be an auth type or an audit type
// Key like "VAULT: secret/password"
type Secret struct {
	Key, Value string
	vault      *Vault
}

func (s *Secret) UnmarshalJSON(b []byte) error {
	var err error
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
		if text, ok := value.(string); ok {
			s.Value = text
			return nil
		}
	}
	return fmt.Errorf("vault secret key(%s) value needs a string type", s.Key)
}
