/*
Copyright © 2023 Isan Rivkin isanrivkin@gmail.com
*/
package cmd

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/isan-rivkin/surf/lib/awsu"
	accessor "github.com/isan-rivkin/surf/lib/common/jsonutil"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	cloudcontrolCmdMultiAWSProfile *[]string
)

// cloudcontrolCmd represents the cloudcontrol command
var cloudcontrolCmd = &cobra.Command{
	Use:   "aws [command]",
	Short: "AWS Resources search",
	Long:  ``,
}

type matchedResourceTypeResult struct {
	Resource *awsu.CCResourceProperty
	Score    awsu.AutoCompleteMatchScore
}

func fuzzyMatchResourceTypes(input string, resourceTypes []*awsu.CCResourceProperty) ([]*matchedResourceTypeResult, error) {
	// var match *awsu.CCResourceProperty
	// var maxScore awsu.AutoCompleteMatchScore
	var matches []*matchedResourceTypeResult
	for _, rt := range resourceTypes {
		score := rt.CheckMatch(strings.ToLower(input))
		if score > 0 {
			matches = append(matches, &matchedResourceTypeResult{Resource: rt, Score: score})
		}
	}
	// sort matches based on max scope to min
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})
	if len(matches) == 0 {
		return nil, fmt.Errorf("no match resource type")
	}
	return matches, nil
}

func setupCommonCloudInitAWSFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&awsProfile, "profile", "p", getDefaultProfileEnvVar(), "~/.aws/credentials chosen account")
	cmd.PersistentFlags().StringVarP(&awsRegion, "region", "r", "", "~/.aws/config default region if empty")
	cloudcontrolCmdMultiAWSProfile = cmd.PersistentFlags().StringArray("aws-session", []string{}, "search in multiple aws profiles & regions (comma separated: --aws-session default,us-east-1 --aws-session dev-account,us-west-2) - overrides --profile and --region")
}
func setupCommonCloudInitFlags(cmd *cobra.Command) {
	setupCommonCloudInitAWSFlags(cmd)
	cmd.Flags().String("type", "", "list resource instances of given type (usage: --type AWS::DynamoDB::Table --type AWS::EC2::VPC)")
	cmd.Flags().Bool("exact", false, "exact match resource type no fuzzy matching")
}

func init() {
	rootCmd.AddCommand(cloudcontrolCmd)

	// init list types sub command
	cloudcontrolCmd.AddCommand(cloudcontrolCmdListTypes)

	// init list sub command
	cloudcontrolCmd.AddCommand(cloudcontrolCmdList)
	setupCommonCloudInitFlags(cloudcontrolCmdList)
	cloudcontrolCmdList.Flags().StringArrayP("additional", "a", []string{}, "set additional required fields for search (usage: -p 'ClusterName=my-cluster' -p 'Field2=Value2')")
	// init get sub command
	cloudcontrolCmd.AddCommand(cloudcontrolCmdGet)
	setupCommonCloudInitFlags(cloudcontrolCmdGet)
	cloudcontrolCmdGet.Flags().String("id", "", "describe resource via identifier (usage: --id <identifier>)")

	// init search sub command
	cloudcontrolCmd.AddCommand(cloudcontrolCmdSearch)
	cloudcontrolCmdSearch.Flags().StringP("query", "q", "", "search query (usage: --query 'my-vpc')")
	cloudcontrolCmdSearch.Flags().StringArrayP("type", "t", []string{}, "search resource types (usage: -t vpc -t 'ec2')")
	cloudcontrolCmdSearch.Flags().StringArrayP("additional", "a", []string{}, "set additional required fields for search (usage: -p 'ClusterName=my-cluster' -p 'Field2=Value2')")
	setupCommonCloudInitAWSFlags(cloudcontrolCmdSearch)
}

