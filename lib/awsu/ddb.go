package awsu

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	ddbAttr "github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	log "github.com/sirupsen/logrus"
)

type DDBTableDescriber interface {
	IsTableDescribed() bool
	IsGlobalTable() bool
	TableName() string
	GetRawGlobalTableDescriber() *dynamodb.DescribeGlobalTableOutput
	GetRawTableDescriber() *dynamodb.DescribeTableOutput
}

type DDBTableWrapper struct {
	name     string
	isGlobal bool
	gTable   *dynamodb.DescribeGlobalTableOutput
	table    *dynamodb.DescribeTableOutput
}

func NewNamedTableWrapper(name string, isGlobal bool) DDBTableDescriber {
	return &DDBTableWrapper{
		name:     name,
		isGlobal: isGlobal,
	}
}
func NewGlobalTableWrapper(raw *dynamodb.DescribeGlobalTableOutput) DDBTableDescriber {
	return &DDBTableWrapper{
		name:     aws.StringValue(raw.GlobalTableDescription.GlobalTableName),
		isGlobal: raw != nil,
		gTable:   raw,
	}
}

func NewTableWrapper(raw *dynamodb.DescribeTableOutput) DDBTableDescriber {
	return &DDBTableWrapper{
		name:     aws.StringValue(raw.Table.TableName),
		isGlobal: false,
		table:    raw,
	}
}

func (w *DDBTableWrapper) IsTableDescribed() bool {
	return w.gTable != nil || w.table != nil
}

func (w *DDBTableWrapper) IsGlobalTable() bool {
	return w.gTable != nil
}

func (w *DDBTableWrapper) TableName() string {
	return w.name
}

func (w *DDBTableWrapper) GetRawGlobalTableDescriber() *dynamodb.DescribeGlobalTableOutput {
	return w.gTable
}
func (w *DDBTableWrapper) GetRawTableDescriber() *dynamodb.DescribeTableOutput {
	return w.table
}

type DDBApi interface {
	DescribeTable(name string, isGlobal bool) (DDBTableDescriber, error)
	ListAllTables() ([]string, error)
	ListAllGlobalTables() ([]*dynamodb.GlobalTable, error)
	ListCombinedTables(fetchNonGlobal, fetchGlobal bool) ([]DDBTableDescriber, error)
	ScanTable(name string) error
}

type DDBClient struct {
	c *dynamodb.DynamoDB
}

func NewDDBClient(c *dynamodb.DynamoDB) DDBApi {
	return &DDBClient{
		c: c,
	}
}

func (ddb *DDBClient) client() *dynamodb.DynamoDB {
	return ddb.c
}

func (ddb *DDBClient) ScanTable(name string) error {
	c := ddb.client()
	// Example iterating over at most 3 pages of a Scan operation.
	pageNum := 0
	err := c.ScanPages(&dynamodb.ScanInput{
		TableName: aws.String(name),
	},
		func(page *dynamodb.ScanOutput, lastPage bool) bool {
			pageNum++
			for _, item := range page.Items {
				for k, v := range item {
					fmt.Println(k)
					fmt.Println(v)
					//https://docs.aws.amazon.com/sdk-for-go/api/service/dynamodb/dynamodbattribute/
					//ddbAttr.UnmarshalMap(item, )
					_ = ddbAttr.Encoder{}
				}
			}
			//fmt.Println(page)
			return pageNum <= 3
		})
	return err
}
func (ddb *DDBClient) DescribeTable(name string, isGlobal bool) (DDBTableDescriber, error) {
	c := ddb.client()
	if isGlobal {
		gtOut, err := c.DescribeGlobalTable(&dynamodb.DescribeGlobalTableInput{GlobalTableName: aws.String(name)})
		return NewGlobalTableWrapper(gtOut), err
	} else {
		tOut, err := c.DescribeTable(&dynamodb.DescribeTableInput{TableName: aws.String(name)})
		return NewTableWrapper(tOut), err
	}
}

func (ddb *DDBClient) ListCombinedTables(fetchNonGlobal, fetchGlobal bool) ([]DDBTableDescriber, error) {
	if !fetchGlobal && !fetchNonGlobal {
		return nil, fmt.Errorf("must set at least global or non global true")
	}
	all := []DDBTableDescriber{}
	if fetchNonGlobal {
		tables, err := ddb.ListAllTables()
		if err != nil {
			return nil, err
		}
		for _, t := range tables {
			all = append(all, NewNamedTableWrapper(t, false))
		}
	}
	if fetchGlobal {
		tables, err := ddb.ListAllGlobalTables()
		if err != nil {
			return nil, err
		}
		for _, t := range tables {
			all = append(all, NewNamedTableWrapper(aws.StringValue(t.GlobalTableName), true))
		}
	}
	return all, nil
}

func (ddb *DDBClient) ListAllGlobalTables() ([]*dynamodb.GlobalTable, error) {
	tables := []*dynamodb.GlobalTable{}
	var exclusiveStartTableName *string
	for {
		log.WithField("start_table", aws.StringValue(exclusiveStartTableName)).Debug("list ddb tables request")

		out, err := ddb.client().ListGlobalTables(&dynamodb.ListGlobalTablesInput{
			ExclusiveStartGlobalTableName: exclusiveStartTableName,
		})

		if err != nil {
			return nil, err
		}
		if out == nil {
			return nil, fmt.Errorf("list global tables output is nil")
		}

		tables = append(tables, out.GlobalTables...)

		if aws.StringValue(out.LastEvaluatedGlobalTableName) == "" {
			break
		}
		exclusiveStartTableName = out.LastEvaluatedGlobalTableName
	}

	return tables, nil
}
func (ddb *DDBClient) ListAllTables() ([]string, error) {
	tables := []string{}
	var exclusiveStartTableName *string
	for {
		log.WithField("start_table", aws.StringValue(exclusiveStartTableName)).Debug("list ddb tables request")

		out, err := ddb.client().ListTables(&dynamodb.ListTablesInput{
			ExclusiveStartTableName: exclusiveStartTableName,
		})
		if err != nil {
			return nil, err
		}
		if out == nil {
			return nil, fmt.Errorf("list tables output is nil")
		}
		tables = append(tables, aws.StringValueSlice(out.TableNames)...)

		if aws.StringValue(out.LastEvaluatedTableName) == "" {
			break
		}
		exclusiveStartTableName = out.LastEvaluatedTableName
	}

	return tables, nil
}
