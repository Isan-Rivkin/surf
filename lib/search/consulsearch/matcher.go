package consulsearch

type Matcher interface {
	IsMatch(needle string, haystack string) (bool, error)
}
