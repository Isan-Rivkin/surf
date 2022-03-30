package vault

import (
	"fmt"
	"strings"

	vaultApi "github.com/hashicorp/vault/api"
)

func SecretToListOfStr(s *vaultApi.Secret) ([]string, error) {
	var result []string

	if s == nil || s.Data == nil {
		return nil, fmt.Errorf("failed converting secret to list")
	}

	for _, v := range s.Data {
		arr := v.([]interface{})
		for _, item := range arr {
			result = append(result, item.(string))
		}
	}

	return result, nil
}

func IsRootPath(basePath string) bool {
	return basePath == "/" || basePath == "" || basePath == "."
}

func IsStorage(pType string) bool {
	return pType == "generic" || pType == "kv"
}

// Vault path to Web url
// ${mount}/database/user -> https://${vaulrAddr}/ui/vault/secrets/${mount}/show/database/user
func PathToWebURL(vaultAddr, path string) string {

	if path == "" {
		return vaultAddr
	}

	proto := ""

	if !strings.HasPrefix(vaultAddr, "http") {
		proto = "https://"
	}

	uiBaseUrl := fmt.Sprintf("%s%s/ui/vault/secrets", proto, vaultAddr)

	splitted := strings.Split(path, "/")
	mount := splitted[0]

	if len(splitted) == 1 {
		return fmt.Sprintf("%s/%s", uiBaseUrl, mount)
	} else {
		pathSuffix := strings.Join(splitted[1:], "/")
		return fmt.Sprintf("%s/%s/show/%s", uiBaseUrl, mount, pathSuffix)
	}

}
