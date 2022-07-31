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
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	listSubAccounts  *bool
	logzToken        *string
	logzAddr         *string
	logzQuery        *string
	logzAccountIds   *[]string
	logzAccountNames *[]string
)

// logzCmd represents the logz command
var logzCmd = &cobra.Command{
	Use:   "logz",
	Short: "Search in logz.io (elasticsearch)",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// account ids
		//*logzAccountIds = getEnvStrSliceOrOverride(logzAccountIds, EnvLogzIOSubAccountIDs)
		var err error

		// address
		logzAddr = getEnvOrOverride(logzAddr, EnvLogzIOURL)
		if logzAddr == nil || *logzAddr == "" {
			log.Fatal("no elastic address provided please use --address or environment variables docs in --help command")
		}
		isLogz := true
		logzToken = getEnvOrOverride(logzToken, EnvLogzIOToken)

		*logzAccountIds, err = getLogzIOAccountIDs()
		if err != nil {
			log.WithError(err).Fatal("failed getting logz account ids")
		}

		if *listSubAccounts {
			resp, err := listLogzIOAccounts()
			if err != nil {
				log.WithError(err).Fatal("failed listing logz-io sub accouns %s", err.Error())
			}
			for _, a := range resp.Accounts {
				fmt.Printf("%s: %d\n", a.AccountName, a.AccountID)
			}
			return
		}
		log.Infof("logz url=%s accountIds=%v query=%s", *logzAddr, *logzAccountIds, *logzQuery)
		confBuilder, err := initESConfWithAuth("", "", *logzToken, isLogz)

		if err != nil {
			log.WithError(err).Fatal("failed initiating configuration for logz, please check auth details provided")
		}

		if getLogLevelFromVerbosity() >= log.DebugLevel {
			confBuilder.WithColoredLogger()
		}

		confBuilder.WithTransport(es.NewLogzioTransport("/v2/whoami", "/v1/search", *logzAccountIds, nil))

		conf := confBuilder.WithURL(*logzAddr).Build()
		client, err := es.NewOpenSearchClient(*conf)
		if err != nil {
			log.WithError(err).Fatal("failed creating logzio client")
		}
		q, err := es.NewQueryBuilder().WithKQL(*logzQuery).Build()
		if err != nil {
			log.WithError(err).Fatalf("failed creating search query %s", *logzQuery)
		}
		res, err := client.Search(es.NewSearchRequest(q, nil, true))
		if err != nil || res == nil {
			log.WithError(err).Error("failed searching logzio")
		}

		fmt.Println(res.RawResponse)
	},
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
		log.Debugf("matches account ids %v", *logzAccountNames, accountIdsResult)
		return accountIdsResult, nil
	}
	*logzAccountIds = getEnvStrSliceOrOverride(logzAccountIds, EnvLogzIOSubAccountIDs)
	return *logzAccountIds, nil
}
func init() {
	listSubAccounts = logzCmd.Flags().Bool("list-accounts", false, "list all sub account and ids to use for --sub-account or env var to search in")
	logzAccountNames = logzCmd.PersistentFlags().StringArray("acc", []string{}, "sub-account names instead of ids to search in (must have list account permission) --acc QA --acc 'Audit Logs'")
	logzToken = logzCmd.PersistentFlags().StringP("token", "t", "", "logz.io token must have access to search in sub accounts (optional list accounts)")
	logzAddr = logzCmd.PersistentFlags().String("address", "", "logz.io api endpoint, if not set will use standard SURF_LOGZ_IO_URL env")
	logzQuery = logzCmd.PersistentFlags().StringP("query", "q", "", "kql or free text search query (example: field:value AND free-text)")
	logzAccountIds = logzCmd.PersistentFlags().StringArrayP("sub-account", "g", []string{}, "sub-account ids to search in (must have permission) -a 1234 -a 4567")

	rootCmd.AddCommand(logzCmd)
}
