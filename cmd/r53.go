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
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/isan-rivkin/surf/lib/awsu"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	r53MultipleProfiles    string
	recusiveSearchMaxDepth *int
	recordInput            *string
	awsProfile             string
	skipNSVerification     bool
	muteR53Logs            bool
)

// r53Cmd represents the r53 command
var r53Cmd = &cobra.Command{
	Use:   `r53 -q <dns-record>`,
	Short: "Query route53 to get your dns record values",
	Long: `
	Query Route53 to get all sorts of information about a dns record.
	
	=== search the target of '*.address.com' ===

	$surf r53 -q '*.address.com'

	=== skip Name Server validation (NS records)  ===

	$surf r53 -q '*.address.com' --ns-skip 

	=== search the target of '*.address.com' in multiple AWS profiles ===

	$surf r53 -q '*.address.com'  --aws-profiles default,prod,dev

	`,
	Run: func(cmd *cobra.Command, args []string) {
		debug := false
		recurse := true
		// TODO(multiple aws account not consolidated)
		var awsProfiles []string
		if r53MultipleProfiles != "" {
			awsProfiles = strings.Split(r53MultipleProfiles, ",")
		} else {
			awsProfiles = append(awsProfiles, awsProfile)
		}
		if len(awsProfiles) < 1 {
			log.Fatal("no aws profiles provided")
		}

		for _, awsProf := range awsProfiles {
			in, err := awsu.NewR53Input(aws.StringValue(recordInput), awsProf, debug, muteR53Logs, skipNSVerification, recurse, *recusiveSearchMaxDepth)

			if err != nil {

				log.WithError(err).Fatal("failed creating r53 input")
			}
			_, err = awsu.SearchRoute53(in)

			if err != nil {
				log.WithError(err).Errorf("failed searching r53 in profile %s", awsProf)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(r53Cmd)
	recordInput = r53Cmd.PersistentFlags().StringP("record", "q", "", "target record to find in R53, wildcard supported")
	r53Cmd.PersistentFlags().StringVarP(&awsProfile, "profile", "p", getDefaultProfileEnvVar(), "~/.aws/credentials chosen account")
	r53Cmd.PersistentFlags().BoolVar(&muteR53Logs, "mute-logs", false, "if flag set then logs from route53-cli sdk will be muted")
	r53Cmd.PersistentFlags().BoolVar(&skipNSVerification, "ns-skip", false, "if set then nameservers will not be verified against the hosted zone result")
	r53Cmd.PersistentFlags().StringVar(&r53MultipleProfiles, "aws-profiles", "", "search in multiple aws profiles (comma separated: --aws-profiles prod,dev,staging) overrides --profile")
	maxDepth := 3
	r53Cmd.PersistentFlags().IntVarP(&maxDepth, "max-depth", "d", maxDepth, "if -R is used then specifies when to stop recursive search depth")
	recusiveSearchMaxDepth = &maxDepth
	r53Cmd.MarkPersistentFlagRequired("record")
}
