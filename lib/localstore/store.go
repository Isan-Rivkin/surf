package localstore

import "fmt"

type Namespace string

type Store interface {
	Write(key, value string) error
	Read(key string) (string, error)
}

func NewStore(serviceName string) Store {
	s := &KeyRingStore{
		Service: serviceName,
	}
	return s
}

type StoreManager[T Store] struct {
	Storage               T
	NamespaceToKeysMapper map[Namespace][]string
}

func NewStoreManager[T Store](s T, namespaces map[Namespace][]string) StoreManager[Store] {

	sm := StoreManager[Store]{
		Storage:               s,
		NamespaceToKeysMapper: namespaces,
	}
	return sm
}

func (sm *StoreManager[T]) withNs(ns Namespace, key string) string {
	return fmt.Sprintf("%s.%s", ns, key)
}

func (sm *StoreManager[T]) GetValues(ns Namespace) (map[string]string, error) {

	result := map[string]string{}

	keys, exist := sm.NamespaceToKeysMapper[ns]

	if !exist {
		return nil, fmt.Errorf("no such namespace %s", ns)
	}

	for _, k := range keys {
		key := sm.withNs(ns, k)
		val, err := sm.Storage.Read(key)
		if err != nil {
			return nil, fmt.Errorf("failed reading key namespace %s for key  %s", k, err.Error())
		}
		result[k] = val
	}

	return result, nil
}

func (sm *StoreManager[T]) ListAll() ([]map[string]string, error) {
	allNsResult := []map[string]string{}
	for ns, _ := range sm.NamespaceToKeysMapper {
		res, err := sm.GetValues(ns)
		if err != nil {
			return nil, err
		}
		allNsResult = append(allNsResult, res)
	}
	return allNsResult, nil
}
func (sm *StoreManager[T]) SetNSValues(ns Namespace, data map[string]string) error {
	baseKeys, exist := sm.NamespaceToKeysMapper[ns]

	if !exist {
		return fmt.Errorf("no such namespace %s", ns)
	}

	for _, bk := range baseKeys {
		bv, exist := data[bk]

		if !exist {
			return fmt.Errorf("must have basic key %s in namespace %s", bk, ns)
		}

		nsKey := sm.withNs(ns, bk)
		if err := sm.Storage.Write(nsKey, bv); err != nil {
			return fmt.Errorf("failed writing base key %s to namespace %s - %s", nsKey, ns, err.Error())
		}

	}

	for k, val := range data {
		key := sm.withNs(ns, k)
		if err := sm.Storage.Write(key, val); err != nil {
			return fmt.Errorf("failed writing key %s to namespace %s", key, ns)
		}
	}
	return nil
}

func (sm *StoreManager[T]) IsNamespaceSet(ns Namespace) bool {
	vals, err := sm.GetValues(ns)

	if err != nil {
		return false
	}

	keys := sm.NamespaceToKeysMapper[ns]
	return len(vals) == len(keys)
}
