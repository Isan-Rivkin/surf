package elastic

import (
	"bytes"
	"encoding/json"
	"io"
)

type QueryBuilder interface {
	WithKQL(query string) QueryBuilder
	Build() (io.Reader, error)
}

type DefaultQueryBuilder struct {
	kqlQuery string
}

func NewQueryBuilder() QueryBuilder {
	return &DefaultQueryBuilder{}
}

// "field:ok some word "
func (qb *DefaultQueryBuilder) WithKQL(query string) QueryBuilder {
	qb.kqlQuery = query
	return qb
}

func (qb *DefaultQueryBuilder) buildKQLQuery() (string, map[string]interface{}) {
	return "query_string", map[string]interface{}{
		"query": qb.kqlQuery,
	}
}
func (qb *DefaultQueryBuilder) Build() (io.Reader, error) {

	var buf bytes.Buffer

	kqlKey, kqlObj := qb.buildKQLQuery()

	query := map[string]interface{}{
		"query": map[string]interface{}{
			kqlKey: kqlObj,
		},
	}

	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, err
	}

	return &buf, nil
}
