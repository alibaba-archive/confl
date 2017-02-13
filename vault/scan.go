package vault

import (
	"errors"
	"reflect"
)

const (
	vaultTag = "vault"
)

// Scan scan the struct for tag `vault`
// replace the value from vault service
func (c *Client) Scan(v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() || rv.Elem().Kind() != reflect.Struct {
		return errors.New("need a struct pointer")
	}
	return c.scanStruct(rv.Elem())
}

// scanStruct iter the struct for lookup tag `vault`
func (c *Client) scanStruct(rv reflect.Value) (err error) {
	for f, fs := 0, rv.NumField(); f < fs; f++ {
		val := rv.Field(f)
		switch val.Kind() {
		case reflect.String:
			// vault tag only appeared in string type
			vaultKey := rv.Type().Field(f).Tag.Get(vaultTag)
			if vaultKey != "" {
				result := ""
				result, err = c.key(vaultKey)
				if err != nil {
					return
				}
				c.addKV(vaultKey, result)
				val.SetString(result)
			}
		case reflect.Ptr:
			e := val.Elem()
			if e.Kind() == reflect.Struct {
				if err = c.scanStruct(e); err != nil {
					return
				}
			}
		case reflect.Struct:
			if err = c.scanStruct(val); err != nil {
				return
			}
		}
	}
	return nil
}
