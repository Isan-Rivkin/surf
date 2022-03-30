package vault

import (
	"errors"

	vaultApi "github.com/hashicorp/vault/api"
	log "github.com/sirupsen/logrus"
)

type Client[A Authenticator] interface {
	Read(secretPath, optionalSecretVersion string) (map[string]interface{}, error)
	ListMounts() (map[string]*vaultApi.MountOutput, error)
	ListTree(basePath string) ([]*Node, error)
	ListTreeFiltered(basePath string) ([]*Node, error)
	GetVaultAddr() string
}

type Vaultclient[A Authenticator] struct {
	Auth    A
	_client *vaultApi.Client
}

func NewClient[A Authenticator](a A) Client[Authenticator] {
	return &Vaultclient[Authenticator]{Auth: a}
}

func (v *Vaultclient[A]) GetVaultAddr() string {
	return v.Auth.GetVaultAddr()
}

func (v *Vaultclient[A]) Read(secretPath, optionalSecretVersion string) (map[string]interface{}, error) {

	// get authenticated client
	client, err := v.getClient()

	if err != nil {
		return nil, err
	}

	// KV vault v2 support
	secretPath, err = AssemblePath(secretPath, client)

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

func (c *Vaultclient[A]) getClient() (*vaultApi.Client, error) {
	var err error
	if c._client == nil {
		c._client, err = c.Auth.Auth()
	}
	return c._client, err
}
