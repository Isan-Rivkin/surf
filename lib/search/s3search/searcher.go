package s3search

import (
	"fmt"
	"math"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go/aws"
	awsu "github.com/isan-rivkin/surf/lib/awsu"
	workPool "github.com/isan-rivkin/surf/lib/common"
	common "github.com/isan-rivkin/surf/lib/search"
	log "github.com/sirupsen/logrus"
)

const TooManyBucketsErr string = "TooManyBucketsErr"

type _s3AsyncRes struct {
	Bucket string
	Keys   []string
	Err    error
}

type Input struct {
	// max number of go routines
	Parallel int
	// pattern to match against s3 buckets to start search from - if empty then all buckets will be searched
	BucketNamePattern string
	// the value to match search against
	Value string
	// prefix for keys to start from
	Prefix string
	// TODO: implement search keys content
	SearchKeysContent bool
	// if BucketNamePattern is provided this is ignored
	// max number to allow listing all buckets when no BucketNamePattern is provided
	// if BucketNamePattern is not provided, the searcher will still allow searching if total buckets number is
	// smaller than MaxAllowedAllBuckets
	// if the total buckets number > MaxAllowedAllBuckets then AllowAllBucket is required to be true
	MaxAllowedAllBuckets int
	// if BucketNamePattern is provided this is ignored
	// if MaxAllowedAllBuckets < actual buckets user has and AllowAllBucket is true, then operation allowed
	AllowAllBucket bool
}

func NewSearchInput(bucketNamePattern, prefix, value string, parallel int, allowAllBuckets bool) *Input {
	return &Input{
		Prefix:               prefix,
		Value:                value,
		BucketNamePattern:    bucketNamePattern,
		Parallel:             parallel,
		MaxAllowedAllBuckets: parallel,
		AllowAllBucket:       allowAllBuckets,
	}
}

type Output struct {
	BucketToMatches map[string][]string
}

type Searcher[C awsu.S3API, M common.Matcher] interface {
	Search(i *Input) (*Output, error)
}

type DefaultSearcher[C awsu.S3API, M common.Matcher] struct {
	Client     awsu.S3API
	Comparator common.Matcher
}

func NewSearcher[C awsu.S3API, Comp common.Matcher](c awsu.S3API, m common.Matcher) Searcher[awsu.S3API, common.Matcher] {
	return &DefaultSearcher[awsu.S3API, common.Matcher]{
		Client:     c,
		Comparator: m,
	}
}

func (s *DefaultSearcher[CC, Matcher]) Search(i *Input) (*Output, error) {
	allBuckets, err := s.Client.ListAllBuckets()
	var targetBuckets []types.Bucket
	filteredResult := &Output{
		BucketToMatches: map[string][]string{},
	}
	if err != nil {
		return nil, fmt.Errorf("searcher failed listing buckets %s", err.Error())
	}

	if i.BucketNamePattern != "" {
		for _, b := range allBuckets {
			match, err := s.Comparator.IsMatch(i.BucketNamePattern, aws.StringValue(b.Name))
			if err != nil {
				return nil, fmt.Errorf("failed matching bucket name in comparator %s", err.Error())
			}
			if match {
				targetBuckets = append(targetBuckets, b)
			}
		}
	} else if i.AllowAllBucket || len(allBuckets) <= i.MaxAllowedAllBuckets {
		log.Warningf("going to search in all buckets %d might impact performance", len(allBuckets))
		targetBuckets = allBuckets
	} else {
		return nil, fmt.Errorf(TooManyBucketsErr)
	}
	log.WithField("buckets_number", len(targetBuckets)).Info("searching in buckets")

	// search keys
	workersNum := math.Min(float64(len(targetBuckets)), float64(i.Parallel))
	pool := workPool.NewWorkerPool(int(workersNum))

	asyncResults := make(chan *_s3AsyncRes, len(targetBuckets))

	for _, b := range targetBuckets {
		bucketName := aws.StringValue(b.Name)
		pool.Submit(func() {
			keys, err := s.Client.ListAllObjects(bucketName, i.Prefix)
			res := &_s3AsyncRes{
				Bucket: bucketName,
			}
			if err != nil {
				res.Err = err
			} else {
				for _, k := range keys {
					if isMatch, err := s.Comparator.IsMatch(i.Value, aws.StringValue(k.Key)); isMatch {
						if err != nil {
							log.WithError(err).WithField("key", aws.StringValue(k.Key)).Error("failed pattern matching key probablly bug")
							continue
						}
						res.Keys = append(res.Keys, aws.StringValue(k.Key))
					}
				}
				asyncResults <- res
			}
		})
	}

	pool.RunAll()
	size := len(targetBuckets)
	counter := 1
	for r := range asyncResults {
		if r.Err != nil {
			log.WithError(err).WithField("bucket", r.Bucket).Error("failed describing keys")
			continue
		}

		filteredResult.BucketToMatches[r.Bucket] = r.Keys

		if counter >= size {
			break
		}

		counter++
	}

	return filteredResult, nil
}
