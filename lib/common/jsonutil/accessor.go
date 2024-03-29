package jsonutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/Jeffail/gabs/v2"
)

func NewJsonContainerFromMap(rootKey string, data map[string]string) (*gabs.Container, error) {
	if rootKey == "" {
		return nil, fmt.Errorf("ErrRootKeyEmptyString")
	}
	var buf bytes.Buffer

	payload := map[string]interface{}{
		rootKey: data,
	}

	if err := json.NewEncoder(&buf).Encode(payload); err != nil {
		return nil, err
	}
	return NewJsonContainerFromBytes(buf.Bytes())
}

func NewJsonContainerFromInterface(rootKey string, data map[string]any) (*gabs.Container, error) {
	if rootKey == "" {
		return nil, fmt.Errorf("ErrRootKeyEmptyString")
	}
	var buf bytes.Buffer

	payload := map[string]any{
		rootKey: data,
	}

	if err := json.NewEncoder(&buf).Encode(payload); err != nil {
		return nil, err
	}
	return NewJsonContainerFromBytes(buf.Bytes())
}

func NewJsonContainerFromReader(body io.ReadCloser) (*gabs.Container, error) {
	buf := &bytes.Buffer{}
	if _, err := buf.ReadFrom(body); err != nil {
		return nil, fmt.Errorf("failed parsing body %s", err.Error())
	}
	data := buf.Bytes()
	return NewJsonContainerFromBytes(data)
}

func NewJsonContainerFromBytes(payload []byte) (*gabs.Container, error) {

	jsonParsed, err := gabs.ParseJSON(payload)
	if err != nil {
		return nil, fmt.Errorf("failed parsing json %s", err.Error())
	}

	return jsonParsed, err
}

func GetNested(obj *gabs.Container, path string) (*gabs.Container, bool) {
	if obj == nil || path == "" {
		return obj, obj != nil
	}
	return obj.Path(path), obj.ExistsP(path)
}
func GetArray(obj *gabs.Container, path string) ([]*gabs.Container, bool) {
	arr, exist := GetNested(obj, path)
	if !exist {
		return nil, exist
	}

	return arr.Children(), arr.Children() != nil
}

func GetValue[T any](obj *gabs.Container, path string) (T, bool) {
	tObj, exist := GetNested(obj, path)
	var t T
	if !exist {
		return t, exist
	}
	t, ok := tObj.Data().(T)
	return t, ok
}

func GetDict(obj *gabs.Container, path string) (map[string]*gabs.Container, bool) {
	dict, exist := GetNested(obj, path)

	if !exist {
		return nil, exist
	}

	return dict.ChildrenMap(), exist
}
