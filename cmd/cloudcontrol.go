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

var ccSupportedArgs = []string{"types", "list", "get", "create", "update", "delete"}

// cloudcontrolCmd represents the cloudcontrol command
var cloudcontrolCmd = &cobra.Command{
	Use:       fmt.Sprintf("aws <%s>", strings.Join(ccSupportedArgs, ",")),
	Short:     "AWS Cloud Control API",
	Long:      ``,
	Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	ValidArgs: ccSupportedArgs,
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
			cfClient, err := awsu.NewCloudFormation(auth)

			if err != nil {
				log.WithError(err).Fatalf("failed creating cloudformation (helper) client")
			}

			api := awsu.NewCloudControlAPIWithDynamicResources(ccClient, cfClient)

			action := strings.ToLower(args[0])
			inputType, _ := cmd.Flags().GetString("type")
			if action != "types" && inputType == "" {
				log.Fatalf("must specify a resource type --type")
			}
			switch action {
			case "types":
				for _, r := range api.ListSupportedResourceTypes() {
					fmt.Println(r)
				}
			case "get":
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
			case "list":
				resourceType, err := fuzzyMatchResourceTypes(inputType, api.ListSupportedResourceTypes())
				if err != nil {
					log.WithError(err).Fatalf("failed matching resources")
				}
				_, err = api.ListResources(resourceType)
				if err != nil {
					log.WithError(err).Fatalf("failed listing resource %s", inputType)
				}
			default:
				log.Fatalf("unknown action %s", action)
			}
		}
	},
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
		return nil, fmt.Errorf("no resource type matched")
	}
	return match, nil
}

func init() {
	rootCmd.AddCommand(cloudcontrolCmd)

	cloudcontrolCmd.PersistentFlags().StringVarP(&awsProfile, "profile", "p", getDefaultProfileEnvVar(), "~/.aws/credentials chosen account")
	cloudcontrolCmd.PersistentFlags().StringVarP(&awsRegion, "region", "r", "", "~/.aws/config default region if empty")
	cloudcontrolCmdMultiAWSProfile = cloudcontrolCmd.PersistentFlags().StringArray("aws-session", []string{}, "search in multiple aws profiles & regions (comma separated: --aws-session default,us-east-1 --aws-session dev-account,us-west-2) - overrides --profile and --region")
	cloudcontrolCmd.Flags().Bool("list-supported", false, "list all supported resource types")
	cloudcontrolCmd.Flags().String("type", "AWS::EC2::Instance", "list resource instances of given type (usage: --type AWS::DynamoDB::Table --type AWS::EC2::VPC)")
	cloudcontrolCmd.Flags().String("id", "", "describe resource via identifier (usage: --id <identifier>)")
}
