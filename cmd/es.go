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
	"errors"
	"fmt"
	"os"
	"sort"

	es "github.com/isan-rivkin/surf/lib/elastic"
	esSearch "github.com/isan-rivkin/surf/lib/search/essearch"
	printer "github.com/isan-rivkin/surf/printer"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	esQuery           *string
	esNotContainQuery *string
	esToken           *string
	esAddr            *string
	esIndexes         *[]string
	esLimitSize       *int
	esNoFmtOutput     *bool
	esTruncateFmt     *bool
	esListIndexes     *bool
)

// esCmd represents the es command
var esCmd = &cobra.Command{
	Use:   "es",
	Short: "Search in Elasticsearch / Opensearch database",
	Long: `

Search docs containing the term 'api' return limit 40 results.

	surf es -q 'api' -l 40

Search docs containing the term 'api' with client field and 'xyz*' pattern and NOT containing the term 'staging'
	
	surf es -q 'api AND client:xyz*' --nq staging
	`,
	Run: func(cmd *cobra.Command, args []string) {
		tui := buildTUI()
		esAddr = getEnvOrOverride(esAddr, EnvElasticsearchURL)
		// hirearchy --address flag > SURF_ELASTICSEARCH_URL > ELASTICSEARCH_URL
		if esAddr == nil || *esAddr == "" {
			addr := os.Getenv("ELASTICSEARCH_URL")
			esAddr = &addr
		}
		*esIndexes = getEnvStrSliceOrOverride(esIndexes, EnvElasticsearchIndexes)

		log.Debugf("es search url=%s indexes=%v query=%s", *esAddr, *esIndexes, *esQuery)

		if esAddr == nil || *esAddr == "" {
			log.Fatal("no elastic address provided please use --address or environment variables docs in --help command")
		}

		isLogz := false
		confBuilder, err := initESConfWithAuth(*username, *password, *esToken, isLogz)

		if err != nil {
			log.WithError(err).Fatal("failed initiating configuration for elastic, please check auth details provided")
		}

		if getLogLevelFromVerbosity() >= log.DebugLevel {
			confBuilder.WithColoredLogger()
		}

		conf := confBuilder.WithURL(*esAddr).Build()
		client, err := es.NewOpenSearchClient(*conf)
		if err != nil {
			log.WithError(err).Fatal("failed creating elastic  client")
		}

		if *esListIndexes {
			displayESListIndices(client, tui)
			return
		}
		//q, err := es.NewQueryBuilder().WithKQL(*esQuery).Build()
		q, jsonQuery, err := esSearch.NewQueryBuilder().
			WithMustContain(*esQuery).
			WithMustNotContain(*esNotContainQuery).
			WithSize(uint64(*esLimitSize)).
			BuildBoolQuery()

		if err != nil {
			log.WithError(err).Fatalf("failed creating search query %s", *logzQuery)
		}

		log.Debugf("query %s", string(jsonQuery))

		if err != nil {
			log.WithError(err).Fatalf("failed creating search query %s", *esQuery)
		}
		tui.GetLoader().Start("searching elasticsearch", "", "green")
		res, err := client.Search(es.NewSearchRequest(q, *esIndexes, true))
		tui.GetLoader().Stop()
		if err != nil || res == nil {
			log.WithError(err).Error("failed searching elastic")
		}

		printEsOutput(res, "", true, *esNoFmtOutput, *esTruncateFmt, tui)
	},
}

func displayESListIndices(client es.ESClient, tui printer.TuiController[printer.Loader, printer.Table]) {
	indicesRes, err := client.ListIndexes()
	if err != nil {
		log.WithError(err).Fatal("failed listing indexes")
	}
	indices, exist := indicesRes.Indices()
	if !exist {
		fmt.Printf("No Indices found \n")
		return
	}
	table := map[string]string{}
	var displayIndices []string
	labels := []string{"Results"}
	for _, idx := range indices {
		if !idx.IsDotIndex() {
			displayIndices = append(displayIndices, idx.GetName())
		}
	}
	sort.Strings(displayIndices)
	for i, idx := range displayIndices {
		num := fmt.Sprintf("# %d", i+1)
		table[num] = idx
		labels = append(labels, num)
	}

	table["Results"] = fmt.Sprintf("User Defined %d, Internal Indexes %d", len(displayIndices), len(indices)-len(displayIndices))
	labels = append(labels, "Name")
	tui.GetTable().PrintInfoBox(table, labels, false)
}

func initESConfWithAuth(uname, pwd, token string, isLogz bool) (*es.ConfigBuilder, error) {

	confBuilder := es.NewConf()
	// if token provided will be used
	token = *getEnvOrOverride(&token, EnvElasticsearchToken)
	if !isLogz && token != "" {
		return confBuilder.WithBasicAuthToken(token), nil
	}

	if isLogz && token != "" {
		return confBuilder.WithHeader(es.LogzIOTokenHeader, token).WithHeader("Content-Type", "application/json"), nil
	}
	// if username / password provided
	uname = *getEnvOrOverride(&uname, EnvElasticsearchUsername)
	pwd = *getEnvOrOverride(&pwd, EnvElasticsearchPwd)

	if uname != "" && pwd != "" {
		return confBuilder.WithUserAuth(uname, pwd), nil
	}
	return nil, errors.New("no valid auth credentials provided")
}

func init() {
	esListIndexes = esCmd.Flags().Bool("list-indexes", false, "list all available indexes --index or env var to search in")
	esToken = esCmd.PersistentFlags().StringP("token", "t", "", "auth with token")
	esNoFmtOutput = esCmd.Flags().Bool("no-fmt", false, "if true the output document will not be formatted, usually when the output is not a json formatted doc we want raw.")
	esTruncateFmt = esCmd.Flags().Bool("truncate", false, "if true the output will be truncated.")
	esAddr = esCmd.PersistentFlags().String("address", "", "elastic endpoint, if not set will use standard ELASTICSEARCH_URL / SURF_ELASTICSEARCH_URL env")
	esQuery = esCmd.PersistentFlags().StringP("query", "q", "", "kql or free text search query (example: field:value AND free-text)")
	esNotContainQuery = esCmd.PersistentFlags().String("nq", "", "kql or free text search query that must NOT match (bool query)")
	esIndexes = esCmd.PersistentFlags().StringArrayP("index", "i", []string{}, "list of indexes to search -i 'index-a-*' -i index-b can be set via env SURF_ELASTICSEARCH_INDEXES='a,b'")
	esLimitSize = esCmd.PersistentFlags().IntP("limit", "l", 10, "limit size of documents to return")
	rootCmd.AddCommand(esCmd)
}
