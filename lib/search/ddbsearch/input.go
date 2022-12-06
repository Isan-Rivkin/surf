package ddbsearch

type Input struct {
	// max number of go routines
	Parallel int
	// pattern to match against ddb tables to start search from - if empty then all buckets will be searched
	TableNamePattern string
	// the value to match search against
	Value string
}
