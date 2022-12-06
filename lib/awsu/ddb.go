package awsu

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	log "github.com/sirupsen/logrus"
)

type DDBApi interface {
	ListAllTables() ([]string, error)
	ListAllGlobalTables() ([]*dynamodb.GlobalTable, error)
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