/*
Copyright © 2023 Isan Rivkin isanrivkin@gmail.com
*/
package cmd

import (
	"fmt"
	"strings"

	"github.com/isan-rivkin/surf/lib/awsu"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	cloudcontrolCmdMultiAWSProfile *[]string
)

var ccSupportedArgs = []string{"types", "list", "get"}

// cloudcontrolCmd represents the cloudcontrol command
var cloudcontrolCmd = &cobra.Command{
	Use:       fmt.Sprintf("aws <%s>", strings.Join(ccSupportedArgs, ",")),
	Short:     "AWS Cloud Control API",
	Long:      ``,
	Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	ValidArgs: ccSupportedArgs,
	// Run: func(cmd *cobra.Command, args []string) {
	// 	sessInputs, err := resolveAWSSessions(cloudcontrolCmdMultiAWSProfile, awsProfile, awsRegion)
	// 	if err != nil {
	// 		log.WithError(err).Fatalf("failed building input for AWS session")
	// 	}
	// 	auths, err := awsu.NewSessionInputMatrix(sessInputs)

	// 	if err != nil {
	// 		log.WithError(err).Fatalf("failed creating session in AWS")
	// 	}

	// 	for _, auth := range auths {
	// 		ccClient, err := awsu.NewCloudControl(auth)

	// 		if err != nil {
	// 			log.WithError(err).Fatalf("failed creating cloudcontrol client")
	// 		}

	// 		api := awsu.NewCloudControlAPI(ccClient)

	// 		action := strings.ToLower(args[0])
	// 		inputType, _ := cmd.Flags().GetString("type")
	// 		if action != "types" && inputType == "" {
	// 			log.Fatalf("must specify a resource type --type")
	// 		}
	// 		switch action {
	// 		case "types":
	// 			for _, r := range api.ListSupportedResourceTypes() {
	// 				fmt.Println(r)
	// 			}
	// 		case "get":
	// 			rID, _ := cmd.Flags().GetString("id")
	// 			if rID == "" {
	// 				log.Fatalf("must specify a resource identifier --id")
	// 			}

	// 			resourceType, err := fuzzyMatchResourceTypes(inputType, api.ListSupportedResourceTypes())
	// 			if err != nil {
	// 				log.WithError(err).Fatalf("failed matching resources")
	// 			}
	// 			result, err := api.GetResource(resourceType, rID)

	// 			if err != nil {
	// 				log.WithError(err).Fatalf("failed getting resource %s", rID)
	// 			}
	// 			fmt.Println(result.GetRawProperties())
	// 		case "list":
	// 			resourceType, err := fuzzyMatchResourceTypes(inputType, api.ListSupportedResourceTypes())
	// 			if err != nil {
	// 				log.WithError(err).Fatalf("failed matching resources")
	// 			}
	// 			fmt.Printf("listing matched resource %s\n", resourceType.String())
	// 			resourceList, err := api.ListResources(resourceType)
	// 			if err != nil {
	// 				log.WithError(err).Fatalf("failed listing resource %s", inputType)
	// 			}
	// 			for _, r := range resourceList.Resources {
	// 				id, err := r.GetIdentifier()
	// 				if err != nil {
	// 					panic(fmt.Errorf("failed parsing identifier %s", err))
	// 				}
	// 				fmt.Printf("ID: %s\nProperties: %s\n", id, r.GetRawProperties())
	// 			}
	// 		default:
	// 			log.Fatalf("unknown action %s", action)
	// 		}
	// 	}
	// },
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

func setupCommonCloudInitFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&awsProfile, "profile", "p", getDefaultProfileEnvVar(), "~/.aws/credentials chosen account")
	cmd.PersistentFlags().StringVarP(&awsRegion, "region", "r", "", "~/.aws/config default region if empty")
	cloudcontrolCmdMultiAWSProfile = cmd.PersistentFlags().StringArray("aws-session", []string{}, "search in multiple aws profiles & regions (comma separated: --aws-session default,us-east-1 --aws-session dev-account,us-west-2) - overrides --profile and --region")
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
