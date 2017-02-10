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
	return c.scan(rv)
}

var ErrStringType = errors.New("elem must be string type with `vault` tag")

// scan iter the struct for look tag `vault`
func (c *Client) scan(rv reflect.Value) error {
	switch rv.Kind() {
	case reflect.Ptr:
		return c.scan(rv.Elem())
	case reflect.Struct:
		for f, fs := 0, rv.NumField(); f < fs; f++ {
			val := rv.Field(f)
			vaultKey := rv.Type().Field(f).Tag.Get(vaultTag)
			if vaultKey != "" {
				if val.Kind() == reflect.String {
					result, err := c.key(vaultKey)
					if err != nil {
						return err
					}
					c.addKV(vaultKey, result)
					val.SetString(result)
				} else {
					return ErrStringType
				}
			} else {
				if (val.Kind() == reflect.Ptr && val.Elem().Kind() == reflect.Struct) || val.Kind() == reflect.Struct {
					if err := c.scan(val); err != nil {
						return err
					}
				}
			}
		}
		return nil
	default:
		return errors.New("only support struct type")
	}
}
