package awsu

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/cloudcontrol"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go/aws"
	log "github.com/sirupsen/logrus"
)

type CCResourceDescriber interface {
	IsDescribed() bool
	GetType() *CCResourceProperty
	GetTypeName() string
	GetIdentifier() (string, error)
	GetRawDescribed() *cloudcontrol.GetResourceOutput
	GetRawProperties() string
}

// CCResourceWrapper implementeds CCResourceDescriber interface
type CCResourceWrapper struct {
	RawResource *cloudcontrol.GetResourceOutput
	Type        *CCResourceProperty
}

func NewResourceFromGetOutput(output *cloudcontrol.GetResourceOutput, inputType *CCResourceProperty) CCResourceDescriber {
	return &CCResourceWrapper{
		RawResource: output,
		Type:        inputType,
	}
}

func NewResourceFromListOutput(output *cloudcontrol.ListResourcesOutput, inputType *CCResourceProperty) CCResourceDescriber {
	return nil
	// return &CCResourceWrapper{
	// 	RawResource: output,
	// 	Type:        inputType,
	// }
}

type CCResourcesList struct {
	Resources []CCResourceDescriber
}

func (cc *CCResourceWrapper) GetRawDescribed() *cloudcontrol.GetResourceOutput {
	return cc.RawResource
}

func (cc *CCResourceWrapper) IsDescribed() bool {
	return cc.RawResource != nil
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
	return "", fmt.Errorf("resource not described")
}

func (cc *CCResourceWrapper) GetRawProperties() string {
	if !cc.IsDescribed() {
		return ""
	}
	return aws.StringValue(cc.GetRawDescribed().ResourceDescription.Properties)
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

func NewCloudControlAPI(c *cloudcontrol.Client) CloudControlAPI {
	return &CloudControlClient{
		c:         c,
		Resources: NewCCResources(),
	}
}

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

// TODO: Add Pagination
func (cc *CloudControlClient) ListResources(resource *CCResourceProperty) (*CCResourcesList, error) {
	resp, err := cc.client().ListResources(context.Background(), &cloudcontrol.ListResourcesInput{
		TypeName: aws.String(resource.String()),
	})
	if err != nil {
		return nil, err
	}
	for _, r := range resp.ResourceDescriptions {
		log.Infof("Resource: %s Props %s", *r.Identifier, *r.Properties)
	}
	return nil, nil
	// return NewResourceFromListOutput(resp, resource), nil
}
