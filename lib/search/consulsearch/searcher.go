package consulsearch

import (
	"fmt"

	consul "github.com/isan-rivkin/surf/lib/consul"
	common "github.com/isan-rivkin/surf/lib/search"
)

type Input struct {
	// base path to start search from
	BasePath string
	// the value to match search against
	Value string
	// TODO: implement search keys content
	SearchKeysContent bool
}

type Output struct {
	Matches []string
}

type Searcher[C consul.ConsulClient, M common.Matcher] interface {
	Search(i *Input) (*Output, error)
}

type DefaultSearcher[C consul.ConsulClient, M common.Matcher] struct {
	Client     consul.ConsulClient
	Comparator common.Matcher
}

func NewSearchInput(value string, basePath string) *Input {
	return &Input{
		Value:    value,
		BasePath: basePath,
	}
}

func NewSearcher[C consul.ConsulClient, Comp common.Matcher](c consul.ConsulClient, m common.Matcher) Searcher[consul.ConsulClient, common.Matcher] {
	return &DefaultSearcher[consul.ConsulClient, common.Matcher]{
		Client:     c,
		Comparator: m,
	}
}

func (s *DefaultSearcher[CC, Matcher]) Search(i *Input) (*Output, error) {
	pairs, err := s.Client.List(i.BasePath)

	if err != nil {
		return nil, fmt.Errorf("failed listing all keys under the prefix %s - %s", i.BasePath, err.Error())
	}

	matches := []string{}
	for _, pair := range pairs {
		key := pair.Key
		value := i.Value
		match, err := s.Comparator.IsMatch(value, key)

		if err != nil {
			return nil, err
		}

		if match {
			matches = append(matches, key)
		}
	}
	return &Output{Matches: matches}, nil
}
