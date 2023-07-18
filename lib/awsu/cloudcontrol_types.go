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

// TODO remove this unused
// TODOcloudcontrol) create full list of resources
// func NewCCResources() []*CCResourceProperty {
// 	// hardcoded initialization to get started from  resource types https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-template-resource-type-ref.html
// 	resources := []*CCResourceProperty{}

// 	for _, inf := range newCCServiceInfos() {
// 		if len(inf.DataTypes) > 0 {
// 			sn := inf.ServiceName
// 			sn = strings.ReplaceAll(sn, "Amazon", "")
// 			sn = strings.ReplaceAll(sn, "AWS", "")
// 			sn = strings.ReplaceAll(sn, " ", "")
// 			if strings.Contains(sn, "ApplicationInsights") {
// 				sn = strings.ReplaceAll(sn, "CloudWatch", "")
// 			}
// 			if strings.Contains(sn, "CloudWatchLogs") {
// 				sn = strings.ReplaceAll(sn, "CloudWatch", "")
// 			}
// 			for _, dy := range inf.DataTypes {
// 				resources = append(resources, &CCResourceProperty{
// 					ServiceProvider: "AWS",
// 					ServiceName:     sn,
// 					DataTypeName:    dy,
// 				})
// 			}
// 		}
// 	}
// 	return resources
// }

// type ccServiceInfo struct {
// 	ServiceName string
// 	DataTypes   []string
// }

// func newCCServiceInfos() []*ccServiceInfo {
// 	return []*ccServiceInfo{
// 		{
// 			ServiceName: "AWS Private CA",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Amplify Console",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Amplify UI Builder",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "API Gateway",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "API Gateway V2",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "AppConfig",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Amazon AppFlow",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "AppIntegrations",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Application Auto Scaling",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "App Mesh",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "App Runner",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "AppStream 2.0",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "AWS AppSync",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "ASK",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Athena",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Audit Manager",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "AWS Auto Scaling",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "AWS Backup",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "AWS Batch",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "BillingConductor",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "AWS Budgets",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "AWS Certificate Manager",
// 			DataTypes: []string{
// 				"Account",
// 				"Certificate",
// 			},
// 		},
// 		{
// 			ServiceName: "Chatbot",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "AWS Cloud9",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "CloudFormation",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "CloudFront",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "AWS Cloud Map",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "CloudTrail",
// 			DataTypes: []string{
// 				"Channel",

// 				"EventDataStore",

// 				"ResourcePolicy",

// 				"Trail",
// 			},
// 		},
// 		{
// 			ServiceName: "CloudWatch",
// 			DataTypes: []string{
// 				"Alarm",

// 				"AnomalyDetector",

// 				"CompositeAlarm",

// 				"Dashboard",

// 				"InsightRule",

// 				"MetricStream",
// 			},
// 		},
// 		{
// 			ServiceName: "CloudWatch Application Insights",
// 			DataTypes: []string{
// 				"Application",
// 			},
// 		},
// 		{
// 			ServiceName: "CloudWatch Logs",
// 			DataTypes: []string{
// 				"Destination",

// 				"LogGroup",

// 				"LogStream",

// 				"MetricFilter",

// 				"QueryDefinition",

// 				"ResourcePolicy",

// 				"SubscriptionFilter",
// 			},
// 		},
// 		{
// 			ServiceName: "CloudWatch Synthetics",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "CodeArtifact",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "CodeBuild",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "CodeCommit",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "CodeDeploy",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "CodeGuru Profiler",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "CodeGuru Reviewer",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "CodePipeline",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "AWS CodeStar",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "AWS CodeStar Connections",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "AWS CodeStar Notifications",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Amazon Cognito",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Comprehend",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "AWS Config",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Amazon Connect",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "ConnectCampaigns",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "AWS Control Tower",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Amazon Connect Customer Profiles",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Cost Explorer",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "CUR",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "DataBrew",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Amazon Data Lifecycle Manager",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "AWS Data Pipeline",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "DataSync",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "DAX",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Detective",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "DeviceFarm",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "DevOpsGuru",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "AWS Directory Service",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "AWS DMS",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Amazon DocumentDB",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "DocDBElastic",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "DynamoDB",
// 			DataTypes: []string{
// 				"Table",
// 				"GlobalTable",
// 			},
// 		},
// 		{
// 			ServiceName: "Amazon EC2",
// 			DataTypes: []string{
// 				"CapacityReservation",

// 				"CapacityReservationFleet",

// 				"CarrierGateway",

// 				"ClientVpnAuthorizationRule",

// 				"ClientVpnEndpoint",

// 				"ClientVpnRoute",

// 				"ClientVpnTargetNetworkAssociation",

// 				"CustomerGateway",

// 				"DHCPOptions",

// 				"EC2Fleet",

// 				"EgressOnlyInternetGateway",

// 				"EIP",

// 				"EIPAssociation",

