package localstore

type Store interface {
	Write(key, value string) error
	Read(key string) (string, error)
}

func NewStore() Store {
	s := &KeyRingStore{
		Service: "vault-searcher",
	}
	return s
}

type StoreManager[T Store] struct {
	Storage T
}

func NewStoreManager[T Store](s T) StoreManager[Store] {
	sm := StoreManager[Store]{Storage: s}
	return sm
}

func (sm *StoreManager[T]) UpdateStore(username, password string) error {
	if err := sm.Storage.Write("username", username); err != nil {
		return err
	}

	return sm.Storage.Write(username, password)
}

func (sm *StoreManager[T]) GetUserAndPwd() (string, string, error) {
	uname, err := sm.Storage.Read("username")
	if err != nil {
		return "", "", err
	}
	pwd, err := sm.Storage.Read(uname)
	return uname, pwd, err
}

func (sm *StoreManager[T]) IsExistAlready() bool {
	uname, err := sm.Storage.Read("username")
	if err != nil {
		return false
	}

	pwd, err := sm.Storage.Read(uname)
	return err == nil && pwd != ""
}
