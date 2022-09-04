package essearch

import (
	"time"

	"bytes"
	"encoding/json"
	"io"

	"github.com/aquasecurity/esquery"
	"github.com/isan-rivkin/surf/lib/common"
)

type QueryBuilder interface {
	// bool query
	WithMustContain(query string) QueryBuilder
	WithMustNotContain(query string) QueryBuilder
	WithShouldContain(query string) QueryBuilder
	WithTimeRangeWindow(windowTime, windowOffsetTime, timeKey, timeFmt string) QueryBuilder
	WithSize(size uint64) QueryBuilder
	BuildBoolQuery() (io.Reader, []byte, error)
	// dummy query
	BuildSimpleKQLQuery(query string) (io.Reader, error)
}

type DefaultQueryBuilder struct {
	kqlQuery         string
	mustContain      []string
	mustNotContain   []string
	shouldContain    []string
	timeFmt          string
	windowTime       string
	windowOffsetTime string
	timeKey          string
	size             uint64
}

func NewQueryBuilder() QueryBuilder {
	return &DefaultQueryBuilder{}
}

func (qb *DefaultQueryBuilder) WithSize(size uint64) QueryBuilder {
	qb.size = size
	return qb
}
func (qb *DefaultQueryBuilder) buildKQLQuery() (string, map[string]interface{}) {
	return "query_string", map[string]interface{}{
		"query": qb.kqlQuery,
	}
}

func (qb *DefaultQueryBuilder) WithMustContain(query string) QueryBuilder {
	if query != "" {
		qb.mustContain = append(qb.mustContain, query)
	}
	return qb
}
func (qb *DefaultQueryBuilder) WithMustNotContain(query string) QueryBuilder {
	if query != "" {
		qb.mustNotContain = append(qb.mustNotContain, query)
	}
	return qb
}
func (qb *DefaultQueryBuilder) WithShouldContain(query string) QueryBuilder {
	if query != "" {
		qb.shouldContain = append(qb.shouldContain, query)
	}
	return qb
}

func (qb *DefaultQueryBuilder) WithTimeRangeWindow(windowTime, windowOffsetTime, timeKey, timeFmt string) QueryBuilder {

	qb.windowTime = windowTime
	qb.windowOffsetTime = windowOffsetTime
	qb.timeKey = timeKey
	qb.timeFmt = timeFmt
	return qb
}

func (qb *DefaultQueryBuilder) buildTimeRangeWindowFilter() (*esquery.RangeQuery, error) {
	windowTime := qb.windowTime
	windowOffsetTime := qb.windowOffsetTime
	timeKey := qb.timeKey
	// if default values set, don't do nothing
	if windowTime == "" && windowOffsetTime == common.TimeNow {
		return nil, nil
	} else if windowTime == "" { // if window size set
		windowTime = "2d"
	}

	if windowTime != "" && windowOffsetTime != "" {

		from, to, err := common.GetTimeWindow(windowTime, windowOffsetTime)
		if err != nil {
			return nil, err
		}

		return BuildTimeRangeFilter(timeKey, qb.timeFmt, &from, &to), nil
	}
	return nil, nil
}

func (qb *DefaultQueryBuilder) BuildBoolQuery() (io.Reader, []byte, error) {
	timeQuery, err := qb.buildTimeRangeWindowFilter()
	if err != nil {
		return nil, nil, err
	}
	_ = common.Get_ISO_UTC_Timeoffset()

	var filters []esquery.Mappable
	if timeQuery != nil {
		filters = append(filters, timeQuery)
	}
	mapQuery, queryJson, err := NewBoolQuery(qb.mustContain, qb.mustNotContain, qb.shouldContain, "", qb.size, filters...)
	if err != nil {
		return nil, nil, err
	}
	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(mapQuery); err != nil {
		return nil, nil, err
	}

	return &buf, queryJson, nil
}

func (qb *DefaultQueryBuilder) BuildSimpleKQLQuery(rawQuery string) (io.Reader, error) {
	qb.kqlQuery = rawQuery
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

func BuildQueryStringQuery(queryStr, timeZone string) *esquery.CustomQueryMap {
	qs := map[string]interface{}{
		"query":            queryStr,
		"analyze_wildcard": true,
	}
	if timeZone != "" {
		qs["time_zone"] = timeZone
	}
	m := map[string]interface{}{
		"query_string": qs,
	}
	return esquery.CustomQuery(m)
}

func BuildTimeRangeFilter(field, timeFmt string, from, to *time.Time) *esquery.RangeQuery {
	q := esquery.Range(field).Format(timeFmt)
	if from != nil {
		q.Gte(from)
	}
	if to != nil {
		q.Lte(to)
	}
	return q
}

func NewBoolQuery(mustQueries, mustNotQueries, shouldQueries []string, timeZone string, size uint64, filters ...esquery.Mappable) (map[string]interface{}, []byte, error) {
	boolQuery := esquery.Bool()
	for _, m := range mustQueries {
		boolQuery.Must(BuildQueryStringQuery(m, timeZone))
	}
	for _, m := range mustNotQueries {
		boolQuery.MustNot(BuildQueryStringQuery(m, timeZone))
	}
	for _, m := range shouldQueries {
		boolQuery.Should(BuildQueryStringQuery(m, timeZone))
	}

	for _, f := range filters {
		boolQuery.Filter(f)
	}
	qb := esquery.Query(boolQuery).Size(size)

	jsonQuery, err := qb.MarshalJSON()
	mapQuery := qb.Map()
	return mapQuery, jsonQuery, err
}
