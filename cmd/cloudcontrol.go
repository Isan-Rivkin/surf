/*
Copyright © 2023 Isan Rivkin isanrivkin@gmail.com
*/
package cmd

import (
	"errors"
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

func setupCommonCloudControlAWSFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&awsProfile, "profile", "p", getDefaultProfileEnvVar(), "~/.aws/credentials chosen account")
	cmd.PersistentFlags().StringVarP(&awsRegion, "region", "r", "", "~/.aws/config default region if empty")
	cloudcontrolCmdMultiAWSProfile = cmd.PersistentFlags().StringArray("aws-session", []string{}, "search in multiple aws profiles & regions (comma separated: --aws-session default,us-east-1 --aws-session dev-account,us-west-2) - overrides --profile and --region")
}
func setupExactTypeCommonFields(cmd *cobra.Command) {
	setupCommonCloudControlAWSFlags(cmd)
	cmd.Flags().String("type", "", "list resource instances of given type (usage: --type AWS::DynamoDB::Table --type AWS::EC2::VPC)")
	cmd.Flags().Bool("exact", false, "exact match resource type no fuzzy matching")
}

func withAdditionalFieldsFlag(cmd *cobra.Command) {
	cmd.Flags().StringArrayP("additional", "a", []string{}, "set additional required fields for search (usage: -a 'ClusterName=my-cluster' -a 'Field2=Value2')")
}
func init() {
	rootCmd.AddCommand(cloudcontrolCmd)
	cloudcontrolCmd.PersistentFlags().Bool("json", false, "Output in JSON format")

	// init list types sub command
	cloudcontrolCmd.AddCommand(cloudcontrolCmdListTypes)

	// init list sub command
	cloudcontrolCmd.AddCommand(cloudcontrolCmdList)
	setupExactTypeCommonFields(cloudcontrolCmdList)
	withAdditionalFieldsFlag(cloudcontrolCmdList)
	// init get sub command
	cloudcontrolCmd.AddCommand(cloudcontrolCmdGet)
	setupExactTypeCommonFields(cloudcontrolCmdGet)
	cloudcontrolCmdGet.Flags().String("id", "", "describe resource via identifier (usage: --id <identifier>)")

	// init search sub command
	cloudcontrolCmd.AddCommand(cloudcontrolCmdSearch)
	cloudcontrolCmdSearch.Flags().StringP("query", "q", "", "search query (usage: --query 'my-vpc')")
	cloudcontrolCmdSearch.Flags().StringArrayP("type", "t", []string{}, "search resource types (usage: -t vpc -t 'ec2')")
	withAdditionalFieldsFlag(cloudcontrolCmdSearch)
	cloudcontrolCmdSearch.Flags().Bool("fail-on-err", false, "Fail on first resources error, otherwise keep searching")
	setupCommonCloudControlAWSFlags(cloudcontrolCmdSearch)
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
		isJson, err := cmd.Flags().GetBool("json")
		if err != nil {
			log.WithError(err).Fatalf("failed getting json flag")
		}
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
		failOnErr, err := cmd.Flags().GetBool("fail-on-err")
		if err != nil {
			log.WithError(err).Fatalf("failed getting fail-on-err")
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
		var allErrs []error
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
							if failOnErr {
								log.WithError(err).Fatalf("failed searching in '%s' q='%s'", matchedType.Resource.String(), q)
							}
							errMsg := fmt.Errorf("resource %s: %w", matchedType.Resource.String(), err)
							if len(auths) > 1 {
								errMsg = fmt.Errorf("auth: %s-%s %w", auth.EffectiveProfile, auth.EffectiveRegion, errMsg)
							}
							if errors.Is(err, awsu.ErrMissingRequiredField) {
								errMsg = fmt.Errorf("try using --additional flag :%w", errMsg)
							}
							allErrs = append(allErrs, errMsg)
							continue
						}
						for _, r := range results {
							allResults = append(allResults, &awsResourceSearchResult{Auth: auth, ResourceType: matchedType.Resource, Resource: r})
						}
					}
				}
				tui.GetLoader().Stop()
			}
			tui.GetLoader().Stop()
		}
		// print as json
		if isJson {
			jsonOutput := map[string]any{
				"errors":  []string{},
				"matches": []map[string]any{},
			}
			for _, r := range allResults {
				rid, _ := r.Resource.GetIdentifier()
				jsonOutput["matches"] = append(jsonOutput["matches"].([]map[string]any), map[string]any{
					"account":  r.Auth.EffectiveProfile + "-" + r.Auth.EffectiveRegion,
					"type":     r.ResourceType.String(),
					"id":       rid,
					"resource": r.Resource.GetRawProperties(),
				})
			}
			for _, e := range allErrs {
				jsonOutput["errors"] = append(jsonOutput["errors"].([]string), e.Error())
			}
			container, err := accessor.NewJsonContainerFromInterface("result", jsonOutput)
			if err != nil {
				log.WithError(err).Fatalf("failed creating json container")
			}
			fmt.Println(container.String())
			return
		}
		// print as table
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
		if len(allErrs) > 0 {
			errRows := []string{"Errors"}
			info := map[string]string{
				"Errors": fmt.Sprintf("summary of %d errors in total", len(allErrs)),
			}
			for idx, e := range allErrs {
				errRows = append(errRows, fmt.Sprintf("%d", idx+1))
				info[fmt.Sprintf("%d", idx+1)] = e.Error()
			}
			tui.GetTable().PrintInfoBox(info, errRows, true)
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
			highScoreMatches := []string{}
			for _, rt := range resourceTypes {
				if rt.Score >= awsu.ServiceMatch {
					highScoreMatches = append(highScoreMatches, rt.Resource.String())
				}
			}
			if len(highScoreMatches) > 1 {
				log.Fatalf("multiple valid matches found, please use exact type: '%s'", strings.Join(highScoreMatches, ", "))
			}
			resourceType := resourceTypes[0].Resource
			result, err := api.GetResource(resourceType, rID)

			if err != nil {
				log.WithError(err).Fatalf("failed getting resource '%s' id '%s'", resourceType, rID)
			}
			// always json output
			fmt.Println(result.GetRawProperties())
		}
	},
}

