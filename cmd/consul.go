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
	printer "github.com/isan-rivkin/surf/printer"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	consulDatacenter *string
	consulPrefix     *string
	consulQuery      *string
	consulAddr       *string
	consulWebOutput  *bool
	consulFilterKV   *bool
)

// consulCmd represents the consul command
var consulCmd = &cobra.Command{
	Use:   "consul",
	Short: "pattern matching against keys in Hasicorp Consul",
	Long: `
Pattern matching against keys in Hasicorp Consul

	- Consul address taken from  CONSUL_HTTP_ADDR or via --address

	$surf consul -q "user=\w+\.\w+"
	$surf consul -q "AWS_SECRET_ACCESS_KEY"
	$surf consul -q ldap -p ops -d op-us-west-2 --output-url=false
	`,
	Run: func(cmd *cobra.Command, args []string) {
		if !*consulFilterKV {
			log.Fatal("for now only key-value search is supported for consul")
		}
		tui := buildTUI()

		client := runConsulDefaultAuth()
		consulUiBaseAddr, uiAddrErr := client.GetConsulUIBaseAddr()
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

		tui.GetLoader().Start("searching consul", "", "green")

		//pairs, err := client.List(*consulPrefix)

		// if err != nil {
		// 	log.WithError(err).Fatalf("failed listing all keys under the prefix %s", *consulPrefix)
		// }

		input := search.NewSearchInput(*consulQuery, *consulPrefix)

		m := common.NewDefaultRegexMatcher()
		s := search.NewSearcher[consul.ConsulClient, common.Matcher](client, m)
		output, err := s.Search(input)

		tui.GetLoader().Stop()

		if err != nil {
			log.WithError(err).Fatal("error while searching for keys")
		}

		if *consulWebOutput && uiAddrErr == nil {
			for _, key := range output.Matches {
				webUrl := consul.GenerateKVWebURL(consulUiBaseAddr, key)
				fmt.Println(printer.FmtURL(webUrl))
			}

			labelsOrder := []string{"Total", "Address", "Datacenter"}
			summary := map[string]string{
				"Matches #":  fmt.Sprintf("%d", len(output.Matches)),
				"Address":    consulAddress,
				"Datacenter": *consulDatacenter,
			}

			if *consulPrefix != "" {
				summary["Prefix"] = *consulPrefix
				labelsOrder = append(labelsOrder, "Prefix")
			}
			if *consulQuery != "" {
				summary["Query"] = *consulQuery
				labelsOrder = append(labelsOrder, "Query")
			}

			tui.GetTable().PrintInfoBox(summary, labelsOrder)
		} else {
			for i, key := range output.Matches {
				fmt.Printf("%d. %s\n", i, key)
			}
			if uiAddrErr != nil {
				log.WithError(uiAddrErr).Error("Not Displaying Link to UI, failed building address UI")
			}
		}
	},
}

func runConsulDefaultAuth() consul.ConsulClient {
	if *consulAddr == "" {
		*consulAddr = os.Getenv("CONSUL_HTTP_ADDR")
	}

	client := consul.NewClient(*consulAddr, *consulDatacenter)
	return *client
}

func init() {
	rootCmd.AddCommand(consulCmd)
	consulAddr = consulCmd.PersistentFlags().String("address", "", "consul address to use, default is CONSUL_HTTP_ADDR")
	consulDatacenter = consulCmd.PersistentFlags().StringP("datacenter", "d", "", "for cross region specify data center or default will be used")
	consulQuery = consulCmd.PersistentFlags().StringP("query", "q", "", "search query regex supported")
	consulPrefix = consulCmd.PersistentFlags().StringP("prefix", "p", "/", "the prefix the search query starts from")
	consulWebOutput = consulCmd.PersistentFlags().Bool("output-url", true, "Output the results with clickable URL links")

	consulFilterKV = consulCmd.PersistentFlags().Bool("filter-kv", true, "compare query input against the key name in the Consul KV engine")

	consulCmd.MarkPersistentFlagRequired("query")
}
