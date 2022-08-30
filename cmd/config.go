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
	"sort"

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
	tokenKey        string       = "token"
	unameKey        string       = "username"
	pwdKey          string       = "password"
	VaultLdap       ls.Namespace = "vault-ldap"
	ElasticSearchNS ls.Namespace = "elastic-auth"
	LogzSearchNS    ls.Namespace = "logz-auth"
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
	"Vault: store locally LDAP auth details":             func() error { return setLocalstoreCredentials(VaultLdap) },
	"Vault: set default mount path to start search from": getEnvConfigOutput(EnvKeyVaultDefaultMount, "enter default search mount"),
	"ElasticSearch: store locally user/password details": func() error { return setLocalstoreCredentials(ElasticSearchNS) },
	"ElasticSearch: store locally token":                 func() error { return setLocalstoreToken(ElasticSearchNS) },
	"ElasticSearch: clear storage":                       func() error { return clearNamespace(ElasticSearchNS) },
	"Logz.io: store locally token":                       func() error { return setLocalstoreToken(LogzSearchNS) },
	"Logz.io: clear storage":                             func() error { return clearNamespace(LogzSearchNS) },
	"List all stored keychain details":                   listAllKeychainDetails,
	"Opt-Out from latest version check at github.com":    getEnvConfigOutput(EnvVersionCheckOptout, "type 'false' to opt-out"),
	"S3: set default bucket name to start search from":   getEnvConfigOutput(EnvKeyS3DefaultBucket, "enter default bucket name (regex pattern)"),
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

	sort.Strings(confOpts)

	for {

		prompt := promptui.Select{
			Label: "Select an option",
			Items: confOpts,
			Size:  len(confOpts)%30 + 1,
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

func checkIsTokenOrUserAuthStored(ns ls.Namespace) (bool, bool, error) {
	var isToken bool
	var isUserPass bool
	var storageErr error

	sm := newStoreManager()

	if sm.IsNamespaceSet(ns) {
		vals, storageErr := sm.GetValues(ns)

		if storageErr != nil {
			return isToken, isUserPass, storageErr
		}

		tokenVal, hasToken := vals[tokenKey]
		unameVal, hasUser := vals[unameKey]
		pwdVal, hasPwd := vals[pwdKey]

		isToken = hasToken && tokenVal != ""
		isUserPass = hasUser && unameVal != "" && hasPwd && pwdVal != ""
	}
	return isToken, isUserPass, storageErr
}

func clearNamespace(ns ls.Namespace) error {
	sm := newStoreManager()
	return sm.DeleteNamespace(ns)
}

func getAccessTokenValue(ns ls.Namespace, tokenVal string, supportedMethods map[string]bool) (string, error) {
	if _, ok := supportedMethods[*method]; !ok {
		return "", fmt.Errorf("only %v methods are supported. not %s", supportedMethods, *method)
	}

	if tokenVal == "" {

		var token string
		var err error
		sm := newStoreManager()

		if sm.IsNamespaceSet(ns) && !*updateLocalCredentials {

			vals, err := sm.GetValues(ns)

			if err != nil {
				return "", err
			}
			token = vals[tokenKey]

		} else {
			token, err = getUserInteractiveToken()
		}

		if *updateLocalCredentials {
			data := map[string]string{
				tokenKey: token,
				unameKey: "",
				pwdKey:   "",
			}

			if err = saveLocalstoreData(sm, ns, data); err != nil {
				return "", err
			}
		}
		return token, err
	}
	return "", nil
}

func setAccessCredentialsValues(ns ls.Namespace, supportedMethods map[string]bool) error {
	if _, ok := supportedMethods[*method]; !ok {
		return fmt.Errorf("only %v methods are supported. not %s", supportedMethods, *method)
	}

	if *username == "" || *password == "" {

		var name, pwd string
		var err error
		sm := newStoreManager()

		if sm.IsNamespaceSet(ns) && !*updateLocalCredentials {

			vals, err := sm.GetValues(ns)

			if err != nil {
				return err
			}
			name = vals[unameKey]
			pwd = vals[pwdKey]

		} else {
			name, pwd, err = getUserInteractiveCredentials()
		}

		if *updateLocalCredentials {
			if err = saveLocalstoreCredentials(sm, ns, name, pwd); err != nil {
				return err
			}
		}
		username = &name
		password = &pwd
		return err
	}
	return nil
}

func setVaultAccessCredentialsValues() error {
	if *method != "ldap" {
		return fmt.Errorf("only ldap method supported not %s", *method)
	}
	return setAccessCredentialsValues(VaultLdap, map[string]bool{"ldap": true})
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

		fmt.Printf("to configure run:\n\t export %s_%s=%s", EnvVarPrefix, env, result)
		os.Exit(0)
		return nil
	}
}
func listAllKeychainDetails() error {
	sm := newStoreManager()
	nsToresult, err := sm.ListAll()

	if err != nil {
		return err
	}
	tui := buildTUI()
	table := map[string]string{}
	labels := []string{}
	for nsName, nsRes := range nsToresult {

		if len(nsRes) > 0 {
			fmt.Printf("")
		}
		prefix := string(nsName)
		for k, v := range nsRes {
			var val string
			if log.GetLevel() >= log.DebugLevel {
				val = v
			} else {
				for i := 0; i < len(v); i++ {
					val += "*"
				}
			}
			label := prefix + "." + k
			table[label] = val
			labels = append(labels, label)
		}
	}
	fmt.Printf("Use `-v` flag to see the actual values instead of *\n")
	sort.Strings(labels)
	tui.GetTable().PrintInfoBox(table, labels, false)
	return nil
}

func getUserInteractiveToken() (string, error) {
	validate := func(input string) error {
		if input == "" {
			return errors.New("no empty input allowed")
		}
		return nil
	}
	prompt := promptui.Prompt{
		Label:    "Token",
		Validate: validate,
		Mask:     '*',
	}

	token, err := prompt.Run()

	if err != nil {
		log.Fatalf("Prompt failed %v\n", err)
		return "", err
	}
	return token, nil
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
		VaultLdap:       {unameKey, pwdKey},
		ElasticSearchNS: {tokenKey, unameKey, pwdKey},
		LogzSearchNS:    {tokenKey},
	})
	return sm
}
func setLocalstoreToken(ns ls.Namespace) error {
	sm := newStoreManager()
	token, err := getUserInteractiveToken()

	if err != nil {
		return nil
	}

	data := map[string]string{
		tokenKey: token,
		pwdKey:   "",
		unameKey: "",
	}

	if err := saveLocalstoreData(sm, ns, data); err != nil {
		return err
	}

	return nil
}

func setLocalstoreCredentials(ns ls.Namespace) error {
	sm := newStoreManager()
	name, pwd, err := getUserInteractiveCredentials()

	if err != nil {
		return nil
	}

	if err := saveLocalstoreCredentials(sm, ns, name, pwd); err != nil {
		return err
	}

	return nil
}
func saveLocalstoreData(sm ls.StoreManager[ls.Store], ns ls.Namespace, data map[string]string) error {
	err := sm.SetNSValues(ns, data)
	return err
}

func saveLocalstoreCredentials(sm ls.StoreManager[ls.Store], ns ls.Namespace, name, pwd string) error {

	data := map[string]string{
		unameKey: name,
		pwdKey:   pwd,
		tokenKey: "",
	}
	err := saveLocalstoreData(sm, ns, data)
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
