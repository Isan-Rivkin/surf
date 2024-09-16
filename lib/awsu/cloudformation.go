package awsu

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	cftypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go/aws"
	accessor "github.com/isan-rivkin/surf/lib/common/jsonutil"
)

const CCNameDelimeter = "::"

type CloudControlResources struct {
	Mutable   []cftypes.TypeSummary
	Immutable []cftypes.TypeSummary
}

func (cr *CloudControlResources) GetResources() ([]*CCResourceProperty, error) {
	// TODO:: for now skip non AWS providers in AWS such as Fastly, Databricks etc )
	var resources []*CCResourceProperty
	for _, m := range cr.Mutable {
		fqdn := aws.StringValue(m.TypeName)
		splitted := strings.Split(fqdn, CCNameDelimeter)
		if splitted[0] != "AWS" {
			continue
		}
		resources = append(resources, &CCResourceProperty{
			ServiceProvider:          splitted[0],
			ServiceName:              splitted[1],
			DataTypeName:             splitted[2],
			ResourceProvisioningType: string(cftypes.ProvisioningTypeFullyMutable),
		})
	}
	for _, m := range cr.Immutable {
		fqdn := aws.StringValue(m.TypeName)
		splitted := strings.Split(fqdn, CCNameDelimeter)
		if splitted[0] != "AWS" {
			continue
		}
		resources = append(resources, &CCResourceProperty{
			ServiceProvider:          splitted[0],
			ServiceName:              splitted[1],
			DataTypeName:             splitted[2],
			ResourceProvisioningType: string(cftypes.ProvisioningTypeImmutable),
		})
	}
	return resources, nil
}

// CloudFormationAPI used as a utility for CloudControl
// https://docs.aws.amazon.com/cloudcontrolapi/latest/userguide/resource-types.html#resource-types-determine-support
// / CloudControl API supported resources https://docs.aws.amazon.com/cloudcontrolapi/latest/userguide/supported-resources.html
type CloudFormationAPI interface {
	GetAllSupportedCloudControlAPIResources() (*CloudControlResources, error)
	DescribeResourceType(resource *CCResourceProperty) (*cloudformation.DescribeTypeOutput, error)
}

func NewCloudFormationAPI(c *cloudformation.Client) CloudFormationAPI {
	return &CloudFormationClient{
		c: c,
	}
}

type CloudFormationClient struct {
	c *cloudformation.Client
}

func (cf *CloudFormationClient) client() *cloudformation.Client {
	return cf.c
}

func (cf *CloudFormationClient) getTypes(pType cftypes.ProvisioningType) ([]cftypes.TypeSummary, error) {
	var result []cftypes.TypeSummary
	paginator := cloudformation.NewListTypesPaginator(cf.client(), &cloudformation.ListTypesInput{
		Visibility:       cftypes.VisibilityPublic,
		ProvisioningType: pType,
	})

	for paginator.HasMorePages() {

		output, err := paginator.NextPage(context.TODO())
		if err != nil {
			return nil, err
		}
		result = append(result, output.TypeSummaries...)
	}
	return result, nil
}

var (
	ErrCloudFormationRateLimit error = errors.New("Rate exceeded")
)

// Some cloud control resources require additional information when descriving them later in surf so with describe we can extract the additional fields
// aws cloudformation describe-type --type RESOURCE --type-name AWS::EKS::Addon | jq .Schema | jq -r  | jq .
// https://docs.aws.amazon.com/cloudcontrolapi/latest/userguide/resource-operations-list.html#resource-operations-list-containers
// API https://docs.aws.amazon.com/cloudcontrolapi/latest/userguide/resource-types.html
func (cf *CloudFormationClient) DescribeResourceType(resource *CCResourceProperty) (*cloudformation.DescribeTypeOutput, error) {
	res, err := cf.client().DescribeType(context.Background(), &cloudformation.DescribeTypeInput{
		Type:     cftypes.RegistryTypeResource,
		TypeName: aws.String(resource.String()),
	})
	if err != nil {
		if strings.Contains(err.Error(), "Throttling: Rate exceeded") {
			return nil, ErrCloudFormationRateLimit
		}
		return nil, fmt.Errorf("aws describing resource type '%s': %w", resource.String(), err)
	}
	return res, nil
}

func NewResourceSchemaFromDescribe(resource *CCResourceProperty, described *cloudformation.DescribeTypeOutput) (*ResourceSchema, error) {
	schema, err := accessor.NewJsonContainerFromBytes([]byte(aws.StringValue(described.Schema)))
	if err != nil {
		return nil, fmt.Errorf("parsing schema for resource '%s': %w", resource.String(), err)
	}
	additionalRequiredKeyNames, ok := accessor.GetArray(schema, "handlers.list.handlerSchema.required")
	if !ok {
		// no additional required keys
		return &ResourceSchema{
			RawSchemaJson:            aws.StringValue(described.Schema),
			AdditionalRequiredFields: []string{},
			TypeIdentifier:           resource.String(),
		}, nil
	}
	var additionalRequiredFields []string
	for _, keyObj := range additionalRequiredKeyNames {
		key, ok := accessor.GetValue[string](keyObj, "")
		if ok {
			additionalRequiredFields = append(additionalRequiredFields, key)
		}
	}
	return &ResourceSchema{
		TypeIdentifier:           resource.String(),
		RawSchemaJson:            aws.StringValue(described.Schema),
		AdditionalRequiredFields: additionalRequiredFields,
	}, nil
}

func (cf *CloudFormationClient) GetAllSupportedCloudControlAPIResources() (*CloudControlResources, error) {
	// aws cloudformation list-types --type RESOURCE --visibility PUBLIC --provisioning-type FULLY_MUTABLE --max-results 100
	mutableTypes, err := cf.getTypes(cftypes.ProvisioningTypeFullyMutable)
	if err != nil {
		return nil, err
	}
	immutableTypes, err := cf.getTypes(cftypes.ProvisioningTypeImmutable)
	if err != nil {
		return nil, err
	}
	resources := &CloudControlResources{Mutable: mutableTypes, Immutable: immutableTypes}
	return resources, nil
}
