package consul

type Matcher interface {
	IsMatch(needle string, haystack string) (bool, error)
}