// add sub command here to list resoure types

var cloudcontrolCmdListTypes = &cobra.Command{
	Use:   "list-types",
	Short: "List Supported resources and additional required fields",
	Long:  `List supported resource types`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		isJson, err := cmd.Flags().GetBool("json")
		if err != nil {
			log.WithError(err).Fatalf("failed getting json flag")
		}
		api := awsu.NewCloudControlAPI(nil)
		schemas := api.GetResourceTypesSchemas()
		jsonOutput := map[string]any{}
		for _, r := range api.ListSupportedResourceTypes() {
			s := schemas[r.String()]
			additionalFields := ""
			if !isJson {
				if len(s.AdditionalRequiredFields) > 0 {
					additionalFields = strings.Join(s.AdditionalRequiredFields, ", ")
					fmt.Printf("%s | Extra_Fields_Required: %s\n", r.String(), additionalFields)
				} else {
					fmt.Println(r.String())
				}
			} else {
				jsonOutput[r.String()] = map[string][]string{
					"additional_required_fields": s.AdditionalRequiredFields,
				}
			}
		}
		container, err := accessor.NewJsonContainerFromInterface("resources", jsonOutput)
		if err != nil {
			log.WithError(err).Fatalf("failed creating json container")
		}
		fmt.Println(container.String())
	},
}

var cloudcontrolCmdList = &cobra.Command{
	Use:   "list --type <resource-type>",
	Short: "List existing cloud resources",
	Long:  `List Existing AWS resources supported`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		isJson, err := cmd.Flags().GetBool("json")
		if err != nil {
			log.WithError(err).Fatalf("failed getting json flag")
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
		allResults := []*awsResourceSearchResult{}
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
			log.Infof("listing matched resource %s", resourceType.String())
			resourceList, err := api.ListResources(resourceType, additionalFieldsMap)
			if err != nil {
				log.WithError(err).Fatalf("failed listing resource %s", inputType)
			}
			for _, r := range resourceList.Resources {
				allResults = append(allResults, &awsResourceSearchResult{Auth: auth, ResourceType: resourceType, Resource: r})
			}
		}
		// print as json
		if isJson {
			jsonOut := []map[string]any{}
			for _, r := range allResults {
				r.Resource.GetRawProperties()
				rid, err := r.Resource.GetIdentifier()
				if err != nil {
					log.WithError(err).Fatalf("failed getting resource identifier")
				}
				properties, err := accessor.NewJsonContainerFromBytes([]byte(r.Resource.GetRawProperties()))
				if err != nil {
					log.WithError(err).Fatalf("failed parsing resource properties")
				}
				jsonOut = append(jsonOut, map[string]any{
					"account":    r.Auth.EffectiveProfile + "-" + r.Auth.EffectiveRegion,
					"type":       r.ResourceType.String(),
					"id":         rid,
					"properties": properties,
				})
			}

			container, err := accessor.NewJsonContainerFromInterface("result", map[string]any{
				"resources": jsonOut,
			})
			if err != nil {
				log.WithError(err).Fatalf("failed creating json container")
			}
			fmt.Println(container.String())
			return
		}
		// print as table
		cols := []string{"Account", "Type", "ID", "Resource"}
		tui := buildTUI()
		for _, r := range allResults {
			rid, err := r.Resource.GetIdentifier()
			if err != nil {
				log.WithError(err).Fatalf("failed getting resource identifier")
			}
			tableInfo := map[string]string{
				"Account":  r.Auth.EffectiveProfile + "-" + r.Auth.EffectiveRegion,
				"Type":     r.ResourceType.String(),
				"ID":       rid,
				"Resource": r.Resource.GetRawProperties(),
			}
			tui.GetTable().PrintInfoBox(tableInfo, cols, true)
		}
	},
}
