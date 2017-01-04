package vault

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	vaultapi "github.com/hashicorp/vault/api"

	"github.com/teambition/confl/sources/common"
	"golang.org/x/net/context"
)

// Client is a wrapper around the vault client
type Client struct {
	client *vaultapi.Client
}

// getParameter get a k/v from parameters
func getParameter(key string, parameters map[string]string) string {
	value := parameters[key]
	if value == "" {
		// panic if a configuration is missing
		panic(fmt.Sprintf("%s is missing from configuration", key))
	}
	return value
}

// panicToError converts a panic to an error
func panicToError(err *error) {
	if r := recover(); r != nil {
		switch t := r.(type) {
		case string:
			*err = errors.New(t)
		case error:
			*err = t
		default: // panic again if we don't know how to handle
			panic(r)
		}
	}
}

// authenticate with the remote client
func authenticate(c *vaultapi.Client, authType string, params map[string]string) (err error) {
	var secret *vaultapi.Secret

	// handle panics gracefully by creating an error
	// this would happen when we get a parameter that is missing
	defer panicToError(&err)

	switch authType {
	case "app-id":
		secret, err = c.Logical().Write("/auth/app-id/login", map[string]interface{}{
			"app_id":  getParameter("app-id", params),
			"user_id": getParameter("user-id", params),
		})
	case "github":
		secret, err = c.Logical().Write("/auth/github/login", map[string]interface{}{
			"token": getParameter("token", params),
		})
	case "token":
		c.SetToken(getParameter("token", params))
		secret, err = c.Logical().Read("/auth/token/lookup-self")
	case "userpass":
		username, password := getParameter("username", params), getParameter("password", params)
		secret, err = c.Logical().Write(fmt.Sprintf("/auth/userpass/login/%s", username), map[string]interface{}{
			"password": password,
		})
	}

	if err != nil {
		return err
	}

	// if the token has already been set
	if c.Token() != "" {
		return nil
	}

	// the default place for a token is in the auth section
	// otherwise, the backend will set the token itself
	c.SetToken(secret.Auth.ClientToken)
	return nil
}

func getConfig(address, cert, key, caCert string) (*vaultapi.Config, error) {
	conf := vaultapi.DefaultConfig()
	conf.Address = address

	tlsConfig := &tls.Config{}
	if cert != "" && key != "" {
		clientCert, err := tls.LoadX509KeyPair(cert, key)
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{clientCert}
		tlsConfig.BuildNameToCertificate()
	}

	if caCert != "" {
		ca, err := ioutil.ReadFile(caCert)
		if err != nil {
			return nil, err
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(ca)
		tlsConfig.RootCAs = caCertPool
	}

	conf.HttpClient.Transport = &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	return conf, nil
}

// NewClient returns an *vault.Client with a connection to named machines.
// It returns an error if a connection to the cluster cannot be made.
func NewClient(address, authType string, params map[string]string) (*Client, error) {
	if authType == "" {
		return nil, errors.New("you have to set the auth type when using the vault backend")
	}
	conf, err := getConfig(address, params["cert"], params["key"], params["caCert"])

	if err != nil {
		return nil, err
	}

	c, err := vaultapi.NewClient(conf)
	if err != nil {
		return nil, err
	}

	if err := authenticate(c, authType, params); err != nil {
		return nil, err
	}
	return &Client{c}, nil
}

func (c *Client) Key(ctx context.Context, key string) (string, error) {
	resp, err := c.client.Logical().Read(key)
	if err != nil {
		return "", err
	}
	if resp == nil || resp.Data == nil {
		return "", fmt.Errorf("key(%s) not exist", key)
	}

	if value, ok := resp.Data["value"]; ok {
		if text, ok := value.(string); ok {
			return text, nil
		}
	}
	return "", errors.New("vautl need a string value")
}

// NOTE
// Watch not implemented at the moment
func (c *Client) Watch(ctx context.Context, key string) <-chan *common.Response {
	return nil
}
