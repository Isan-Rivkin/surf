package vault

import (
	"errors"
	"os"
	"time"

	vaultApi "github.com/hashicorp/vault/api"
	log "github.com/sirupsen/logrus"
)

type Client interface {
	Read(secretPath, optionalSecretVersion string) (map[string]interface{}, error)
}

type Vaultclient struct {
	_client                        *vaultApi.Client
	expiryDate                     time.Time
	appRole, vaultAddr, authMethod string
}

func NewClientFromApprole(vaultAddr, appRole string) Client {
	os.Setenv("VAULT_SKIP_VERIFY", "true")
	c, ttlSeconds, err := newVaultClient(vaultAddr, appRole)
	if err != nil {
		panic(err)
	}
	return &Vaultclient{
		_client:    c,
		vaultAddr:  vaultAddr,
		appRole:    appRole,
		authMethod: "approle",
		expiryDate: time.Now().Add(time.Duration(ttlSeconds-120) * time.Second),
	}
}

func (v *Vaultclient) Read(secretPath, optionalSecretVersion string) (map[string]interface{}, error) {

	// get authenticated client
	client, err := v.getClient()

	if err != nil {
		return nil, err
	}

	// KV vault v2 support
	secretPath, err = assemblePath(secretPath, client)

	if err != nil {
		return nil, err
	}

	// check if v2 optional secret version
	var versionParam map[string][]string
	if optionalSecretVersion != "" {
		versionParam = map[string][]string{
			"version": {optionalSecretVersion},
		}
	}

	// read secret

	secrets, err := client.Logical().ReadWithData(secretPath, versionParam)
	fields := log.Fields{
		"path":          secretPath,
		"secretVersion": optionalSecretVersion,
	}

	if err != nil {
		log.WithError(err).WithFields(fields).Error("reading from path")
		return nil, err
	}
	if secrets == nil {
		log.WithError(err).WithFields(fields).Error("reading from path")
		return nil, errors.New("ErrSecretsNilFromPath")
	}

	return secrets.Data, nil
}

func (c *Vaultclient) getClient() (*vaultApi.Client, error) {
	if c._client == nil || time.Now().After(c.expiryDate) {
		if c.authMethod == "approle" {
			clnt, ttlSeconds, err := newVaultClient(c.vaultAddr, c.appRole)
			if err != nil {
				return nil, err
			}
			c._client = clnt
			c.expiryDate = time.Now().Add(time.Duration(ttlSeconds-120) * time.Second)
		}
	}

	if c._client == nil {
		return nil, errors.New("NoVaultClient")
	}

	return c._client, nil
}
