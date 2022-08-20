package common

import (
	"bytes"
	"fmt"
	"io"

	"github.com/Jeffail/gabs/v2"
)

type JsonAccessor struct {
	obj *gabs.Container
}

func NewJsonAccessorFromResponse(body io.ReadCloser) (*JsonAccessor, error) {
	buf := &bytes.Buffer{}
	if _, err := buf.ReadFrom(body); err != nil {
		return nil, fmt.Errorf("failed parsing body %s", err.Error())
	}
	// retrieve a byte slice from bytes.Buffer
	data := buf.Bytes()
	return NewJsonAccessor(data)
}

func NewJsonAccessor(payload []byte) (*JsonAccessor, error) {
	jsonParsed, err := gabs.ParseJSON(payload)
	if err != nil {
		return nil, fmt.Errorf("failed parsing json %s", err.Error())
	}
	return &JsonAccessor{obj: jsonParsed}, nil
}

func (ja *JsonAccessor) Keys(path string) []string {
	var keys []string
	obj := ja.obj
	if path != "" {
		obj = ja.obj.S("hits", "hits")
	}
	obj = ja.obj.Path("hits.hits")
	//fmt.Println("nuu ", obj)
	for _, c := range obj.Children() {
		for k, o := range c.ChildrenMap() {
			if k == "_source" {
				for sk := range o.ChildrenMap() {
					fmt.Printf("\t\t\t %s -> %s\n", k, sk)
				}
			} else {
				fmt.Printf("\t%s -> %v\n", k, o.Data())
			}
		}
	}
	for key := range obj.ChildrenMap() {
		keys = append(keys, key)
	}
	return keys
}
