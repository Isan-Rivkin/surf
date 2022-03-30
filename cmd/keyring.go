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

	ls "github.com/isan-rivkin/vault-searcher/lib/localstore"
	"github.com/manifoldco/promptui"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// keyringCmd represents the keyring command
var keyringCmd = &cobra.Command{
	Use:   "keyring",
	Short: "Store secrets in OS Keyring",
	Long:  `Used for easy authentication, for example LDAP credentials, depending on Auth method.`,
	Run: func(cmd *cobra.Command, args []string) {

		s := ls.NewStore()

		sm := ls.NewStoreManager(s)
		username, password, err := getUserInteractiveCredentials()

		if err != nil {
			log.WithError(err).Fatal("wrong input")
		}

		if err := sm.UpdateStore(username, password); err != nil {
			log.WithError(err).Fatal("failed storing credentials in keystore")
			return
		}

		fmt.Println("Successfuly updated key store!")
	},
}

func getUserInteractiveCredentials() (string, string, error) {
	validate := func(input string) error {
		if input == "" {
			return errors.New("no empty input allowed")
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:    "Password",
		Validate: validate,
		Mask:     '*',
	}

	pwd, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return "", "", err
	}

	prompt = promptui.Prompt{
		Label:    "Username",
		Validate: validate,
	}

	name, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return "", "", err
	}

	return name, pwd, err
}

func init() {
	authCmd.AddCommand(keyringCmd)

}
