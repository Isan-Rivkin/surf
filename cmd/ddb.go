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
	common "github.com/isan-rivkin/surf/lib/search"
	search "github.com/isan-rivkin/surf/lib/search/ddbsearch"
	"github.com/isan-rivkin/surf/printer"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	tableNamePattern       string
	ddbQuery               string
	ddbIncludeGlobalTables *bool
	ddbListTables          *bool
	ddbFilterTables        *bool
	ddbFilterKeys          *bool
	ddbFilterData          *bool
	ddbFilterAllOpts       *bool
	ddbStopOnFirstMatch    *bool
	ddbFailFast            *bool
)

// ddbCmd represents the ddb command
var ddbCmd = &cobra.Command{
	Use:   "ddb",
	Short: "DynamoDB search tool",
	Long: `
	TBD...

	surf ddb -q val -t table-* [--filter-tables, --filter-keys, --filter-data, --filter-all]

	surf ddb --list-tables
`,
	Run: func(cmd *cobra.Command, args []string) {
		tui := buildTUI()
		// MARSHAL ATTRIBUTES UTILITY https://docs.aws.amazon.com/sdk-for-go/api/service/dynamodb/dynamodbattribute/
		auth, err := awsu.NewSessionInput(awsProfile, awsRegion)

		if err != nil {
			log.WithError(err).Fatalf("failed creating session in AWS")
		}

		client, err := awsu.NewDDB(auth)

		if err != nil {
			log.WithError(err).Fatalf("failed creating ddb session")
		}
		ddb := awsu.NewDDBClient(client)
		if *ddbListTables {
			listDDBTables(ddb, true, *ddbIncludeGlobalTables)
			return
		} else {

			parallel := 30
			m := common.NewDefaultRegexMatcher()
			p := search.NewParserFactory()
			s := search.NewSearcher[awsu.DDBApi, common.Matcher](ddb, m, p)
			i, err := search.NewSearchInput(tableNamePattern, ddbQuery, *ddbFailFast, *ddbIncludeGlobalTables, search.ObjectMatch, parallel)
			if err != nil {
				log.WithError(err).Error("failed creating search input")
			}
			tui.GetLoader().Start("searching dynamodb", "", "green")
			output, err := s.Search(i)
			tui.GetLoader().Stop()

			if err != nil {
				log.WithError(err).Fatalf("failed running search on dynamodb")
			}
			printDDBSearchOutput(i, output, tui)
		}
	},
}

func printDDBSearchOutput(input *search.Input, output *search.Output, tui printer.TuiController[printer.Loader, printer.Table]) {

	for idx, match := range output.Matches {
		labels := []string{fmt.Sprintf("#%d Table", idx+1)}
		table := map[string]string{
			fmt.Sprintf("#%d Table", idx+1): match.TableName,
		}
		for k, v := range match.ObjectData {
			keyLabel := fmt.Sprintf("key.%s", k)
			labels = append(labels, keyLabel)
			table[keyLabel] = aws.StringValue(v)
		}
		tui.GetTable().PrintInfoBox(table, labels, true)
	}
	tui.GetTable().PrintInfoBox(
		map[string]string{
			"Total Matches": fmt.Sprintf("%d", len(output.Matches)),
			"Query":         input.Value,
		},
		[]string{
			"Total Matches",
			"Query",
		}, true)
}

func listDDBTables(ddb awsu.DDBApi, withNonGlobal, withGlobal bool) {

	tables, err := ddb.ListCombinedTables(withNonGlobal, withGlobal)
	if err != nil {
		log.WithError(err).Fatalf("failed listing tables")
	}
	// TODO pretty print
	for _, t := range tables {
		fmt.Printf("> %s \n", t.TableName())
	}
}

func init() {
	rootCmd.AddCommand(ddbCmd)

	ddbCmd.PersistentFlags().StringVarP(&awsProfile, "profile", "p", getDefaultProfileEnvVar(), "~/.aws/credentials chosen account")
	ddbCmd.PersistentFlags().StringVarP(&awsRegion, "region", "r", "", "~/.aws/config default region if empty")
	ddbCmd.PersistentFlags().StringVarP(&ddbQuery, "query", "q", "", "filter query regex supported")
	ddbCmd.PersistentFlags().StringVarP(&tableNamePattern, "table", "t", "", "regex table pattern name to match")

	ddbFailFast = ddbCmd.Flags().Bool("fail-fast", false, "fail on first error seen")
	ddbListTables = ddbCmd.Flags().Bool("list-tables", false, "list all available tables")
	ddbIncludeGlobalTables = ddbCmd.Flags().Bool("include-global-tables", true, "if true will include global tables during search")
	ddbStopOnFirstMatch = ddbCmd.Flags().Bool("stop-first-match", false, "if true stop stop searching on first match found")
}
