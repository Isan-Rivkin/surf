package ddbsearch

import "fmt"

type MatchLevel string

var (
	// if match found object considered a match
	ObjectMatch MatchLevel = "object"
	// if table then will return all table match once sinmce object matched
	TableMatch MatchLevel = "table"
	// if set then will only match against table names
	TableNameOnlyMatch MatchLevel = "table_name_only"
	// if set then will match query against schema keys values and thats it
	SchemaKeysOnlyMatch MatchLevel = "schema_keys_only"
)

type Input struct {
	// max number of go routines
	Parallel int
	// pattern to match against ddb tables to start search from - if empty true then all tables will be searched
	TableNamePattern string
	// the value to match search against
	Value string
	//error tollerance if true will exit on any error
	FailFast bool
	// include global tables
	WithGlobalTables bool
	// Match level to search
	Match MatchLevel
}

func NewSearchInput(table, query string, failFast, withGlobalTables bool, match MatchLevel, parallel int) (*Input, error) {

	if query == "" {
		return nil, fmt.Errorf("query must not be empty tables %s", table)
	}
	if match == TableNameOnlyMatch && table == "" {
		return nil, fmt.Errorf("when tables level only must specify table name")
	}

	return &Input{
		FailFast:         failFast,
		Parallel:         parallel,
		TableNamePattern: table,
		Value:            query,
		WithGlobalTables: withGlobalTables,
		Match:            match,
	}, nil
}
