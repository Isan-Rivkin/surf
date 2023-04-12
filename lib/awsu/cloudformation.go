package awsu

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	cftypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go/aws"
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
	DescribeResourceType(resource *CCResourceProperty) error
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
		Visibility: cftypes.VisibilityPublic,
		// TODO: also fetch use IMMUTABALE
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

func (cf *CloudFormationClient) DescribeResourceType(resource *CCResourceProperty) error {
	// describe schema and all mutability attributes https://docs.aws.amazon.com/cloudcontrolapi/latest/userguide/supported-resources.html
	// via this api
	// https://docs.aws.amazon.com/cloudcontrolapi/latest/userguide/resource-types.html
	panic("implement me")
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
	return &CloudControlResources{Mutable: mutableTypes, Immutable: immutableTypes}, nil
}
