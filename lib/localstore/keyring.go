package localstore

import (
	kr "github.com/zalando/go-keyring"
)

type KeyRingStore struct {
	Service string
}

func (k *KeyRingStore) Write(key, value string) error {
	return kr.Set(k.Service, key, value)

}

func (k *KeyRingStore) Read(key string) (string, error) {
	secret, err := kr.Get(k.Service, key)
	return secret, err
}
