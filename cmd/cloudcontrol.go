/*
Copyright © 2023 Isan Rivkin isanrivkin@gmail.com
*/
package cmd

import (
	"fmt"
	"regexp"
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

func fuzzyMatchResourceTypes(input string, resourceTypes []*awsu.CCResourceProperty) (*awsu.CCResourceProperty, error) {
	var match *awsu.CCResourceProperty
	var maxScore awsu.AutoCompleteMatchScore

	for _, rt := range resourceTypes {
		score := rt.CheckMatch(strings.ToLower(input))
		if score == awsu.ExactMatch {
			return rt, nil
		}
		if score > maxScore {
			maxScore = score
			match = rt
		}
	}
	if match == nil {
		return nil, fmt.Errorf("no match resource type")
	}
	return match, nil
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

	// init get sub command
	cloudcontrolCmd.AddCommand(cloudcontrolCmdGet)
	setupCommonCloudInitFlags(cloudcontrolCmdGet)
	cloudcontrolCmdGet.Flags().String("id", "", "describe resource via identifier (usage: --id <identifier>)")

	// init search sub command
	cloudcontrolCmd.AddCommand(cloudcontrolCmdSearch)
	cloudcontrolCmdSearch.Flags().StringP("query", "q", "", "search query (usage: --query 'my-vpc')")
	cloudcontrolCmdSearch.Flags().StringArrayP("type", "t", []string{}, "search resource types (usage: -t vpc -t 'ec2')")
	setupCommonCloudInitAWSFlags(cloudcontrolCmdSearch)
}

// add sub command for search
var cloudcontrolCmdSearch = &cobra.Command{
	Use:   "search",
	Short: "Search existing cloud resources",
	Long:  `Search Existing AWS resources supported`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
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
		for _, i := range inputTypes {
			fmt.Println("searching for resource type", i)
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
			for _, inputType := range inputTypes {
				resourceType, err := fuzzyMatchResourceTypes(inputType, api.ListSupportedResourceTypes())
				if err != nil {
					log.WithError(err).Warnf("failed matching '%s' resource, skipping", inputType)
					continue
				}

				fmt.Printf("searching for resource '%s' q='%s'\n", resourceType.String(), q)

				resourceList, err := api.ListResources(resourceType)
				if err != nil {
					log.WithError(err).Fatalf("failed listing resource %s", inputType)
				}
				for _, r := range resourceList.Resources {
					id, err := r.GetIdentifier()
					if err != nil {
						panic(fmt.Errorf("failed parsing identifier %s", err))
					}
					// check if q (regex) match id or name
					rid, err := r.GetIdentifier()
					if err != nil {
						log.WithError(err).Fatalf("failed getting resource identifier")
					}
					matchedID, err := regexp.MatchString(q, rid)
					if err != nil {
						log.WithError(err).Fatalf("failed matching regex %s", q)
					}
					if matchedID {
						fmt.Printf("ID: %s\nProperties: %s\n", id, r.GetRawProperties())
						continue
					}
					// try describe and match properties
					result, err := api.GetResource(resourceType, rid)

					if err != nil {
						log.WithError(err).Fatalf("failed getting resource %s", rid)
					}

					// fmt.Printf("NO MATCH: %s\nProperties: %s\n", id, result.GetRawProperties())
					// convert result.GetRawProperties() into map[any]any and iterate each key nested
					// if any key value matches q then print

					obj, err := accessor.NewJsonContainerFromBytes([]byte(result.GetRawProperties()))
					if err != nil {
						log.WithError(err).Fatalf("failed parsing properties")
					}
					flat, err := obj.Flatten()
					if err != nil {
						log.WithError(err).Fatalf("failed flattening properties")
					}

					for _, v := range flat {
						if strVal, ok := v.(string); ok {
							matchedID, err := regexp.MatchString(q, strVal)
							if err != nil {
								log.WithError(err).Fatalf("failed matching regex %s", q)
							}
							if matchedID {
								fmt.Printf("ID: %s\nProperties: %s\n", id, obj.String())
								break
							}
						}
					}
				}
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

			resourceType, err := fuzzyMatchResourceTypes(inputType, api.ListSupportedResourceTypes())
			if err != nil {
				log.WithError(err).Fatalf("failed matching resources")
			}
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

			resourceType, err := fuzzyMatchResourceTypes(inputType, api.ListSupportedResourceTypes())
			if err != nil {
				log.WithError(err).Fatalf("failed matching resources")
			}
			fmt.Printf("listing matched resource %s\n", resourceType.String())
			resourceList, err := api.ListResources(resourceType)
			if err != nil {
				log.WithError(err).Fatalf("failed listing resource %s", inputType)
			}
			for _, r := range resourceList.Resources {
				id, err := r.GetIdentifier()
				if err != nil {
					panic(fmt.Errorf("failed parsing identifier %s", err))
				}
				fmt.Printf("ID: %s\nProperties: %s\n", id, r.GetRawProperties())
			}
		}
	},
}
