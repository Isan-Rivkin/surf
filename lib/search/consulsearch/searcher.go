package consulsearch

import (
	consul "github.com/isan-rivkin/surf/lib/consul"
	log "github.com/sirupsen/logrus"
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

type Searcher[C consul.ConsulClient, M Matcher] interface {
	Search(i *Input) (*Output, error)
}

type DefaultSearcher[C consul.ConsulClient, M Matcher] struct {
	Client     consul.ConsulClient
	Comparator Matcher
}

func NewSearchInput(value string, basePath string) *Input {
	return &Input{
		Value:    value,
		BasePath: basePath,
	}
}

func NewSearcher[C consul.ConsulClient, Comp Matcher](c consul.ConsulClient, m Matcher) Searcher[consul.ConsulClient, Matcher] {
	return &DefaultSearcher[consul.ConsulClient, Matcher]{
		Client:     c,
		Comparator: m,
	}
}

func (s *DefaultSearcher[CC, Matcher]) Search(i *Input) (*Output, error) {
	pairs, err := s.Client.List(i.BasePath)

	if err != nil {
		log.Warnf("Could not list all keys under prefix %s", i.BasePath)
		return nil, err
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