func searchResourceInstance(api awsu.CloudControlAPI, query string, resourceType *awsu.CCResourceProperty, additionalFields map[string]string) ([]awsu.CCResourceDescriber, error) {
	var results []awsu.CCResourceDescriber
	resourceList, err := api.ListResources(resourceType, additionalFields)
	if err != nil {
		return nil, fmt.Errorf("listing resource %s: %w", resourceType.String(), err)
	}
	for _, r := range resourceList.Resources {
		id, err := r.GetIdentifier()
		if err != nil {
			return nil, fmt.Errorf("parsing identifier %s: %w", resourceType.String(), err)
		}
		// check if q (regex) match id or name
		rid, err := r.GetIdentifier()
		if err != nil {
			return nil, fmt.Errorf("getting resource identifier %s: %w", resourceType.String(), err)
		}

		// try describe and match properties
		describedResource, err := api.GetResource(resourceType, rid)

		if err != nil {
			return nil, fmt.Errorf("getting resource %s: %w", rid, err)
		}

		obj, err := accessor.NewJsonContainerFromBytes([]byte(describedResource.GetRawProperties()))
		if err != nil {
			return nil, fmt.Errorf("parsing properties: %w", err)
		}
		flat, err := obj.Flatten()
		if err != nil {
			return nil, fmt.Errorf("flattening properties: %w", err)
		}

		for _, v := range flat {
			if strVal, ok := v.(string); ok {
				matched, err := regexp.MatchString(query, strVal)
				if err != nil {
					return nil, fmt.Errorf("matching regex %s: %w", query, err)
				}
				if matched {
					log.WithFields(log.Fields{"id": id, "properties": obj.String()}).Debug("found resource")
					results = append(results, describedResource)
					break
				}
			}
		}
	}
	return results, nil
}

type awsResourceSearchResult struct {
	Auth         *awsu.AuthInput
	ResourceType *awsu.CCResourceProperty
	Resource     awsu.CCResourceDescriber
}

// add sub command for search
var cloudcontrolCmdSearch = &cobra.Command{
	Use:   "search",
	Short: "Search existing cloud resources",
	Long: `Search Existing AWS resources supported
surf aws search  -q 'my-prod'  -t vpc -t eks::cluster -a 'ClusterName=lakefs-cloud'
	`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		tui := buildTUI()
		defer tui.GetLoader().Stop()
		q, err := cmd.Flags().GetString("query")
		if err != nil {
			log.WithError(err).Fatalf("failed getting query")
		}
		if q == "" {
			log.Fatalf("must specify a query --query (see --help)")
		}
		inputTypes, err := cmd.Flags().GetStringArray("type")
		if err != nil {
			log.WithError(err).Fatalf("failed getting types")
		}
		if len(inputTypes) == 0 {
			log.Fatalf("must specify a resource types --type (see --help)")
		}
		additionalFields, err := cmd.Flags().GetStringArray("additional")
		if err != nil {
			log.WithError(err).Fatalf("failed getting additional fields")
		}
		additionalFieldsMap := map[string]string{}
		for _, f := range additionalFields {
			parts := strings.Split(f, "=")
			if len(parts) != 2 {
				log.Fatalf("invalid additional field %s, must be in format 'key=value'", f)
			}
			additionalFieldsMap[parts[0]] = parts[1]
		}
		sessInputs, err := resolveAWSSessions(cloudcontrolCmdMultiAWSProfile, awsProfile, awsRegion)
		if err != nil {
			log.WithError(err).Fatalf("failed building input for AWS session")
		}
		auths, err := awsu.NewSessionInputMatrix(sessInputs)

		if err != nil {
			log.WithError(err).Fatalf("failed creating session in AWS")
		}
		var allResults []*awsResourceSearchResult
		for _, auth := range auths {
			ccClient, err := awsu.NewCloudControl(auth)

			if err != nil {
				tui.GetLoader().Stop()
				log.WithError(err).Fatalf("failed creating cloudcontrol client")
			}

			api := awsu.NewCloudControlAPI(ccClient)
			for _, inputType := range inputTypes {
				tui.GetLoader().Stop()
				matchedTypes, err := fuzzyMatchResourceTypes(inputType, api.ListSupportedResourceTypes())
				if err != nil {
					log.WithError(err).Warnf("failed matching '%s' resource, skipping", inputType)
					continue
				}
				for _, matchedType := range matchedTypes {
					tui.GetLoader().Stop()
					if matchedType.Score >= awsu.ServiceMatch {
						tui.GetLoader().Start(fmt.Sprintf("searching in '%s' q='%s'", matchedType.Resource.String(), q), "", "green")
						results, err := searchResourceInstance(api, q, matchedType.Resource, additionalFieldsMap)
						if err != nil {
							tui.GetLoader().Stop()
							log.WithError(err).Fatalf("failed searching in '%s' q='%s'", matchedType.Resource.String(), q)
						}
						for _, r := range results {
							allResults = append(allResults, &awsResourceSearchResult{Auth: auth, ResourceType: matchedType.Resource, Resource: r})
						}
					}
				}
				tui.GetLoader().Stop()
			}
			tui.GetLoader().Stop()
			cols := []string{"Account", "Type", "ID", "Resource"}
			for _, r := range allResults {
				rid, _ := r.Resource.GetIdentifier()
				info := map[string]string{
					"Account":  r.Auth.EffectiveProfile + "-" + r.Auth.EffectiveRegion,
					"Type":     r.ResourceType.String(),
					"ID":       rid,
					"Resource": r.Resource.GetRawProperties(),
				}
				tui.GetTable().PrintInfoBox(info, cols, true)
			}
		}
	},
}

