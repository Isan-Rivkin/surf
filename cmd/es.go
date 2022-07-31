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

	es "github.com/isan-rivkin/surf/lib/elastic"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	esQuery   *string
	esToken   *string
	esAddr    *string
	esIndexes *[]string
)

// esCmd represents the es command
var esCmd = &cobra.Command{
	Use:   "es",
	Short: "Search in Elasticsearch / Opensearch database",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		esAddr = getEnvOrOverride(esAddr, EnvElasticsearchURL)
		// hirearchy --address flag > SURF_ELASTICSEARCH_URL > ELASTICSEARCH_URL
		if esAddr == nil || *esAddr == "" {
			addr := os.Getenv("ELASTICSEARCH_URL")
			esAddr = &addr
		}
		*esIndexes = getEnvStrSliceOrOverride(esIndexes, EnvElasticsearchIndexes)

		log.Infof("es search url=%s indexes=%v query=%s", *esAddr, *esIndexes, *esQuery)

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

		q, err := es.NewQueryBuilder().WithKQL(*esQuery).Build()
		if err != nil {
			log.WithError(err).Fatalf("failed creating search query %s", *esQuery)
		}
		res, err := client.Search(es.NewSearchRequest(q, *esIndexes, true))
		if err != nil || res == nil {
			log.WithError(err).Error("failed searching elastic")
		}

		fmt.Println(res.RawResponse)
	},
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
	esToken = esCmd.PersistentFlags().StringP("token", "t", "", "auth with token")
	esAddr = esCmd.PersistentFlags().String("address", "", "elastic endpoint, if not set will use standard ELASTICSEARCH_URL / SURF_ELASTICSEARCH_URL env")
	esQuery = esCmd.PersistentFlags().StringP("query", "q", "", "kql or free text search query (example: field:value AND free-text)")
	esIndexes = esCmd.PersistentFlags().StringArrayP("index", "i", []string{}, "list of indexes to search -i 'index-a-*' -i index-b can be set via env SURF_ELASTICSEARCH_INDEXES='a,b'")
	rootCmd.AddCommand(esCmd)
}
