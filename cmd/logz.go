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

	es "github.com/isan-rivkin/surf/lib/elastic"
	esSearch "github.com/isan-rivkin/surf/lib/search/essearch"
	"github.com/isan-rivkin/surf/printer"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	listSubAccounts            *bool
	logzToken                  *string
	logzAddr                   *string
	logzQuery                  *string
	logzNotQuery               *string
	logzTimeFmt                *string
	logzWindowTime             *string
	logzWindowOffsetTime       *int
	logzTimeRangefield         *string
	logzAccountIds             *[]string
	logzAccountNames           *[]string
	logzLimitSize              *int
	logzNoFmtOutput            *bool
	logzOutputJson             *bool
	logzTruncateOutput         *bool
	logzUpdateLocalCredentials *bool
)

// logzCmd represents the logz command
var logzCmd = &cobra.Command{
	Use:   "logz",
	Short: "Search in logz.io (elasticsearch)",
	Long: `

Search docs containing the word 'exception' with limit size 200 

	surf logz -q 'exception' -l 200
	
Search docs containing the word 'something' across ALL sub-accounts matching production/automation

	surf logz -q 'something' --acc production --acc automation 

Search docs containing the word 'production', field errorCode with value Access* and are not containing 'dummy' 

	surf logz -q 'production AND errorCode:Access*' --nq 'dummy'

Search docs across 10 day window with 2 days offset (e.g all matches between 12 days ago until 2 days ago)

	surf logz -o 2 -w 10d -q 'some pattern'
	
	` + getEnvVarConfig("logz"),
	Run: func(cmd *cobra.Command, args []string) {
		updateLocalCredentials = logzUpdateLocalCredentials
		globalToken = logzToken

		tui := buildTUI()
		var err error
		// address
		logzAddr = getEnvOrOverride(logzAddr, EnvLogzIOURL)
		if logzAddr == nil || *logzAddr == "" {
			log.Fatal("no elastic address provided please use --address or environment variables docs in --help command")
		}
		isLogz := true
		logzToken = getEnvOrOverride(logzToken, EnvLogzIOToken)

		confBuilder, err := initESConfWithAuth("", "", *logzToken, isLogz)

		if err != nil {
			log.WithError(err).Fatal("failed initiating logz config")
		}

		*logzAccountIds, err = getLogzIOAccountIDs()
		if err != nil {
			log.WithError(err).Fatal("failed getting logz account ids")
		}

		if *listSubAccounts {
			resp, err := listLogzIOAccounts()
			if err != nil {
				log.WithError(err).Fatal("failed listing logz-io sub accounts")
			}
			labelsOrder := []string{"Name"}
			table := map[string]string{
				"Name": "ID",
			}

			for _, a := range resp.Accounts {
				labelsOrder = append(labelsOrder, a.AccountName)
				table[a.AccountName] = fmt.Sprint(a.AccountID)
			}
			tui.GetTable().PrintInfoBox(table, labelsOrder, false)
			return
		}

		log.Debugf("logz url=%s accountIds=%v query=%s", *logzAddr, *logzAccountIds, *logzQuery)

		if err != nil {
			log.WithError(err).Fatal("failed initiating configuration for logz, please check auth details provided")
		}

		if getLogLevelFromVerbosity() >= log.DebugLevel {
			confBuilder.WithColoredLogger()
		}

		confBuilder.WithTransport(es.NewLogzioTransport("/v2/whoami", "/v1/search", *logzAccountIds, *logzWindowOffsetTime, nil))

		conf := confBuilder.WithURL(*logzAddr).Build()
		client, err := es.NewOpenSearchClient(*conf)

		if err != nil {
			log.WithError(err).Fatal("failed creating logz-io client")
		}

		offsetTimeStr := ""

		if *logzWindowOffsetTime > 0 {
			offsetTimeStr = fmt.Sprintf("%dd", *logzWindowOffsetTime)
		}

		//q, err := es.NewQueryBuilder().WithKQL(*logzQuery).Build()
		q, jsonQuery, err := esSearch.NewQueryBuilder().
			WithMustContain(*logzQuery).
			WithMustNotContain(*logzNotQuery).
			WithTimeRangeWindow(*logzWindowTime, offsetTimeStr, *logzTimeRangefield, *logzTimeFmt).
			WithSize(uint64(*logzLimitSize)).
			BuildBoolQuery()

		if err != nil {
			log.WithError(err).Fatalf("failed creating search query %s", *logzQuery)
		}

		log.Debugf("query %s", string(jsonQuery))
		tui.GetLoader().Start("searching logz-io", "", "green")
		res, err := client.Search(es.NewSearchRequest(q, nil, true))
		tui.GetLoader().Stop()
		if err != nil || res == nil {
			log.WithError(err).Fatal("failed searching logzio")
		}
		printLogzOutput(res, tui)
	},
}

