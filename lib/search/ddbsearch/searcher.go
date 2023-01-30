package ddbsearch

import (
	"fmt"
	"math"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	awsu "github.com/isan-rivkin/surf/lib/awsu"
	workPool "github.com/isan-rivkin/surf/lib/common"
	common "github.com/isan-rivkin/surf/lib/search"
	log "github.com/sirupsen/logrus"
)

type OutputHit struct {
	TableName  string
	HitLevel   MatchLevel
	ObjectData map[string]*string
}

type _AsyncOutputs struct {
	Hits      []*OutputHit
	TableName string
	Err       error
}

type Output struct {
	Matches []*OutputHit
}

type Searcher[Client awsu.DDBApi, Matcher common.Matcher] interface {
	Search(i *Input) (*Output, error)
}

type DefaultSearcher[Client awsu.DDBApi, Matcher common.Matcher] struct {
	Client     Client
	Comparator Matcher
	Parser     ObjParserFactory
}

func NewSearcher[Client awsu.DDBApi, Matcher common.Matcher](c awsu.DDBApi, m common.Matcher, p ObjParserFactory) Searcher[awsu.DDBApi, common.Matcher] {
	return &DefaultSearcher[awsu.DDBApi, common.Matcher]{
		Client:     c,
		Comparator: m,
		Parser:     p,
	}
}

func (s *DefaultSearcher[CC, Matcher]) Search(i *Input) (*Output, error) {
	output := &Output{}
	// list all tables (pre describe)
	allTables, err := s.Client.ListCombinedTables(true, i.WithGlobalTables)
	var tablesToDescribe []awsu.DDBTableDescriber
	if err != nil {
		return nil, fmt.Errorf("failed listing tables %s", err.Error())
	}
	// filter only tables to describe
	if i.TableNamePattern != "" {
		for _, t := range allTables {
			isMatch, err := s.Comparator.IsMatch(i.TableNamePattern, t.TableName())
			log.WithError(err).WithFields(log.Fields{
				"table_name_pattern": i.TableNamePattern,
				"table_evaluated":    t.TableName(),
				"is_match":           isMatch,
			}).Trace("match evaluation for table name")

			if err != nil {
				log.WithError(err).Error(err)
				return nil, err
			}
			if isMatch {
				tablesToDescribe = append(tablesToDescribe, t)
			}
		}
	} else {
		log.Debug("all tables where chosen to search since table name pattern is empty")
		tablesToDescribe = allTables
	}

	log.Debugf("table pattern %s matched %d tables to search in", i.TableNamePattern, len(tablesToDescribe))

	if len(tablesToDescribe) == 0 {
		return output, nil
	}
	// search inside tables
	// TODO parallel search inside tables not only between tables
	asyncResults := make(chan *_AsyncOutputs, len(tablesToDescribe))

	workersNum := math.Min(float64(len(tablesToDescribe)), float64(i.Parallel))
	pool := workPool.NewWorkerPool(int(workersNum))
	for _, t := range tablesToDescribe {
		pool.Submit(func() {
			lg := log.WithFields(log.Fields{
				"query": i.Value,
				"table": t.TableName(),
			})
			lg.Debug("starting search task routine")
			res := &_AsyncOutputs{TableName: t.TableName()}
			tDescriber, err := s.Client.DescribeTable(t.TableName(), t.IsGlobalTable())
			if err != nil {
				lg.WithError(err).Debug("failed during describing table")
				res.Err = err
			} else {
				// TODO: use search level input param
				schemas, err := tDescriber.GetSchemaDefinitions()
				if err != nil {
					lg.WithError(err).Debug("failed while fetching schema definitions")
					res.Err = err
				} else {
					p := s.Parser.New(
						WithFmtProto(schemas, false, true, " "),
						WithFmtGoString(schemas, false, false),
					)
					searchables, err := s.SearchTableData(tDescriber.TableName(), i, p, lg)
					if err != nil {
						lg.WithError(err).Debug("failed while searchle a single table data")
						res.Err = err
					} else {
						res.Hits = searchables
					}
				}
			}
			asyncResults <- res
		})
	}
	log.Debug("start running all jobs")
	pool.RunAll()
	size := len(tablesToDescribe)
	counter := 1
	for r := range asyncResults {
		if r.Err != nil {
			log.WithError(err).WithField("table", r.TableName).Error("failed searching in table")
			if i.FailFast {
				return nil, err
			}
			continue
		}

		output.Matches = append(output.Matches, r.Hits...)

		if counter >= size {
			break
		}

		counter++
	}

	return output, nil
}

func (s *DefaultSearcher[CC, Matcher]) SearchSingleObject(input *Input, obj map[string]*string, lg *log.Entry) (bool, error) {
	lg.WithField("obj", fmt.Sprintf("%#v", obj)).Trace("starting match evaluation inside a single object")
	for k, v := range obj {
		lgo := lg.WithFields(
			log.Fields{
				"search_level": input.Match,
				"key":          k,
				"value":        aws.StringValue(v),
			})

		match, err := s.Comparator.IsMatch(k, input.Value)
		lgo.WithError(err).WithField("is_key_match", match).Trace("key match evaluation")
		if err != nil {
			return false, err
		}
		if match {
			return true, nil
		}
		if v == nil || input.Match != ObjectMatch {
			lgo.Trace("skipping object search due to conditions")
			continue
		}

		match, err = s.Comparator.IsMatch(input.Value, aws.StringValue(v))
		lgo.WithError(err).WithField("is_value_match", match).Trace("value match evaluation")
		if err != nil {
			return false, err
		}
		if match {
			return true, nil
		}
	}
	lg.Debug("no matches in single object at all")
	return false, nil
}

func (s *DefaultSearcher[CC, Matcher]) SearchTableData(name string, input *Input, parser ObjParser, lg *log.Entry) ([]*OutputHit, error) {
	var searchables []*OutputHit
	var parsedErr error
	err := s.Client.ScanTable(name, func(items []map[string]*dynamodb.AttributeValue) bool {
		lg.WithField("items", len(items)).Debug("scaning table page items")
		for _, item := range items {
			parsedItem, parsedErr := parser.ParseToStrings(item)
			if parsedErr != nil {
				lg.WithField("fail_fast", input.FailFast).WithError(parsedErr).Warningf("error parsing object to string %#v", item)
				if input.FailFast {
					return false
				}
			}
			match, err := s.SearchSingleObject(input, parsedItem, lg)
			if err != nil {
				lg.WithField("fail_fast", input.FailFast).WithError(parsedErr).Warningf("error searching matches in object %#v", item)
				if input.FailFast {
					return false
				}
			}

			if match {
				hit := &OutputHit{
					TableName:  name,
					HitLevel:   input.Match,
					ObjectData: parsedItem,
				}
				searchables = append(searchables, hit)
			}
		}
		return true
	})

	if parsedErr != nil {
		return nil, parsedErr
	}

	return searchables, err
}
