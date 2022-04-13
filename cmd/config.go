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
	"errors"
	"fmt"
	"os"

	ls "github.com/isan-rivkin/surf/lib/localstore"
	"github.com/manifoldco/promptui"
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

const (
	unameKey  string       = "username"
	pwdKey    string       = "password"
	VaultLdap ls.Namespace = "vault-ldap"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "app related configuration",
	Long:  `Use the config command to configure everything from Auth to Parameters and platforms.`,
	Run: func(cmd *cobra.Command, args []string) {
		runInteractiveConfigLoop()
	},
}

var optsToHandlers = map[string]func() error{
	"Vault: store locally LDAP auth details":             setLocalstoreCredentials,
	"Vault: set default mount path to start search from": getEnvConfigOutput(EnvKeyVaultDefaultMount, "enter default search mount"),
	"List all stored keychain details":                   listAllKeychainDetails,
	"exit": func() error {
		os.Exit(0)
		return nil
	},
}

func runInteractiveConfigLoop() {
	var confOpts []string
	for o := range optsToHandlers {
		confOpts = append(confOpts, o)
	}

	for {

		prompt := promptui.Select{
			Label: "Select an option:",
			Items: confOpts,
		}

		_, result, err := prompt.Run()

		if err != nil {
			log.WithError(err).Fatal("failed parsing prompt info")
		}

		h, exist := optsToHandlers[result]

		if !exist {
			log.Warn("no such options")
		} else {
			err := h()
			if err != nil {
				log.WithError(err).Error("failed executing command")
			}
		}
	}

}

func setVaultAccessCredentialsValues() error {
	if *method != "ldap" {
		return fmt.Errorf("only ldap method supported not %s", *method)
	}

	if *username == "" || *password == "" {

		var name, pwd string
		var err error
		sm := newStoreManager()

		if sm.IsNamespaceSet(VaultLdap) && !*updateLocalCredentials {

			vals, err := sm.GetValues(VaultLdap)

			if err != nil {
				return err
			}
			name = vals[unameKey]
			pwd = vals[pwdKey]

		} else {
			name, pwd, err = getUserInteractiveCredentials()
		}

		if *updateLocalCredentials {
			if err = saveLocalstoreCredentials(sm, name, pwd); err != nil {
				return err
			}
		}

		username = &name
		password = &pwd
		return err
	}
	return nil
}

func getEnvConfigOutput(env, label string) func() error {
	return func() error {
		prompt := promptui.Prompt{
			Label: label,
		}

		result, err := prompt.Run()

		if err != nil {
			return err
		}

		fmt.Printf("to configure run:\n\t export %s=%s_%s", env, EnvVarPrefix, result)
		os.Exit(0)
		return nil
	}
}
func listAllKeychainDetails() error {
	sm := newStoreManager()
	result, err := sm.ListAll()

	if err != nil {
		return err
	}

	for _, res := range result {
		for k, v := range res {
			if log.GetLevel() >= log.DebugLevel {
				fmt.Println(fmt.Sprintf("%s: %s", k, v))
			} else {
				fmt.Println(fmt.Sprintf("%s: val_len = %d", k, len(v)))
			}
		}
	}
	return nil
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
		log.Fatalf("Prompt failed %v\n", err)
		return "", "", err
	}

	prompt = promptui.Prompt{
		Label:    "Username",
		Validate: validate,
	}

	name, err := prompt.Run()

	if err != nil {
		log.Fatalf("Prompt failed %v\n", err)
		return "", "", err
	}

	return name, pwd, err
}

func newStoreManager() ls.StoreManager[ls.Store] {

	s := ls.NewStore(AppName)
	sm := ls.NewStoreManager(s, map[ls.Namespace][]string{
		VaultLdap: {unameKey, pwdKey},
	})
	return sm
}

func setLocalstoreCredentials() error {
	sm := newStoreManager()
	name, pwd, err := getUserInteractiveCredentials()

	if err != nil {
		return nil
	}

	if err := saveLocalstoreCredentials(sm, name, pwd); err != nil {
		return err
	}

	return nil
}

func saveLocalstoreCredentials(sm ls.StoreManager[ls.Store], name, pwd string) error {

	err := sm.SetNSValues(VaultLdap, map[string]string{
		unameKey: name,
		pwdKey:   pwd,
	})

	username = &name
	password = &pwd
	return err
}

func init() {
	rootCmd.AddCommand(configCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// configCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// configCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
