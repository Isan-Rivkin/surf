package vaultsearch

import (
	"math"

	"github.com/isan-rivkin/surf/lib/vault"
)

func TEST() {}

type VC vault.Client[vault.Authenticator]

type Input struct {
	// number of parallel go routines while searching
	Prallel int
	// if true should stop after the first match
	StopIfFound bool
	// base path to start search from
	BasePath string
	// the value to match search against
	Value string
	// TODO:: not implemented yet, search inside secrets
	SearchSecretContent bool
}

type Output struct {
	Matches []*vault.Node
}

func NewSearchInput(val, basePath string, parallel int) *Input {

	return &Input{

		Prallel:             int(math.Max(1, float64(parallel))),
		BasePath:            basePath,
		StopIfFound:         false,
		Value:               val,
		SearchSecretContent: false,
	}
}

type Searcher[C VC, M Matcher] interface {
	Search(i *Input) (*Output, error)
}
