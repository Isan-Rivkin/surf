package elastic

import (
	"io"
	"strings"

	"github.com/Jeffail/gabs/v2"
	accessor "github.com/isan-rivkin/surf/lib/common/jsonutil"
)

type ESIndicesResponse interface {
	Indices() ([]EsIndice, bool)
}
type ESIndexRespObj struct {
	container *gabs.Container
	indices   []EsIndice
}

func NewESIndexRespObj(body io.ReadCloser) (ESIndicesResponse, error) {
	container, err := accessor.NewJsonContainerFromReader(body)
	if err != nil {
		return nil, err
	}
	var indices []EsIndice
	if dict, exist := accessor.GetDict(container, ""); exist {
		for name := range dict {
			indices = append(indices, NewESIndiceObj(name))
		}
	}
	return &ESIndexRespObj{container: container, indices: indices}, nil
}

func (ir *ESIndexRespObj) Indices() ([]EsIndice, bool) {
	return ir.indices, len(ir.indices) > 0
}

type EsIndice interface {
	GetName() string
	IsDotIndex() bool
}

type ESIndiceObj struct {
	name  string
	isDot bool
}

func NewESIndiceObj(name string) EsIndice {
	return &ESIndiceObj{
		name:  name,
		isDot: strings.HasPrefix(name, "."),
	}
}
func (ei *ESIndiceObj) GetName() string {
	return ei.name
}
func (ei *ESIndiceObj) IsDotIndex() bool {
	return ei.isDot
}
