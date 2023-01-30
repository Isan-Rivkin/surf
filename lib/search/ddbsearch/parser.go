package ddbsearch

import (
	"strings"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	ddbAttr "github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	awsu "github.com/isan-rivkin/surf/lib/awsu"
	protoutil "github.com/isan-rivkin/surf/lib/common/proto"
)

type ParserOpt func(tableObj map[string]*dynamodb.AttributeValue, searchObj map[string]*string)

// TODO this is a weak parser since itll present false positives it will parse the wrapper that the sdk ads arround the objects not only the pure data
func WithFmtGoString(schemas map[string]*awsu.DDBSchemaKey, onlySchemaKeys bool, overrideIfExist bool) ParserOpt {
	skipIfVal := []string{"<binary>", "<invalid value>"}
	return func(tableObj map[string]*dynamodb.AttributeValue, searchObj map[string]*string) {
		for k, v := range tableObj {
			_, isSchema := schemas[k]
			if (isSchema && onlySchemaKeys) || !onlySchemaKeys {
				if _, exist := searchObj[k]; !exist || overrideIfExist {
					isNil := v.NULL != nil && *v.NULL == false
					if !isNil {
						if valStr := v.GoString(); valStr != "" {
							containsSkipVals := false
							for _, skipStr := range skipIfVal {
								if strings.Contains(valStr, skipStr) {
									containsSkipVals = true
									break
								}
							}
							if !containsSkipVals {
								searchObj[k] = &valStr
							} else if v.B != nil {
								str := string(v.B)
								searchObj[k] = &str
							}
						}
					}
				}
			}
		}
	}
}

func WithFmtProto(schemas map[string]*awsu.DDBSchemaKey, includeSchemaKeys bool, overrideIfExist bool, delimeter string) ParserOpt {
	return func(tableObj map[string]*dynamodb.AttributeValue, searchObj map[string]*string) {
		for k, v := range tableObj {
			// if not binary format s
			if len(v.B) == 0 && len(v.BS) == 0 {
				return
			}
			// if already exist and not override
			if _, exist := searchObj[k]; exist && !overrideIfExist {
				return
			}

			// if key should be skipped proto parsing
			_, isSchema := schemas[k]
			if !includeSchemaKeys && isSchema {
				return
			}
			// try parse obj
			var bytesSet [][]byte
			if len(v.B) > 0 {
				bytesSet = append(bytesSet, v.B)
			} else if len(v.BS) > 0 {
				bytesSet = v.BS
			}
			accumulated := &protoutil.Accomulator{}
			for _, bs := range bytesSet {
				_ = protoutil.ParseUnknown(bs, accumulated)
				if !accumulated.IsProtoPayload() {
					return
				}
			}
			if accumulated.IsProtoPayload() {
				strVal := accumulated.ToString(delimeter)
				searchObj[k] = &strVal
			}
		}
	}
}

type ObjParserFactory interface {
	New(opts ...ParserOpt) ObjParser
}

type ParserFactory struct {
}

func NewParserFactory() ObjParserFactory {
	return &ParserFactory{}
}

func (f *ParserFactory) New(opts ...ParserOpt) ObjParser {
	return &ObjDefaultParser{
		opts: opts,
	}
}

type ObjParser interface {
	ParseToStrings(tableObj map[string]*dynamodb.AttributeValue) (map[string]*string, error)
}
type ObjDefaultParser struct {
	opts []ParserOpt
}

func NewObjDefaultParser(opts ...ParserOpt) ObjParser {
	return &ObjDefaultParser{opts: opts}
}

func (p *ObjDefaultParser) ParseToStrings(tableObj map[string]*dynamodb.AttributeValue) (map[string]*string, error) {
	searchObj := map[string]*string{}

	for _, opt := range p.opts {
		opt(tableObj, searchObj)
	}

	return searchObj, nil
}

// TODO: unmarshaller is more for known structs, we need string represenatation to searxh so its not a fit
func ____WithFmtAWSMarshaller(schemas map[string]*awsu.DDBSchemaKey, onlySchemaKeys bool, overrideIfExist bool, delimeter string) ParserOpt {
	return func(tableObj map[string]*dynamodb.AttributeValue, searchObj map[string]*string) {
		unmarshalled := map[string]any{}
		ddbAttr.UnmarshalMap(tableObj, unmarshalled)

		for k, v := range tableObj {
			_, isSchema := schemas[k]
			if (isSchema && onlySchemaKeys) || !onlySchemaKeys {
				if _, exist := searchObj[k]; !exist || overrideIfExist {
					panic("UMARSHAL FULLY not imeplemented " + v.GoString())
				}
			}
		}
	}
}
