package vault

import (
	"errors"
	"fmt"
	"path"
	"strings"

	vaultApi "github.com/hashicorp/vault/api"
	log "github.com/sirupsen/logrus"
)

func assemblePath(path string, client *vaultApi.Client) (string, error) {

	// check if already contains v2

	if splitted := strings.Split(path, "/"); len(splitted) > 2 {
		if splitted[1] == "data" {
			//already appended before
			return path, nil
		}
	}

	// check if v2
	mount, v2, err := isKVV2(path, client)

	if err != nil {
		log.WithError(err).WithField("keyPath", path).Error("failed checking mount version")
		return "", err
	}

	if v2 {
		// for read its data for list its metadata
		path = AddPrefixToVKVPath(path, mount, "data")
	}

	return path, nil
}

// isKVV2 check if path belongs to a kv v2 mounts taken from vault/kv_helpers.god
func isKVV2(path string, client *vaultApi.Client) (string, bool, error) {
	mountPath, version, err := KvPreflightVersionRequest(client, path)
	if err != nil {
		return "", false, err
	}
	return mountPath, version == 2, nil
}

// KvPreflightVersionRequest taken from vault/command/kv_helpers.go
// check if the path given is kv v2 or v1
func KvPreflightVersionRequest(client *vaultApi.Client, path string) (string, int, error) {

	endpoint := fmt.Sprintf("%s/%s", "sys/internal/ui/mounts/", path)
	secret, err := client.Logical().Read(endpoint)
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{
				"path": path,
			}).Error("failed to read from path during kv version check")
		return "", 0, err
	}
	if secret == nil {
		log.WithFields(log.Fields{
			"path": path,
		}).Error("nil response from pre-flight request")
		return "", 0, errors.New("NilResponseKvVersionCheckErr")
	}

	// extract the mount from path
	var mountPath string
	if mountPathRaw, ok := secret.Data["path"]; ok {
		mountPath = mountPathRaw.(string)
	}

	// extract the KV version
	options := secret.Data["options"]
	if options == nil {
		return mountPath, 1, nil
	}
	versionRaw := options.(map[string]interface{})["version"]
	if versionRaw == nil {
		return mountPath, 1, nil
	}
	version := versionRaw.(string)
	switch version {
	case "", "1":
		return mountPath, 1, nil
	case "2":
		return mountPath, 2, nil
	}
	return mountPath, 1, nil
}

// AddPrefixToVKVPath in v2 di/bluh will become di/data/bluh and an optional version param in the url
func AddPrefixToVKVPath(p, mountPath, apiPrefix string) string {
	switch {
	case p == mountPath, p == strings.TrimSuffix(mountPath, "/"):
		return path.Join(mountPath, apiPrefix)
	default:
		p = strings.TrimPrefix(p, mountPath)
		return path.Join(mountPath, apiPrefix, p)
	}
}
