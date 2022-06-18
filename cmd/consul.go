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
	search "github.com/isan-rivkin/surf/lib/search/vaultsearch"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	consulPrefix *string
	consulQuery  *string
)

// consulCmd represents the consul command
var consulCmd = &cobra.Command{
	Use:   "consul",
	Short: "pattern matching against keys in Consul",
	Long: `
	$surf consul -q opsdog -p ops
	$surf consul --query "^web/(production|sandboxes)"
	$surf consul --query --query "AWS_SECRET_ACCESS_KEY"
	`,
	Run: func(cmd *cobra.Command, args []string) {

		client := runConsulDefaultAuth()
		log.WithFields(log.Fields{
			"address":   client.GetConsulAddr(),
			"base_path": *consulQuery,
			"query":     *consulPrefix,
		}).Info("starting search")

		m := search.NewDefaultRegexMatcher()
		s := consul.NewSearcher[consul.ConsulClient, search.Matcher](client, m)
		output, err := s.Search(consul.NewSearchInput(*consulPrefix, *consulQuery))

		if err != nil {
			log.WithError(err).Fatal("error while searching for keys")
		}

		for i, keys := range output.Matches {
			fmt.Printf("%d. %s\n", i, keys)
		}
	},
}

func runConsulDefaultAuth() consul.ConsulClient {
	consulAddr := os.Getenv("CONSUL_HTTP_ADDR")
	client := consul.NewClient(consulAddr)
	return *client
}

func init() {
	rootCmd.AddCommand(consulCmd)
	consulPrefix = consulCmd.PersistentFlags().StringP("query", "q", "", "search query regex supported")
	consulQuery = consulCmd.PersistentFlags().StringP("prefix", "p", "/", "the prefix the search query starts from")

	vaultCmd.MarkPersistentFlagRequired("query")
}
