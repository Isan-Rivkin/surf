package awsu

import (
	"fmt"
	"strings"

	cftypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	cfg "github.com/isan-rivkin/surf/lib/awsu/cloudformationgenerated"
)

type AutoCompleteMatchScore int

const (
	ProviderMatch AutoCompleteMatchScore = 10
	ServiceMatch  AutoCompleteMatchScore = 40
	DataTypeMatch AutoCompleteMatchScore = 50
	ExactMatch    AutoCompleteMatchScore = 100
)

type CCSupportedResources interface {
	GetResources() ([]*CCResourceProperty, error)
}

/*
	"properties": {
		"RoleArn": {
			"description": "The IAM role ARN that the pod identity association is created for.",
			"pattern": "^arn:aws(-cn|-us-gov|-iso(-[a-z])?)?:iam::\\d{12}:(role)\\/*",
			"type": "string"
		},
		"ServiceAccount": {
			"description": "The Kubernetes service account that the pod identity association is created for.",
			"type": "string"
		}
	},
*/
// Define the main struct
type ResourceSchema struct {
	TypeIdentifier           string
	RawSchemaJson            string   `json:"rawSchemaJson"`
	AdditionalRequiredFields []string `json:"additionalRequiredFields"`
}

// resource types https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-template-resource-type-ref.html
type CCResourceProperty struct {
	//service-provider::service-name::data-type-name
	ServiceProvider          string `json:"service_provider"`
	ServiceName              string `json:"service_name"`
	DataTypeName             string `json:"data_type_name"`
	ResourceProvisioningType string
}

func (cc *CCResourceProperty) String() string {
	return fmt.Sprintf("%s%s%s%s%s", cc.ServiceProvider, CCNameDelimeter, cc.ServiceName, CCNameDelimeter, cc.DataTypeName)
}

func (cc *CCResourceProperty) ProvisioningType() string {
	return cc.ResourceProvisioningType
}
func (cc *CCResourceProperty) IsMutable() bool {
	return cc.ResourceProvisioningType == string(cftypes.ProvisioningTypeFullyMutable)
}

// TODO: use intelligent matching based on distance
// https://github.com/adrg/strutil
func (cc *CCResourceProperty) CheckMatch(input string) AutoCompleteMatchScore {
	score := 0
	parts := strings.Split(input, CCNameDelimeter)
	for _, p := range parts {
		score += cc.evalPartMatch(p)
	}
	// weight parts matches as well if 2 parts given and only 1 match that's lower score than 1 part given and 1 match
	score = score / len(parts)
	return AutoCompleteMatchScore(score)
}

func (cc *CCResourceProperty) evalPartMatch(input string) int {
	score := 0
	if input == strings.ToLower(cc.ServiceProvider) {
		score += int(ProviderMatch)
	}
	if input == strings.ToLower(cc.ServiceName) {
		score += int(ServiceMatch)
	}
	if input == strings.ToLower(cc.DataTypeName) {
		score += int(DataTypeMatch)
	}
	return score
}

func NewCloudControlResourcesFromGeneratedCode() []*CCResourceProperty {
	resources := []*CCResourceProperty{}
	for _, p := range cfg.GenCloudformationProperties {
		rp := CCResourceProperty(p)
		resources = append(resources, &rp)
	}
	return resources
}

func NewResourceSchemaFromGeneratedCode() map[string]ResourceSchema {
	resources := map[string]ResourceSchema{}
	for k, v := range cfg.GenCloudFormationResourceSchemas {
		resources[k] = ResourceSchema(v)
	}
	return resources
}