// 				"EnclaveCertificateIamRoleAssociation",

// 				"FlowLog",

// 				"GatewayRouteTableAssociation",

// 				"Host",

// 				"Instance",

// 				"InternetGateway",

// 				"IPAM",

// 				"IPAMAllocation",

// 				"IPAMPool",

// 				"IPAMPoolCidr",

// 				"IPAMResourceDiscovery",

// 				"IPAMResourceDiscoveryAssociation",

// 				"IPAMScope",

// 				"KeyPair",

// 				"LaunchTemplate",

// 				"LocalGatewayRoute",

// 				"LocalGatewayRouteTable",

// 				"LocalGatewayRouteTableVirtualInterfaceGroupAssociation",

// 				"LocalGatewayRouteTableVPCAssociation",

// 				"NatGateway",

// 				"NetworkAcl",

// 				"NetworkAclEntry",

// 				"NetworkInsightsAccessScope",

// 				"NetworkInsightsAccessScopeAnalysis",

// 				"NetworkInsightsAnalysis",

// 				"NetworkInsightsPath",

// 				"NetworkInterface",

// 				"NetworkInterfaceAttachment",

// 				"NetworkInterfacePermission",

// 				"NetworkPerformanceMetricSubscription",

// 				"PlacementGroup",

// 				"PrefixList",

// 				"Route",

// 				"RouteTable",

// 				"SecurityGroup",

// 				"SecurityGroupEgress",

// 				"SecurityGroupIngress",

// 				"SpotFleet",

// 				"Subnet",

// 				"SubnetCidrBlock",

// 				"SubnetNetworkAclAssociation",

// 				"SubnetRouteTableAssociation",

// 				"TrafficMirrorFilter",

// 				"TrafficMirrorFilterRule",

// 				"TrafficMirrorSession",

// 				"TrafficMirrorTarget",

// 				"TransitGateway",

// 				"TransitGatewayAttachment",

// 				"TransitGatewayConnect",

// 				"TransitGatewayMulticastDomain",

// 				"TransitGatewayMulticastDomainAssociation",

// 				"TransitGatewayMulticastGroupMember",

// 				"TransitGatewayMulticastGroupSource",

// 				"TransitGatewayPeeringAttachment",

// 				"TransitGatewayRoute",

// 				"TransitGatewayRouteTable",

// 				"TransitGatewayRouteTableAssociation",

// 				"TransitGatewayRouteTablePropagation",

// 				"TransitGatewayVpcAttachment",

// 				"Volume",

// 				"VolumeAttachment",

// 				"VPC",

// 				"VPCCidrBlock",

// 				"VPCDHCPOptionsAssociation",

// 				"VPCEndpoint",

// 				"VPCEndpointConnectionNotification",

// 				"VPCEndpointService",

// 				"VPCEndpointServicePermissions",

// 				"VPCGatewayAttachment",

// 				"VPCPeeringConnection",

// 				"VPNConnection",

// 				"VPNConnectionRoute",

// 				"VPNGateway",

// 				"VPNGatewayRoutePropagation",
// 			},
// 		},
// 		{
// 			ServiceName: "Amazon EC2 Auto Scaling",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Amazon ECR",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Amazon ECS",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Amazon EFS",
// 			DataTypes: []string{
// 				"AccessPoint",

// 				"FileSystem",

// 				"MountTarget",
// 			},
// 		},
// 		{
// 			ServiceName: "Amazon EKS",
// 			DataTypes: []string{
// 				"Addon",

// 				"Cluster",

// 				"FargateProfile",

// 				"IdentityProviderConfig",

// 				"Nodegroup",
// 			},
// 		},
// 		{
// 			ServiceName: "Elastic Beanstalk",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Elastic Load Balancing",
// 			DataTypes: []string{
// 				"LoadBalancer",
// 			},
// 		},
// 		{
// 			ServiceName: "Elastic Load Balancing V2",
// 			DataTypes: []string{
// 				"Listener",

// 				"ListenerCertificate",

// 				"ListenerRule",

// 				"LoadBalancer",

// 				"TargetGroup",
// 			},
// 		},
// 		{
// 			ServiceName: "Amazon EMR",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Amazon EMR Serverless",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Amazon EMR on EKS",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "ElastiCache",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "EventBridge",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "EventBridge Pipes",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "EventBridge Scheduler",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "EventBridge Schemas",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Evidently",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "FinSpace",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "AWS FIS",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Firewall Manager",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Forecast",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Amazon Fraud Detector",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Amazon FSx",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "GameLift",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Global Accelerator",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "AWS Glue",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Amazon Managed Grafana",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "AWS Ground Station",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "GuardDuty",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "HealthLake",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "IAM",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "IdentityStore",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "IAM Access Analyzer",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Image Builder",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Incident Manager",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Incident Manager Contacts",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Amazon Inspector classic",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Amazon Inspector",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "InternetMonitor",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "AWS IoT",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "AWS IoT 1-Click",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "AWS IoT Analytics",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "AWS IoT Core Device Advisor",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "AWS IoT Events",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "AWS IoT Fleet Hub",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "IoTFleetWise",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "AWS IoT Greengrass",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "AWS IoT Greengrass V2",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "AWS IoT SiteWise",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "AWS IoT TwinMaker",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "AWS IoT Wireless",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Amazon IVS",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Amazon IVS Chat",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Amazon Kendra",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "KendraRanking",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Amazon Keyspaces",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Kinesis",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Kinesis Data Analytics",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Kinesis Data Analytics V2",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Kinesis Data Firehose",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "KinesisVideo",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "AWS KMS",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Lake Formation",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Lambda",
// 			DataTypes: []string{
// 				"Alias",

