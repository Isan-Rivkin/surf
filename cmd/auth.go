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

	ls "github.com/isan-rivkin/search-unified-recusive-fast/lib/localstore"
	search "github.com/isan-rivkin/search-unified-recusive-fast/lib/search/vaultsearch"
	"github.com/isan-rivkin/search-unified-recusive-fast/lib/vault"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	username               *string
	password               *string
	method                 *string
	updateLocalCredentials *bool
	vaultAddr              *string
)

// authCmd represents the auth command
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication related config",
	Long: `
$auth --update-creds 
	`,
	Run: func(cmd *cobra.Command, args []string) {
		search.TEST()

		client := runDefaultAuth()

		nodes, err := client.ListTreeFiltered("some-mount")

		if err != nil {
			panic(err)
		}

		for _, n := range nodes {
			fmt.Println("-> " + n.T + " : " + vault.NodeType(n.KeyValue))
		}
	},
}

func runDefaultAuth() vault.Client[vault.Authenticator] {
	vaultAddr := os.Getenv("VAULT_ADDR")
	if err := setAccessCredentialsValues(); err != nil {
		log.WithError(err).Fatal("failed auth to Vault")
	}
	auth := vault.NewLdapAuth(*username, *password, vaultAddr)

	client := vault.NewClient(auth)
	return client
}

func setAccessCredentialsValues() error {
	if *method != "ldap" {
		return fmt.Errorf("only ldap method supported not %s", *method)
	}

	if *username == "" || *password == "" {

		var name, pwd string
		var err error

		s := ls.NewStore()
		sm := ls.NewStoreManager(s)

		if sm.IsExistAlready() && !*updateLocalCredentials {

			name, pwd, err = sm.GetUserAndPwd()
		} else {
			name, pwd, err = getUserInteractiveCredentials()
		}

		if *updateLocalCredentials {
			err = sm.UpdateStore(name, pwd)
		}

		username = &name
		password = &pwd
		return err
	}
	return nil
}
func init() {
	rootCmd.AddCommand(authCmd)
}
