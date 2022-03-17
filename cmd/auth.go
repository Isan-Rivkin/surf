/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

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

	ls "github.com/isan-rivkin/vault-searcher/lib/localstore"
	"github.com/isan-rivkin/vault-searcher/lib/vault"
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
	Short: "Authentication related stuff",
	Long: `
$auth --update-creds 
	`,
	Run: func(cmd *cobra.Command, args []string) {

		vaultAddr := os.Getenv("VAULT_ADDR")
		if err := setAccessCredentialsValues(); err != nil {
			panic(err)
		}
		vault.LdapLogin(*username, *password, vaultAddr)
	},
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
	password = authCmd.PersistentFlags().StringP("password", "p", "", "store password for future auth locally on your OS keyring")
	username = authCmd.PersistentFlags().StringP("username", "u", "", "store username for future auth locally on your OS keyring")
	updateLocalCredentials = authCmd.PersistentFlags().Bool("update-creds", false, "update credentials in locally on your OS keyring")
	method = authCmd.PersistentFlags().StringP("method", "m", "ldap", "authentication method")
}
