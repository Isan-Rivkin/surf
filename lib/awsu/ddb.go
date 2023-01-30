package awsu

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	log "github.com/sirupsen/logrus"
)

type DDBAttributesHandler = func([]map[string]*dynamodb.AttributeValue) bool

type DDBSchemaKey struct {
	Name string
	// S (string),N (number) ,B (binary)
	KeyType string
	// HASH - partition key
	// RANGE - sort key
	KeyRole string
}
type DDBTableDescriber interface {
	IsTableStatusOK() bool
	IsTableDescribed() bool
	IsGlobalTable() bool
	TableName() string
	GetRawGlobalTableDescriber() *dynamodb.DescribeGlobalTableOutput
	GetRawTableDescriber() *dynamodb.DescribeTableOutput
	GetSchemaDefinitions() (map[string]*DDBSchemaKey, error)
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

func (w *DDBTableWrapper) IsTableStatusOK() bool {
	if w.IsTableDescribed() {
		s := aws.StringValue(w.GetRawTableDescriber().Table.TableStatus)
		return s == "ACTIVE" || s == "UPDATING"
	}
	return false
}
func (w *DDBTableWrapper) GetSchemaDefinitions() (map[string]*DDBSchemaKey, error) {
	if !w.IsTableDescribed() {
		return nil, fmt.Errorf("table %s not described ", w.TableName())
	}
	//NOTE: global desciption is irrelevant we only care for the replica info in schema key
	desc := w.GetRawTableDescriber()
	if desc == nil {
		return nil, fmt.Errorf("table %s descriptor is nil", w.TableName())
	}

	result := map[string]*DDBSchemaKey{}

	for _, kschem := range desc.Table.KeySchema {
		name := aws.StringValue(kschem.AttributeName)
		result[name] = &DDBSchemaKey{
			Name:    name,
			KeyRole: aws.StringValue(kschem.KeyType),
		}
	}
	for _, attr := range desc.Table.AttributeDefinitions {
		name := aws.StringValue(attr.AttributeName)
		result[name].KeyType = aws.StringValue(attr.AttributeType)
	}
	return result, nil
}

type DDBApi interface {
	DescribeTable(name string, isGlobal bool) (DDBTableDescriber, error)
	ListAllTables() ([]string, error)
	ListAllGlobalTables() ([]*dynamodb.GlobalTable, error)
	ListCombinedTables(fetchNonGlobal, fetchGlobal bool) ([]DDBTableDescriber, error)
	ScanTable(name string, pageHandler DDBAttributesHandler) error
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

func (ddb *DDBClient) ScanTable(name string, pageHandler DDBAttributesHandler) error {
	c := ddb.client()
	err := c.ScanPages(
		&dynamodb.ScanInput{
			TableName: aws.String(name),
		},
		func(page *dynamodb.ScanOutput, lastPage bool) bool {
			return pageHandler(page.Items)
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
		log.WithField("start_table", aws.StringValue(exclusiveStartTableName)).Debug("list global ddb tables request")

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
