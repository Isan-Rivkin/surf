package consul

import (
	log "github.com/sirupsen/logrus"
)

type Input struct {
	// base path to start search from
	BasePath string
	// the value to match search against
	Value string
}

type Output struct {
	Matches []string
}

type Searcher[C ConsulClient, M Matcher] interface {
	Search(i *Input) (*Output, error)
}

type DefaultSearcher[C ConsulClient, M Matcher] struct {
	Client     ConsulClient
	Comparator Matcher
}

func NewSearchInput(value string, basePath string) *Input {
	return &Input{
		Value:    value,
		BasePath: basePath,
	}
}

func NewSearcher[C ConsulClient, Comp Matcher](c ConsulClient, m Matcher) Searcher[ConsulClient, Matcher] {
	return &DefaultSearcher[ConsulClient, Matcher]{
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
