package vaultsearch

import (
	"math"
	"sync"

	"github.com/isan-rivkin/surf/lib/vault"
	log "github.com/sirupsen/logrus"
)

type RecursiveSearcher[C VC, M Matcher] struct {
	Client     VC
	Comparator Matcher
}

func NewRecursiveSearcher[C VC, Comp Matcher](c VC, m Matcher) Searcher[VC, Matcher] {
	return &RecursiveSearcher[VC, Matcher]{
		Client:     c,
		Comparator: m,
	}
}

//
func SplitIntoNChunks[T any](items []T, n int) [][]T {

	chunksList := [][]T{}
	size := len(items)

	if size < n {
		n = size
	}

	fullChunkSize := size / n

	leftOverSize := size % n

	previousChunkIdx := 0

	for i := 0; i < n; i++ {
		from := previousChunkIdx
		to := previousChunkIdx + fullChunkSize

		// last elemt and has left over
		if leftOverSize > 0 && i == n-1 {
			to = size
		}

		chunk := items[from:to]
		chunksList = append(chunksList, chunk)
		previousChunkIdx = to
	}

	return chunksList
}

func (s *RecursiveSearcher[VC, Matcher]) filter(i *Input, all []*vault.Node) ([]*vault.Node, error) {
	filtered := []*vault.Node{}
	for _, n := range all {
		isMatch, err := s.Comparator.IsMatch(i.Value, n.GetFullPath())
		if err != nil {
			return nil, err
		}

		if isMatch {
			filtered = append(filtered, n)
		}
	}
	return filtered, nil
}

func (s *RecursiveSearcher[VC, Matcher]) Search(i *Input) (*Output, error) {
	var result []*vault.Node
	basePath := i.BasePath
	nodes, err := s.Client.ListTreeFiltered(basePath)

	if err != nil {
		return nil, err
	}

	if nodes == nil {
		log.Warnf("no results to query from base path given %s ", basePath)
		return nil, nil
	}

	poolSize := int(math.Min(float64(len(nodes)), float64(i.Prallel)))
	log.WithField("parallel", poolSize).Info("parallel pool size")

	nodeChunks := SplitIntoNChunks(nodes, poolSize)

	chunksResult := make(chan []*vault.Node)
	readerRoutineFinished := make(chan bool)
	go func() {
		for {

			nodesChunk, more := <-chunksResult
			if !more {
				log.Debug("closing channel!")
				break
			} else {
				log.Info("parsed chunk result of size ", len(nodesChunk))
				result = append(result, nodesChunk...)
			}
		}
		readerRoutineFinished <- true
	}()

	var wg sync.WaitGroup

	wg.Add(len(nodeChunks))

	for _, chunk := range nodeChunks {

		go func(n []*vault.Node) {

			subFolders, folderErr := s.expandFolders(n)

			if folderErr != nil {
				log.WithError(folderErr).Error("failed expanding folders ", basePath)

			} else {

				chunksResult <- subFolders
			}
			wg.Done()
		}(chunk)

	}

	wg.Wait()
	close(chunksResult)
	<-readerRoutineFinished
	filtered, err := s.filter(i, result)

	if err != nil {
		return nil, err
	}
	log.WithField("matches_found", len(filtered)).Info("finished.")
	return &Output{Matches: filtered}, nil
}

func (s *RecursiveSearcher[VC, Matcher]) expandFolders(nodes []*vault.Node) ([]*vault.Node, error) {

	result := &[]*vault.Node{}

	if nodes == nil {
		return nil, nil
	}

	for _, node := range nodes {

		log.WithField("root_path", node.GetFullPath()).Info("searching...")
		if err := s.expandSingleFolder(node, result); err != nil {
			return nil, err
		}
	}

	return *result, nil
}

func (s *RecursiveSearcher[VC, Matcher]) expandSingleFolder(node *vault.Node, result *[]*vault.Node) error {

	if node == nil {
		return nil
	}

	fullPath := node.GetFullPath()

	// if already full path to a secret type

	if node.T == vault.Secret {
		log.Debug("recursion stopping no more folders", node.GetFullPath())
		*result = append(*result, node)
		return nil
	}

	leafNodes, err := s.Client.ListTreeFiltered(fullPath)

	if err != nil {
		return err
	}

	for _, lf := range leafNodes {
		if err := s.expandSingleFolder(lf, result); err != nil {
			return err
		}
	}
	return nil
}