// add sub command for get
var cloudcontrolCmdGet = &cobra.Command{
	Use:   "get --type <resource-type> --id <resource-id>",
	Short: "Get existing cloud resources",
	Long:  `Get Existing AWS resources supported`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		sessInputs, err := resolveAWSSessions(cloudcontrolCmdMultiAWSProfile, awsProfile, awsRegion)
		if err != nil {
			log.WithError(err).Fatalf("failed building input for AWS session")
		}
		auths, err := awsu.NewSessionInputMatrix(sessInputs)

		if err != nil {
			log.WithError(err).Fatalf("failed creating session in AWS")
		}

		for _, auth := range auths {
			ccClient, err := awsu.NewCloudControl(auth)

			if err != nil {
				log.WithError(err).Fatalf("failed creating cloudcontrol client")
			}

			api := awsu.NewCloudControlAPI(ccClient)

			inputType, _ := cmd.Flags().GetString("type")
			if inputType == "" {
				log.Fatalf("must specify a resource type --type (usage: --type AWS::DynamoDB::Table --type AWS::EC2::VPC, see --help)")
			}

			rID, _ := cmd.Flags().GetString("id")
			if rID == "" {
				log.Fatalf("must specify a resource identifier --id")
			}

			resourceTypes, err := fuzzyMatchResourceTypes(inputType, api.ListSupportedResourceTypes())
			if err != nil {
				log.WithError(err).Fatalf("failed matching resources")
			}
			resourceType := resourceTypes[0].Resource
			result, err := api.GetResource(resourceType, rID)

			if err != nil {
				log.WithError(err).Fatalf("failed getting resource %s", rID)
			}
			fmt.Println(result.GetRawProperties())
		}
	},
}

// add sub command here to list resoure types

var cloudcontrolCmdListTypes = &cobra.Command{
	Use:   "list-types",
	Short: "List Supported resources",
	Long:  `List supported resource types`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		api := awsu.NewCloudControlAPI(nil)
		for _, r := range api.ListSupportedResourceTypes() {
			fmt.Println(r)
		}
	},
}

var cloudcontrolCmdList = &cobra.Command{
	Use:   "list --type <resource-type>",
	Short: "List existing cloud resources",
	Long:  `List Existing AWS resources supported`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		additionalFields, err := cmd.Flags().GetStringArray("additional")
		if err != nil {
			log.WithError(err).Fatalf("failed getting additional fields")
		}
		additionalFieldsMap := map[string]string{}
		for _, f := range additionalFields {
			parts := strings.Split(f, "=")
			if len(parts) != 2 {
				log.Fatalf("invalid additional field %s, must be in format 'key=value'", f)
			}
			additionalFieldsMap[parts[0]] = parts[1]
		}
		sessInputs, err := resolveAWSSessions(cloudcontrolCmdMultiAWSProfile, awsProfile, awsRegion)
		if err != nil {
			log.WithError(err).Fatalf("failed building input for AWS session")
		}
		auths, err := awsu.NewSessionInputMatrix(sessInputs)

		if err != nil {
			log.WithError(err).Fatalf("failed creating session in AWS")
		}

		for _, auth := range auths {
			ccClient, err := awsu.NewCloudControl(auth)

			if err != nil {
				log.WithError(err).Fatalf("failed creating cloudcontrol client")
			}

			api := awsu.NewCloudControlAPI(ccClient)

			inputType, _ := cmd.Flags().GetString("type")
			if inputType == "" {
				log.Fatalf("must specify a resource type --type (usage: --type AWS::DynamoDB::Table --type AWS::EC2::VPC, see --help)")
			}

			resourceTypes, err := fuzzyMatchResourceTypes(inputType, api.ListSupportedResourceTypes())
			if err != nil {
				log.WithError(err).Fatalf("failed matching resources")
			}
			resourceType := resourceTypes[0].Resource
			fmt.Printf("listing matched resource %s\n", resourceType.String())
			resourceList, err := api.ListResources(resourceType, additionalFieldsMap)
			if err != nil {
				log.WithError(err).Fatalf("failed listing resource %s", inputType)
			}
			for _, r := range resourceList.Resources {
				id, err := r.GetIdentifier()
				if err != nil {
					panic(fmt.Errorf("parsing identifier %s", err))
				}
				fmt.Printf("ID: %s\nProperties: %s\n", id, r.GetRawProperties())
			}
		}
	},
}
