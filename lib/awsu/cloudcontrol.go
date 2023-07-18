package awsu

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/cloudcontrol"
	cctypes "github.com/aws/aws-sdk-go-v2/service/cloudcontrol/types"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go/aws"
	log "github.com/sirupsen/logrus"
)

type CCResourceDescriber interface {
	IsDescribed() bool
	IsShallowDescribe() bool
	GetType() *CCResourceProperty
	GetTypeName() string
	GetIdentifier() (string, error)
	GetRawDescribed() *cloudcontrol.GetResourceOutput
	GetRawShallowDescribed() *cctypes.ResourceDescription
	GetRawProperties() string
}

// CCResourceWrapper implementeds CCResourceDescriber interface
type CCResourceWrapper struct {
	RawResource     *cloudcontrol.GetResourceOutput
	RawResourceList *cctypes.ResourceDescription
	Type            *CCResourceProperty
}

func NewResourceFromGetOutput(output *cloudcontrol.GetResourceOutput, inputType *CCResourceProperty) CCResourceDescriber {
	return &CCResourceWrapper{
		RawResource: output,
		Type:        inputType,
	}

}
func NewResourceFromListOutput(output *cloudcontrol.ListResourcesOutput, inputType *CCResourceProperty) []CCResourceDescriber {
	var resources []CCResourceDescriber
	for _, r := range output.ResourceDescriptions {
		capture := r
		resources = append(resources, &CCResourceWrapper{
			RawResourceList: &capture,
			Type:            inputType,
		})
	}
	return resources
}

type CCResourcesList struct {
	Resources []CCResourceDescriber
}

func (cc *CCResourceWrapper) GetRawDescribed() *cloudcontrol.GetResourceOutput {
	return cc.RawResource
}

func (cc *CCResourceWrapper) GetRawShallowDescribed() *cctypes.ResourceDescription {
	return cc.RawResourceList
}

func (cc *CCResourceWrapper) IsDescribed() bool {
	return cc.RawResource != nil
}

func (cc *CCResourceWrapper) IsShallowDescribe() bool {
	return cc.RawResourceList != nil
}

func (cc *CCResourceWrapper) GetType() *CCResourceProperty {
	return cc.Type
}
func (cc *CCResourceWrapper) GetTypeName() string {
	if cc.IsDescribed() {
		return aws.StringValue(cc.RawResource.TypeName)
	}
	return cc.Type.String()
}

func (cc *CCResourceWrapper) GetIdentifier() (string, error) {
	if cc.IsDescribed() {
		return aws.StringValue(cc.RawResource.ResourceDescription.Identifier), nil
	}
	if cc.IsShallowDescribe() {
		return aws.StringValue(cc.RawResourceList.Identifier), nil
	}
	return "", fmt.Errorf("resource not described")
}

func (cc *CCResourceWrapper) GetRawProperties() string {
	if cc.IsDescribed() {
		return aws.StringValue(cc.GetRawDescribed().ResourceDescription.Properties)
	}
	if cc.IsShallowDescribe() {
		return aws.StringValue(cc.RawResourceList.Properties)
	}
	return ""
}

type CloudControlAPI interface {
	ListResources(resource *CCResourceProperty) (*CCResourcesList, error)
	GetResource(resource *CCResourceProperty, identifier string) (CCResourceDescriber, error)
	ListSupportedResourceTypes() []*CCResourceProperty
}

type CloudControlClient struct {
	c         *cloudcontrol.Client
	Resources []*CCResourceProperty
}

func NewCloudControlAPI(c *cloudcontrol.Client) CloudControlAPI {
	return &CloudControlClient{
		c:         c,
		Resources: NewCloudControlResourcesFromGeneratedCode(),
	}
}

func NewCloudControlAPIWithDynamicResources(c *cloudcontrol.Client, cf *cloudformation.Client) CloudControlAPI {
	// TODO unify Ctor of CC API no need for all this here,  make resources external dependency
	resp, err := NewCloudFormationAPI(cf).GetAllSupportedCloudControlAPIResources()
	if err != nil {
		panic(err)
	}
	resources, err := resp.GetResources()
	if err != nil {
		panic(err)
	}
	return &CloudControlClient{
		c:         c,
		Resources: resources,
	}
}

// TODO: remove this and usage of NewCCResources
// func NewCloudControlAPI(c *cloudcontrol.Client) CloudControlAPI {
// 	return &CloudControlClient{
// 		c:         c,
// 		Resources: NewCCResources(),
// 	}
// }

func (cc *CloudControlClient) client() *cloudcontrol.Client {
	return cc.c
}

func (cc *CloudControlClient) ListSupportedResourceTypes() []*CCResourceProperty {
	return cc.Resources
}

// Get Resource from Cloud Control API by Resource Type and Identifier (ARN) with paging
func (cc *CloudControlClient) GetResource(resource *CCResourceProperty, identifier string) (CCResourceDescriber, error) {
	resp, err := cc.client().GetResource(context.Background(), &cloudcontrol.GetResourceInput{
		TypeName:   aws.String(resource.String()),
		Identifier: aws.String(identifier),
	})
	if err != nil {
		return nil, err
	}
	return NewResourceFromGetOutput(resp, resource), nil
}

// TODO: make this function really return the type not nils
func (cc *CloudControlClient) ListResources(resource *CCResourceProperty) (*CCResourcesList, error) {
	input := &cloudcontrol.ListResourcesInput{
		TypeName: aws.String(resource.String()),
	}
	var result []CCResourceDescriber
	for {
		resp, err := cc.client().ListResources(context.Background(), input)
		if err != nil {
			return nil, fmt.Errorf("listing resources: %w", err)
		}
		result = append(result, NewResourceFromListOutput(resp, resource)...)
		for _, r := range resp.ResourceDescriptions {
			log.Debugf("Resource: %s Props %s", *r.Identifier, *r.Properties)
		}
		if resp.NextToken == nil {
			break
		}
		input.NextToken = resp.NextToken
	}
	return &CCResourcesList{Resources: result}, nil
}
