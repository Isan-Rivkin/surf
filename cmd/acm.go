/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/isan-rivkin/surf/lib/awsu"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	awsRegion string
)

// acmCmd represents the acm command
var acmCmd = &cobra.Command{
	Use:   "acm",
	Short: "search in ACM",
	Long: `Options to search:
	- Domain Based
	- Attached Resources 
	- Certificate ID 
`,
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Println("acm called")
		auth, err := awsu.NewSessionInput(awsProfile, awsRegion)

		if err != nil {
			log.Panicf("failed creating session in AWS %s", err.Error())
		}

		acmClient, err := awsu.NewACM(auth)

		if err != nil {
			log.Panicf("failed creating ACM client %s", err.Error())
		}

		api := awsu.NewAcmClient(acmClient)

		result, err := api.ListAndFilter(1, false, func(c *acm.CertificateDetail) bool {

			return true
		})

		for _, c := range result.Certificates {
			fmt.Println(aws.StringValue(c.CertificateArn))
		}
	},
}

func init() {
	rootCmd.AddCommand(acmCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// acmCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// acmCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	acmCmd.PersistentFlags().StringVarP(&awsProfile, "profile", "p", "default", "~/.aws/credentials chosen account")
	acmCmd.PersistentFlags().StringVarP(&awsRegion, "region", "r", "", "~/.aws/config default region if empty")
}
