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
	accessor "github.com/isan-rivkin/surf/lib/common/jsonutil"
	common "github.com/isan-rivkin/surf/lib/search"
	search "github.com/isan-rivkin/surf/lib/search/ddbsearch"
	"github.com/isan-rivkin/surf/printer"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	tableNamePattern       string
	ddbQuery               string
	ddbOutputType          string
	ddbMatchAll            *bool
	ddbIncludeGlobalTables *bool
	ddbListTables          *bool
	ddbAllowAllTables      *bool
	ddbStopOnFirstMatch    *bool
	ddbFailFast            *bool
	sanitizeOutput         *bool
)

var validDDBOutputs = map[string]bool{
	"pretty": true,
	"json":   true,
}

// ddbCmd represents the ddb command
var ddbCmd = &cobra.Command{
	Use:   "ddb",
	Short: "Search data in DynamoDB (formats: protobuf, base64, json, binary, bytes)",
	Long: `
	
Search free text patterns inside Bytes, Binary, Protobuf, Base64 and Json formats.

=== list existing tables ===

	surf ddb --list-tables
	
=== use -p for aws profile, -r for region ===

	$surf ddb -q val -t table -p my-aws-profile -r us-east-1

=== search all tables with production in their name, where the data containing the pattern val ===

	$surf ddb -q val --all-tables -t production

=== search all tables data containing the word val, output as JSON ===
	
	$surf ddb -q val --all-tables -o json

=== stop on first match, search all tables data containing the word val ===
	
	$surf ddb -q val -t my-prefix-table --stop-first-match

`,
	Run: func(cmd *cobra.Command, args []string) {

		if !*ddbListTables {
			if !*ddbMatchAll && ddbQuery == "" {
				log.Fatalf("invalid query input empty, use --help or --all")
			}

			if *ddbMatchAll && ddbQuery != "" {
				log.Fatalf("invalid query input %s not empty used with --all, use --help", ddbQuery)
			}

			if _, exist := validDDBOutputs[ddbOutputType]; !exist {
				log.Fatalf("invalid output type %s only valid %v, use --help", ddbOutputType, validDDBOutputs)
			}
			if !*ddbAllowAllTables && tableNamePattern == "" {
				log.Fatal("must use --all-tables explicitly or --table <pattern> flag, use --help")
			}

		}
		tui := buildTUI()
		// MARSHAL ATTRIBUTES UTILITY https://docs.aws.amazon.com/sdk-for-go/api/service/dynamodb/dynamodbattribute/
		auth, err := awsu.NewSessionInput(awsProfile, awsRegion)

		if err != nil {
			log.WithError(err).Fatalf("failed creating session in AWS")
		}
		awsRegion = auth.EffectiveRegion

		client, err := awsu.NewDDB(auth)

		if err != nil {
			log.WithError(err).Fatalf("failed creating ddb session")
		}
		ddb := awsu.NewDDBClient(client)
		if *ddbListTables {
			tui.GetLoader().Start("listing dynamodb tables", "", "green")
			if err := listDDBTables(ddb, true, *ddbIncludeGlobalTables, tui); err != nil {
				log.WithError(err).Error("failed listing tables")
			}
			return
		} else {

			if *ddbMatchAll {
				ddbQuery = "\\..*"
			}

			parallel := 30
			m := common.NewDefaultRegexMatcher()
			p := search.NewParserFactory()
			s := search.NewSearcher[awsu.DDBApi, common.Matcher](ddb, m, p)
			i, err := search.NewSearchInput(tableNamePattern, ddbQuery, *ddbFailFast, *ddbIncludeGlobalTables, *ddbStopOnFirstMatch, search.ObjectMatch, parallel)
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

func printDDBSearchOutputAsJSON(input *search.Input, output *search.Output) {
	tablesSearched := map[string]bool{}
	jsonTable := map[string]any{}
	for idx, match := range output.Matches {
		table := map[string]string{}
		tableKey := "Table"
		tableVal := match.TableName
		table[tableKey] = tableVal
		tablesSearched[match.TableName] = true
		for k, v := range match.ObjectData {
			val := aws.StringValue(v)
			if sanitizeOutput != nil && *sanitizeOutput {
				val = printer.SanitizeASCII(val)
			}
			table[k] = val
		}
		jsonTable[fmt.Sprintf("hit_%d", idx)] = table
	}

	summary := map[string]string{
		"Total_Matches":  fmt.Sprintf("%d", len(output.Matches)),
		"Tables_Scanned": fmt.Sprintf("%d", len(tablesSearched)),
		"Query":          input.Value,
	}

	j := map[string]any{
		"Hits":    jsonTable,
		"Summary": summary,
	}

	obj, err := accessor.NewJsonContainerFromInterface("Result", j)
	if err != nil {
		log.WithError(err).Fatalf("failed parsing map to json container")
	}
	fmt.Println(printer.PrettyJson(obj.String()))
}
func printDDBSearchOutput(input *search.Input, output *search.Output, tui printer.TuiController[printer.Loader, printer.Table]) {

	if ddbOutputType == "json" {
		printDDBSearchOutputAsJSON(input, output)
		return
	}
	tablesSearched := map[string]bool{}
	labels := []string{}
	table := map[string]string{}

	for idx, match := range output.Matches {

		tableKey := fmt.Sprintf("#%d Table", idx+1)
		tableVal := printer.ColorHiYellow(match.TableName)

		labels = append(labels, tableKey)
		table[tableKey] = tableVal
		tablesSearched[match.TableName] = true
		for k, v := range match.ObjectData {
			keyLabel := fmt.Sprintf("key.%s", k)
			labels = append(labels, keyLabel)
			val := aws.StringValue(v)
			if sanitizeOutput != nil && *sanitizeOutput {
				val = printer.SanitizeASCII(val)
			}
			table[keyLabel] = val
		}
	}
	if ddbOutputType == "pretty" {
		tui.GetTable().PrintInfoBox(table, labels, true)

		tui.GetTable().PrintInfoBox(
			map[string]string{
				"Total Matches":  fmt.Sprintf("%d", len(output.Matches)),
				"Tables Scanned": fmt.Sprintf("%d", len(tablesSearched)),
				"Query":          input.Value,
			},
			[]string{
				"Total Matches",
				"Tables Scanned",
				"Query",
			}, true)
	}
}

func listDDBTables(ddb awsu.DDBApi, withNonGlobal, withGlobal bool, tui printer.TuiController[printer.Loader, printer.Table]) error {
	columns := []string{"#", "URL"}
	table := map[string]string{
		"#": "Table Name",
	}
	tables, err := ddb.ListCombinedTables(withNonGlobal, withGlobal)
	tui.GetLoader().Stop()
	if err != nil {
		log.WithError(err).Error("failed listing tables")
		return err
	}
	for idx, t := range tables {
		rowNum := fmt.Sprintf("%d", idx+1)
		columns = append(columns, rowNum)
		url := awsu.GenerateDDBWebURL(t.TableName(), awsRegion)
		table[rowNum] = fmt.Sprintf("%s\n%s\n", t.TableName(), printer.ColorFaint(url))
	}

	tui.GetTable().PrintInfoBox(table, columns, true)
	return nil
}

func init() {
	rootCmd.AddCommand(ddbCmd)

	ddbCmd.PersistentFlags().StringVarP(&awsProfile, "profile", "p", getDefaultProfileEnvVar(), "~/.aws/credentials chosen account")
	ddbCmd.PersistentFlags().StringVarP(&awsRegion, "region", "r", "", "~/.aws/config default region if empty")
	ddbCmd.PersistentFlags().StringVarP(&ddbQuery, "query", "q", "", "filter query regex supported (if used with --all will error)")
	ddbCmd.PersistentFlags().StringVarP(&tableNamePattern, "table", "t", "", "regex table pattern name to match")
	ddbCmd.PersistentFlags().StringVarP(&ddbOutputType, "out", "o", "pretty", "output format [json, pretty]")
	ddbFailFast = ddbCmd.Flags().Bool("fail-fast", false, "fail on first error seen")
	ddbListTables = ddbCmd.Flags().Bool("list-tables", false, "list all available tables")
	ddbMatchAll = ddbCmd.Flags().Bool("all", false, "match all data (same as using -q '\\\\..*') if used with --query will error")
	ddbIncludeGlobalTables = ddbCmd.Flags().Bool("include-global-tables", true, "if true will include global tables during search")
	ddbStopOnFirstMatch = ddbCmd.Flags().Bool("stop-first-match", false, "if true stop stop searching on first match found")
	sanitizeOutput = ddbCmd.Flags().Bool("sanitize", true, "if true will remove all non-ascii charts from outputs")
	ddbAllowAllTables = ddbCmd.Flags().Bool("all-tables", false, "when not providing --table pattern this flag required (potentially expensive)")
}
