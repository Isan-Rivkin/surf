/*
Copyright © 2022 Isan Rivkin isanrivkin@gmail.com

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
	"github.com/isan-rivkin/surf/lib/awsu"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	recordInput *string
	awsProfile  string
)

// r53Cmd represents the r53 command
var r53Cmd = &cobra.Command{
	Use:   "r53 -q '*.some.dns.record.com'\n",
	Short: "Query route53 to get your dns record values",
	Long: `Query Route53 to get all sorts of information about a dns record.
	r53 will use your default AWS credentials`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("r53 called")

		debug := true
		mute := true
		skipNSVerification := false
		recurse := true
		recurseDepth := 3
		in, err := awsu.NewR53Input(aws.StringValue(recordInput), awsProfile, debug, mute, skipNSVerification, recurse, recurseDepth)

		if err != nil {

			log.WithError(err).Fatal("failed creating r53 input")
		}
		_, err = awsu.SearchRoute53(in)

		if err != nil {
			fmt.Println("nu!!! ", err.Error())
			log.WithError(err).Fatal("failed searching r53")
		}

	},
}

func init() {
	rootCmd.AddCommand(r53Cmd)
	recordInput = r53Cmd.PersistentFlags().StringP("record", "q", "", "target record to find in R53, wildcard supported")
	r53Cmd.PersistentFlags().StringVarP(&awsProfile, "profile", "p", "default", "~/.aws/credentials chosen account")
	r53Cmd.MarkPersistentFlagRequired("record")
}
