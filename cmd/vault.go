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
	"os"
	"path/filepath"

	search "github.com/isan-rivkin/surf/lib/search"
	vaultSearch "github.com/isan-rivkin/surf/lib/search/vaultsearch"
	"github.com/isan-rivkin/surf/lib/vault"
	printer "github.com/isan-rivkin/surf/printer"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	query        *string
	parallel     *int
	mount        *string
	prefix       *string
	outputWebURL *bool
)

// vaultCmd represents the vault command
var vaultCmd = &cobra.Command{
	Use:   "vault",
	Short: "pattern matching against keys in Vault",
	Long: `
	$surf vault -q aws -m backend-secrets/prod  -t 15
	$surf vault -q aws -m 'user_.*pro' 
	`,
	Run: func(cmd *cobra.Command, args []string) {
		tui := buildTUI()
		mount := getEnvOrOverride(mount, EnvKeyVaultDefaultMount)
		prefix := getEnvOrOverride(prefix, EnvKeyVaultDefaultPrefix)

		basePath := filepath.Join(*mount, *prefix)

		client := runVaultDefaultAuth()

		log.WithFields(log.Fields{
			"address":   client.GetVaultAddr(),
			"base_path": basePath,
			"query":     *query,
		}).Info("starting search")

		m := search.NewDefaultRegexMatcher()
		s := vaultSearch.NewRecursiveSearcher[vaultSearch.VC, search.Matcher](client, m)

		tui.GetLoader().Start("searching vault", "", "green")

		output, err := s.Search(vaultSearch.NewSearchInput(*query, basePath, *parallel))

		tui.GetLoader().Stop()

		if err != nil {
			log.Fatalf("failed searching vault %s", err.Error())
		}

		if output != nil {
			for _, i := range output.Matches {
				path := i.GetFullPath()
				if *outputWebURL {
					fmt.Println(printer.FmtURL(vault.PathToWebURL(client.GetVaultAddr(), path)))
				} else {
					fmt.Println(path)
				}
			}
		}

	},
}

func runVaultDefaultAuth() vault.Client[vault.Authenticator] {
	vaultAddr := os.Getenv("VAULT_ADDR")
	if err := setVaultAccessCredentialsValues(); err != nil {
		log.WithError(err).Fatal("failed auth to Vault")
	}
	auth := vault.NewLdapAuth(*username, *password, vaultAddr)

	client := vault.NewClient(auth)
	return client
}

func init() {
	rootCmd.AddCommand(vaultCmd)
	query = vaultCmd.PersistentFlags().StringP("query", "q", "", "search query regex supported")
	mount = vaultCmd.PersistentFlags().StringP("mount", "m", "", "mount to start the search at the root")
	prefix = vaultCmd.PersistentFlags().StringP("prefix", "p", "", "$mount/prefix inside the mount to search in")
	parallel = vaultCmd.PersistentFlags().IntP("threads", "t", 10, "parallel search number")

	outputWebURL = vaultCmd.PersistentFlags().Bool("output-url", true, "default output is web urls to click on and go to the browser UI")

	vaultCmd.MarkPersistentFlagRequired("query")
}