func printEsOutput(res *es.SearchResponse, timeRangeField string, outIndex, fmtDoc, isTruncate, jsonOutput bool, tui printer.TuiController[printer.Loader, printer.Table]) {
	if jsonOutput {
		out, err := res.GetResponseAsJson()
		if err != nil {
			log.Fatalf("failed gettings result as json %s", err.Error())
		}
		fmt.Print(out)
		return
	}

	hits, err := res.GetHits()
	if err != nil {
		log.Fatalf("failed gettings hits result %s", err.Error())
	}

	hitLabels := []string{"ID", "Score"}
	if timeRangeField != "" {
		hitLabels = append(hitLabels, "Time")
	}
	if outIndex {
		hitLabels = append(hitLabels, "Index")
	}
	for _, hit := range hits {
		hitTable := map[string]string{}
		if id, ok := hit.GetID(); ok {
			hitTable["ID"] = id
		}
		if score, ok := hit.GetScore(); ok {
			hitTable["Score"] = fmt.Sprintf("%v", score)
		}
		if timeRangeField != "" {
			if ts, ok := hit.GetSourceStrVal(timeRangeField); ok {
				hitTable["Time"] = ts
			}
		}
		if outIndex {
			if idx, ok := hit.GetIndex(); ok {
				hitTable["Index"] = idx
			}
		}
		tui.GetTable().PrintInfoBox(hitTable, hitLabels, true)

		if source, ok := hit.GetSourceAsJson(); ok {
			docOut := source
			if isTruncate && !jsonOutput {
				docOut = printer.TruncateText(docOut, 200, "")
			}
			if !fmtDoc && !isTruncate {
				docOut = printer.PrettyJson(source)
			}
			fmt.Println(docOut)

		}

	}

	// summary table
	total, err := res.GetHitsCount()
	if err != nil {
		log.Debugf("failed getting hits count %s", err.Error())
		total = 0
	}

	maxScoreStr := ""
	maxScore, err := res.GetMaxScore()
	if err == nil {
		maxScoreStr = fmt.Sprintf("%v", maxScore)
	} else {
		log.Debugf("failed getting max score %s", err.Error())
	}

	summaryLabels := []string{"Max Score", "Total Hits #", "Hits (Limit) #"}
	summary := map[string]string{
		"Total Hits #":   fmt.Sprintf("%d", total),
		"Hits (Limit) #": fmt.Sprintf("%d", len(hits)),
		"Max Score":      maxScoreStr,
	}
	tui.GetTable().PrintInfoBox(summary, summaryLabels, false)
}
func printLogzOutput(res *es.SearchResponse, tui printer.TuiController[printer.Loader, printer.Table]) {
	printEsOutput(res, *logzTimeRangefield, false, *logzNoFmtOutput, *logzTruncateOutput, *logzOutputJson, tui)
}

func listLogzIOAccounts() (*es.LogzAccountsListResponse, error) {
	logzToken = getEnvOrOverride(logzToken, EnvElasticsearchToken)
	resp, err := es.NewLogzHttpClient(*logzAddr, *logzToken).ListTimeBasedAccounts()
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func getLogzIOAccountIDs() ([]string, error) {
	if logzAccountNames != nil && len(*logzAccountNames) > 0 {
		var accountIdsResult []string
		res, err := listLogzIOAccounts()
		if err != nil || res == nil {
			return nil, fmt.Errorf("failed listing accounts %s", err.Error())
		}
		for _, desiredName := range *logzAccountNames {
			desired := strings.ToLower(desiredName)
			for _, remoteAcc := range res.Accounts {
				remote := strings.ToLower(remoteAcc.AccountName)
				if strings.Contains(remote, desired) {
					log.Debugf("account %s match %s", remoteAcc.AccountName, desiredName)
					accountIdsResult = append(accountIdsResult, fmt.Sprintf("%d", remoteAcc.AccountID))
				}
			}
		}
		log.Debugf("matches account %v ids %v", *logzAccountNames, accountIdsResult)
		return accountIdsResult, nil
	}
	*logzAccountIds = getEnvStrSliceOrOverride(logzAccountIds, EnvLogzIOSubAccountIDs)
	return *logzAccountIds, nil
}
func init() {
	logzOutputJson = logzCmd.Flags().Bool("json", false, "if set the output will be in JSON format (for script usage)")
	listSubAccounts = logzCmd.Flags().Bool("list-accounts", false, "list all sub account and ids to use for --sub-account or env var to search in")
	logzNoFmtOutput = logzCmd.Flags().Bool("no-fmt", false, "if true the output document will not be formatted, usually when the output is not a json formatted doc we want raw.")
	logzTruncateOutput = logzCmd.Flags().Bool("truncate", false, "if true the output will be truncated.")
	logzAccountNames = logzCmd.PersistentFlags().StringArray("acc", []string{}, "sub-account names contains value instead of ids to search in (must have list account permission) --acc QA --acc 'Audit Logs'")
	logzToken = logzCmd.PersistentFlags().StringP("token", "t", "", "logz.io token must have access to search in sub accounts (optional list accounts)")
	logzAddr = logzCmd.PersistentFlags().String("address", "", "logz.io api endpoint, if not set will use standard SURF_LOGZ_IO_URL env")
	logzQuery = logzCmd.PersistentFlags().StringP("query", "q", "", "kql or free text search query (example: field:value AND free-text)")
	logzAccountIds = logzCmd.PersistentFlags().StringArrayP("sub-account", "g", []string{}, "sub-account ids to search in (must have permission) -g 1234 -g 4567")
	logzNotQuery = logzCmd.PersistentFlags().String("nq", "", "kql or free text search query that must NOT match (bool query)")
	logzTimeFmt = logzCmd.PersistentFlags().String("time-fmt", "strict_date_optional_time", "default logz.io time, see range query time format in elastic docs")
	logzWindowTime = logzCmd.PersistentFlags().StringP("window", "w", "", "last time window from now (2d, 50m, 3h units) default last 2 days")
	logzWindowOffsetTime = logzCmd.PersistentFlags().IntP("days-offset", "o", 0, "days offset of last result window i.e return results max from 2 days ago")
	logzTimeRangefield = logzCmd.PersistentFlags().String("time-key", "@timestamp", "the field to use for time range query's")
	logzLimitSize = logzCmd.PersistentFlags().IntP("limit", "l", 10, "limit size of documents to return")
	logzUpdateLocalCredentials = logzCmd.PersistentFlags().Bool("update-creds", false, "update credentials locally on your OS keyring")
	rootCmd.AddCommand(logzCmd)
}
