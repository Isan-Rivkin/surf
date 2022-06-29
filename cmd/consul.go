/*
Copyright © 2022 Ali Ramberg lryahli@gmail.com

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
	"os"

	consul "github.com/isan-rivkin/surf/lib/consul"
	common "github.com/isan-rivkin/surf/lib/search"
	search "github.com/isan-rivkin/surf/lib/search/consulsearch"
	tui "github.com/isan-rivkin/surf/printer"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	consulDatacenter *string
	consulPrefix     *string
	consulQuery      *string
	consulWebOutput  *bool
	consulFilterKV   *bool
)

// consulCmd represents the consul command
var consulCmd = &cobra.Command{
	Use:   "consul",
	Short: "pattern matching against keys in Consul",
	Long: `
	- The CONSUL_HTTP_ADDR envrionment variable is required to run this command
	$surf consul -q "user=\w+\.\w+"
	$surf consul -q "AWS_SECRET_ACCESS_KEY"
	$surf consul -q ldap -p ops -d op-us-west-2 --output-url=false
	`,
	Run: func(cmd *cobra.Command, args []string) {
		if !*consulFilterKV {
			log.Fatal("for now only key-value search is supported for consul")
		}

		client := runConsulDefaultAuth()
		consulAddress := client.GetConsulAddr()

		if *consulDatacenter == "" {
			dc, err := client.GetCurrentDatacenter()
			if err != nil {
				log.WithError(err).Fatal("failed getting current data center info from agent")
			}
			*consulDatacenter = dc
		}

		log.WithFields(log.Fields{
			"address":      consulAddress,
			"base_path":    *consulPrefix,
			"query":        *consulQuery,
			"dc":           *consulDatacenter,
			"outputWebURL": *consulWebOutput,
		}).Info("starting search")

		m := common.NewDefaultRegexMatcher()
		s := search.NewSearcher[consul.ConsulClient, common.Matcher](client, m)
		output, err := s.Search(search.NewSearchInput(*consulQuery, *consulPrefix))

		if err != nil {
			log.WithError(err).Fatal("error while searching for keys")
		}

		if *consulWebOutput {
			for _, key := range output.Matches {
				webUrl := consul.GenerateWebURL(consulAddress, *consulDatacenter, key)
				fmt.Println(tui.FmtURL(webUrl))

			}
		} else {
			for i, key := range output.Matches {
				fmt.Printf("%d. %s\n", i, key)
			}
		}
	},
}

func runConsulDefaultAuth() consul.ConsulClient {
	consulAddr := os.Getenv("CONSUL_HTTP_ADDR")
	client := consul.NewClient(consulAddr, *consulDatacenter)
	return *client
}

func init() {
	rootCmd.AddCommand(consulCmd)
	consulDatacenter = consulCmd.PersistentFlags().StringP("datacenter", "d", "", "for cross region specify data center or default will be used")
	consulQuery = consulCmd.PersistentFlags().StringP("query", "q", "", "search query regex supported")
	consulPrefix = consulCmd.PersistentFlags().StringP("prefix", "p", "/", "the prefix the search query starts from")
	consulWebOutput = consulCmd.PersistentFlags().Bool("output-url", true, "Output the results with clickable URL links")

	consulFilterKV = consulCmd.PersistentFlags().Bool("filter-kv", true, "compare query input against the key name in the Consul KV engine")

	consulCmd.MarkPersistentFlagRequired("query")
}