// 				"CodeSigningConfig",

// 				"EventInvokeConfig",

// 				"EventSourceMapping",

// 				"Function",

// 				"LayerVersion",

// 				"LayerVersionPermission",

// 				"Permission",

// 				"Url",

// 				"Version",
// 			},
// 		},
// 		{
// 			ServiceName: "Lex",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "License Manager",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Lightsail",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Amazon Location",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Lookout for Equipment",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Lookout for Metrics",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Lookout for Vision",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "M2",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Macie",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Amazon Managed Blockchain",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "MediaConnect",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "MediaConvert",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "MediaLive",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "MediaPackage",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "MediaTailor",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "MediaStore",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Amazon MQ",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Amazon MemoryDB",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Amazon MSK",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "KafkaConnect",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "MWAA",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Amazon Neptune",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Network Firewall",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Network Manager",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Amazon Nimble Studio",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "OpenSearch Service (legacy Elasticsearch resource)",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Omics",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "OpenSearch Service",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "OpenSearch Serverless",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "OpsWorks",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "OpsWorks-CM",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "AWS Organizations",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Panorama",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Personalize",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Amazon Pinpoint",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Amazon Pinpoint Email",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Amazon Managed Service for Prometheus",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "QLDB",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Amazon QuickSight",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "AWS RAM",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Amazon RDS",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Amazon Redshift",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "RedshiftServerless",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "RefactorSpaces",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Rekognition",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "ResilienceHub",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "ResourceExplorer2",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Resource Groups",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "AWS RoboMaker",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "RolesAnywhere",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Route 53",
// 			DataTypes: []string{
// 				"CidrCollection",

// 				"DNSSEC",

// 				"HealthCheck",

// 				"HostedZone",

// 				"KeySigningKey",

// 				"RecordSet",

// 				"RecordSetGroup",
// 			},
// 		},
// 		{
// 			ServiceName: "Route 53 Recovery Control",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Route 53 Recovery Readiness",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Route 53 Resolver",
// 			DataTypes: []string{
// 				"FirewallDomainList",

// 				"FirewallRuleGroup",

// 				"FirewallRuleGroupAssociation",

// 				"ResolverConfig",

// 				"ResolverDNSSECConfig",

// 				"ResolverEndpoint",

// 				"ResolverQueryLoggingConfig",

// 				"ResolverQueryLoggingConfigAssociation",

// 				"ResolverRule",

// 				"ResolverRuleAssociation",
// 			},
// 		},
// 		{
// 			ServiceName: "RUM",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Amazon S3",
// 			DataTypes: []string{
// 				"AccessPoint",
// 				"Bucket",
// 				"BucketPolicy",
// 				"MultiRegionAccessPoint",
// 				"MultiRegionAccessPointPolicy",
// 				"StorageLens",
// 			},
// 		},
// 		{
// 			ServiceName: "Amazon S3 Object Lambda",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Amazon S3 on Outposts",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Amazon SageMaker",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Secrets Manager",
// 			DataTypes: []string{
// 				"ResourcePolicy",

// 				"RotationSchedule",

// 				"Secret",

// 				"SecretTargetAttachment",
// 			},
// 		},
// 		{
// 			ServiceName: "AWS Service Catalog",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "AWS Service Catalog AppRegistry",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Security Hub",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Amazon SES",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Amazon SimpleDB",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Signer",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "SimSpace Weaver",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Amazon SNS",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Amazon SQS",
// 			DataTypes: []string{
// 				"Queue",

// 				"QueuePolicy",
// 			},
// 		},
// 		{
// 			ServiceName: "IAM Identity Center",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Step Functions",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Systems Manager",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "AWS Support App",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "SystemsManagerSAP",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Amazon Timestream",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "AWS Transfer Family",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "VoiceID",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "AWS WAF",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "AWS WAF Regional",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "AWS WAF V2",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "Wisdom",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "WorkSpaces",
// 			DataTypes:   []string{},
// 		},
// 		{
// 			ServiceName: "X-Ray",
// 			DataTypes:   []string{},
// 		},
// 	}
// }
