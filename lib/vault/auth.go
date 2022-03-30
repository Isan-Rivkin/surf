package vault

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	vaultApi "github.com/hashicorp/vault/api"
	log "github.com/sirupsen/logrus"
)

type Authenticator interface {
	Auth() (*vaultApi.Client, error)
	GetVaultAddr() string
}

type LdapAuthenticator struct {
	username, password, vaultAddr string
}

func NewLdapAuth(username, password, vaultAddr string) Authenticator {
	return &LdapAuthenticator{
		username:  username,
		password:  password,
		vaultAddr: vaultAddr,
	}
}

func (la *LdapAuthenticator) GetVaultAddr() string {
	return la.vaultAddr
}

func (la *LdapAuthenticator) Auth() (*vaultApi.Client, error) {
	os.Setenv("VAULT_SKIP_VERIFY", "true")
	httpClient := &http.Client{}

	clientConf := &vaultApi.Config{Address: la.vaultAddr, HttpClient: httpClient}
	c, _ := vaultApi.NewClient(clientConf)

	data := map[string]any{
		"password": la.password,
	}

	path := fmt.Sprintf("auth/ldap/login/%s", la.username)

	token, err := c.Logical().Write(path, data)

	if err != nil {
		return nil, err
	}

	c.SetToken(token.Auth.ClientToken)
	return c, err
}

//getTokenFromAppRole will return token
func getTokenFromAppRole(client *vaultApi.Client, approle string) (string, int, error) {

	data := map[string]interface{}{
		"role_id": approle,
	}

	resp, err := client.Logical().Write("auth/approle/login", data)

	if err != nil {
		return "", 0, err
	}

	if resp.Auth == nil {
		return "", 0, errors.New("Response.Auth object empty")
	}

	log.WithField("ttlSeconds", resp.Auth.LeaseDuration).Debug("created new token")

	return resp.Auth.ClientToken, resp.Auth.LeaseDuration, nil
}

func newVaultClient(vaultAddr, approle string) (*vaultApi.Client, int, error) {

	httpClient := &http.Client{}

	clientConf := &vaultApi.Config{Address: vaultAddr, HttpClient: httpClient}
	client, _ := vaultApi.NewClient(clientConf)
	// get init token with app role
	sessionToken, ttlSeconds, err := getTokenFromAppRole(client, approle)

	if err != nil {
		log.WithError(err).Error("Getting session Token from Approle")
		panic(err)
	}

	log.Debug("Success receiving session token from AppRole")

	if err != nil {
		log.WithError(err).Error("error calling token after receiving")
		return nil, 0, err
	}

	// log token only if in debug mode
	log.WithField("token", sessionToken).Debug("created new token")

	client.SetToken(sessionToken)
	return client, ttlSeconds, err
}
