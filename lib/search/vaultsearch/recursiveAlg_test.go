package vaultsearch_test

import (
	"testing"

	search "github.com/isan-rivkin/search-unified-recusive-fast/lib/search/vaultsearch"
	"github.com/magiconair/properties/assert"
)

func flat[T any](items [][]T) []T {
	var f []T
	for _, c := range items {
		f = append(f, c...)
	}
	return f
}
func toSet(items []int) map[int]bool {
	s := map[int]bool{}
	for _, i := range items {
		s[i] = true
	}
	return s
}

func TestSplitIntoNChunks(t *testing.T) {
	cases := []struct {
		Items          []int
		ChunksDesired  int
		ExpectedChunks int
	}{
		{[]int{1, 2, 3, 4, 5}, 5, 5},
		{[]int{1, 2, 3, 4, 5}, 1, 1},
		{[]int{1, 2, 3, 4, 5}, 3, 3},
		{[]int{1, 2, 3, 4, 5}, 10, 5},
	}

	for _, c := range cases {
		chunks := search.SplitIntoNChunks(c.Items, c.ChunksDesired)
		got := len(chunks)
		// compare chunks as expected
		assert.Equal(t, got, c.ExpectedChunks, "chunks not equal")
		// compare content inside as expected
		flatted := flat(chunks)
		shouldExist := toSet(c.Items)
		for _, v := range flatted {
			exist, _ := shouldExist[v]
			assert.Equal(t, exist, true, "content is wrong after split")
		}

	}
}
