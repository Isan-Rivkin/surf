package elastic

import (
	"fmt"

	"github.com/Jeffail/gabs/v2"
	accessor "github.com/isan-rivkin/surf/lib/common/jsonutil"
)

type EsObj interface {
	GetIndex() (string, bool)
	GetType() (string, bool)
	GetID() (string, bool)
	GetScore() (float64, bool)
	GetSourceKeys() ([]string, bool)
	GetSourceAsJson() (string, bool)
	GetSourceStrVal(path string) (string, bool)
}

type EsDoc struct {
	raw *gabs.Container
}

func NewEsDoc(raw *gabs.Container) EsObj {
	return &EsDoc{raw: raw}
}

func (o *EsDoc) GetSourceStrVal(path string) (string, bool) {
	return accessor.GetValue[string](o.raw, fmt.Sprintf("_source.%s", path))
}

func (o *EsDoc) GetSourceAsJson() (string, bool) {
	srcObj, ok := accessor.GetNested(o.raw, "_source")
	if !ok || srcObj == nil {
		return "", false
	}
	return srcObj.String(), ok
}

func (o *EsDoc) GetSourceKeys() ([]string, bool) {

	var keys []string
	childMap, ok := accessor.GetDict(o.raw, "_source")
	if !ok {
		return nil, ok
	}
	for k := range childMap {
		keys = append(keys, k)
	}
	return keys, ok
}

func (o *EsDoc) GetIndex() (string, bool) {
	return accessor.GetValue[string](o.raw, "_index")
}

func (o *EsDoc) GetType() (string, bool) {
	return accessor.GetValue[string](o.raw, "_type")
}

func (o *EsDoc) GetID() (string, bool) {
	return accessor.GetValue[string](o.raw, "_id")
}

func (o *EsDoc) GetScore() (float64, bool) {
	return accessor.GetValue[float64](o.raw, "_score")
}

type ESResponse interface {
	GetHitsCount() (int, error)
	GetHits() ([]EsObj, error)
	GetMaxScore() (float64, error)
}

func (sr *SearchResponse) Result() (*gabs.Container, error) {
	if sr.Container == nil {
		var err error
		sr.Container, err = accessor.NewJsonContainerFromReader(sr.RawResponse.Body)
		return sr.Container, err
	}
	return sr.Container, nil
}

func (sr *SearchResponse) GetMaxScore() (float64, error) {
	obj, err := sr.Result()
	if err != nil {
		return 0, err
	}
	score, ok := accessor.GetValue[float64](obj, "hits.max_score")
	if !ok {
		return 0, nil
	}
	return score, nil
}

func (sr *SearchResponse) GetHitsCount() (int, error) {
	obj, err := sr.Result()
	if err != nil {
		return 0, err
	}

	total, ok := accessor.GetValue[float64](obj, "hits.total")

	if !ok {
		return 0, nil
	}
	return int(total), nil
}

func (sr *SearchResponse) GetHits() ([]EsObj, error) {
	var result []EsObj
	obj, err := sr.Result()
	if err != nil {
		return nil, err
	}
	rawDocs, ok := accessor.GetArray(obj, "hits.hits")
	if !ok {
		return nil, fmt.Errorf("NoHits %s", "hits.hits")
	}
	for _, rd := range rawDocs {
		d := NewEsDoc(rd)
		result = append(result, d)
	}
	return result, nil
}
