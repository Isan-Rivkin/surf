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
	"strings"

	"github.com/isan-rivkin/surf/lib/awsu"
	common "github.com/isan-rivkin/surf/lib/search"
	search "github.com/isan-rivkin/surf/lib/search/s3search"
	printer "github.com/isan-rivkin/surf/printer"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	s3MultiAWSProfile *[]string
	bucketName        string
	keyPrefix         string
	s3WebOutput       *bool
	allowAllBuckets   *bool
)

// s3Cmd represents the s3 command
var s3Cmd = &cobra.Command{
	Use:   "s3",
	Short: "Pattern matching against paths and objects in S3",
	Long: `
	Pattern matching against keys in AWS S3 buckets.
	
=== use -p to specify aws profile ===

	$surf s3 -q my-key -b bucket-prefix -p my-aws-profile
	
=== use --prefix to start search from specific key in a bucket ===

	$surf s3 -q my-key --prefix prefix-key -b my-bucket

=== use --aws-session to search in multiple aws profiles/regions ===
	
	$surf s3 -q my-key --aws-session profile1,region1 --aws-session profile2,region2

=== Regex on bucket names to search in ===

	$surf s3  -q '\.json$' -b '^(prod)(.*)-public'

	` + getEnvVarConfig("s3"),
	Run: func(cmd *cobra.Command, args []string) {
		tui := buildTUI()
		sessionInputs, err := resolveAWSSessions(s3MultiAWSProfile, awsProfile, awsRegion)
		if err != nil {
			log.WithError(err).Fatalf("failed building input for AWS session")
		}
		auths, err := awsu.NewSessionInputMatrix(sessionInputs)

		if err != nil {
			log.WithError(err).Fatalf("failed creating session in AWS")
		}

		for _, auth := range auths {

			s3Client, err := awsu.NewS3(auth)

			if err != nil {
				log.WithError(err).Fatalf("failed creating S3 client")
			}

			api := awsu.NewS3Client(s3Client)
			parallel := 30

			bucketName = *getEnvOrOverride(&bucketName, EnvKeyS3DefaultBucket)

			input := search.NewSearchInput(bucketName, keyPrefix, filterQuery, parallel, *allowAllBuckets)
			m := common.NewDefaultRegexMatcher()
			s := search.NewSearcher[awsu.S3API, common.Matcher](api, m)

			tui.GetLoader().Start("searching s3", "", "green")

			output, err := s.Search(input)

			tui.GetLoader().Stop()

			if err != nil {
				msg := "error while searching keys"
				if err.Error() == search.TooManyBucketsErr {
					msg = "too many buckets, use --bucket <pattern> to filter buckets or use --all-buckets to allow anyway (discouraged)"
				}
				log.WithError(err).Fatalf(msg)
			}

			if !*s3WebOutput {
				for bucketName, matchedKeys := range output.BucketToMatches {
					for _, k := range matchedKeys {
						fmt.Printf("s3://%s/%s\n", bucketName, k)
					}
				}
				return
			}
			labelsOrder := []string{"Match", "Bucket", "AWS Session", "Num #"}
			labelsOrderSummary := []string{"Bucket", "Query"}
			tables := []map[string]string{}
			summaryTable := map[string]string{
				"Bucket": "Num #",
				"Query":  filterQuery,
			}
			if keyPrefix != "" {
				summaryTable["Prefix"] = keyPrefix
			}

			for bucketName, matchedKeys := range output.BucketToMatches {
				bucketInfo := map[string]string{}
				matches := fmt.Sprintf("%d", len(matchedKeys))
				bucketInfo["Bucket"] = bucketName
				bucketInfo["Num #"] = matches
				bucketInfo["AWS Session"] = fmt.Sprintf("%s %s", auth.EffectiveProfile, auth.EffectiveRegion)
				summaryTable[bucketName] = matches
				labelsOrderSummary = append(labelsOrderSummary, bucketName)

				if len(matchedKeys) == 0 {
					continue
				}

				for _, k := range matchedKeys {
					url := awsu.GenerateS3WebURL(bucketName, auth.EffectiveRegion, k)
					url = printer.FmtURL(url)
					val := bucketInfo["Match"]
					bucketInfo["Match"] = fmt.Sprintf("%s\n%s", val, url)
				}
				tables = append(tables, bucketInfo)
			}

			for _, t := range tables {
				tui.GetTable().PrintInfoBox(t, labelsOrder, false)
			}

			if getLogLevelFromVerbosity() >= log.DebugLevel {
				tui.GetTable().PrintInfoBox(summaryTable, labelsOrderSummary, false)
			}
		}
	},
}

func resolveAWSSessions(multiple *[]string, profile, region string) ([]*awsu.AWSSessionInput, error) {
	if multiple != nil && len(*multiple) > 0 {
		log.Debugf("using multiple aws sessions, got %v", *multiple)
		return inputToMultipleAWSSessions(*multiple)
	}
	log.Debugf("using since aws sessions profile=%s region=%s", profile, region)
	return []*awsu.AWSSessionInput{
		{
			Profile: profile,
			Region:  region,
		},
	}, nil
}

func inputToMultipleAWSSessions(input []string) ([]*awsu.AWSSessionInput, error) {
	var sessions []*awsu.AWSSessionInput
	for _, pair := range input {
		tuple := strings.Split(pair, ",")
		if len(tuple) != 2 {
			return nil, fmt.Errorf("invalid input, must be in pairs of profile,region; got %v", input)
		}
		profile := tuple[0]
		region := tuple[1]
		sessions = append(sessions, &awsu.AWSSessionInput{
			Profile: profile,
			Region:  region,
		})
	}
	return sessions, nil
}

func init() {
	rootCmd.AddCommand(s3Cmd)
	s3Cmd.PersistentFlags().StringVarP(&awsProfile, "profile", "p", getDefaultProfileEnvVar(), "~/.aws/credentials chosen account")
	s3Cmd.PersistentFlags().StringVarP(&awsRegion, "region", "r", "", "~/.aws/config default region if empty")
	s3Cmd.PersistentFlags().StringVarP(&keyPrefix, "prefix", "k", "", "key prefix to start search from")
	s3Cmd.PersistentFlags().StringVarP(&filterQuery, "query", "q", "", "filter query regex supported")
	s3Cmd.PersistentFlags().StringVarP(&bucketName, "bucket", "b", "", "bucket query to start from search")
	s3MultiAWSProfile = s3Cmd.PersistentFlags().StringArray("aws-session", []string{}, "search in multiple aws profiles & regions (comma separated: --aws-session default,us-east-1 --aws-session dev-account,us-west-2) - overrides --profile and --region")
	s3WebOutput = s3Cmd.PersistentFlags().Bool("output-url", true, "Output the results with clickable URL links")
	allowAllBuckets = s3Cmd.PersistentFlags().Bool("all-buckets", false, "when not providing --bucket pattern this flag required to allow all buckets search")
	s3Cmd.MarkPersistentFlagRequired("query")
}
