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
	"os/signal"
	"strings"

	v "github.com/isan-rivkin/cliversioner"
	"github.com/isan-rivkin/surf/printer"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

const (
	AppVersion = "2.0.0"
	AppName    = "surf"
)

var (
	cfgFile      string
	verboseLevel *int
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     AppName,
	Short:   "Free Text Search across your infrastructure platforms via CLI.",
	Long:    getEnvVarConfig(),
	Version: AppVersion,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		setLogLevel()
		go VersionCheck()
	},
	// Run: func(cmd *cobra.Command, args []string) {

	// },
}

func buildTUI() printer.TuiController[printer.Loader, printer.Table] {
	s := &printer.SpinnerApi{}
	t := printer.NewTablePrinter()
	tui := printer.NewPrinter[printer.Loader, printer.Table](s, t)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		sig := <-c
		log.WithField("signal", sig).Debug("shutting down gracefully")
		tui.GetLoader().Stop()
		os.Exit(0)
	}()
	return tui
}

func getDefaultProfileEnvVar() string {
	profile := os.Getenv("AWS_PROFILE")
	if profile != "" {
		return profile
	}
	return "default"
}

func getAppEnvVarStr(envName string) string {
	return viper.GetString(envName)
}

func getAppEnvVarStrSlice(envName string) []string {
	strVal := viper.GetStringSlice(envName)
	if len(strVal) == 1 {
		return strings.Split(strVal[0], ",")
	}
	return strVal
}

func getEnvStrSliceOrOverride(flagVal *[]string, envName string) []string {
	if flagVal != nil && len(*flagVal) == 0 {
		return getAppEnvVarStrSlice(envName)
	}
	return *flagVal
}

func getEnvOrOverride(flagVal *string, envName string) *string {
	v := viper.GetString(envName)
	if v != "" && *flagVal == "" {
		return &v
	}
	return flagVal
}

func getEnvVarConfig() string {
	m := `
	Environment Variables Available: 

`
	for _, e := range confEnvVars {
		m += fmt.Sprintf("\t%s_%s \n\t%s\n\n", EnvVarPrefix, e.Value, e.Description)

	}
	return m
}

func setLogLevel() {
	lvl := getLogLevelFromVerbosity()
	log.SetLevel(lvl)
	if lvl >= log.TraceLevel {
		log.SetReportCaller(true)
	}
}
func getLogLevelFromVerbosity() log.Level {
	switch *verboseLevel {
	case 0:
		return log.InfoLevel
	case 1:
		return log.DebugLevel
	default:
		return log.TraceLevel
	}
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func VersionCheck() {
	var err error
	optoutVar := fmt.Sprintf("%s_%s", EnvVarPrefix, EnvVersionCheckOptout)
	i := v.NewInput(AppName, "https://github.com/isan-rivkin", AppVersion, &optoutVar)
	out, err := v.CheckVersion(i)

	if err != nil || out == nil {
		log.WithError(err).Debug("failed checking latest version from github.com")
		return
	}

	if out.Outdated {
		m := fmt.Sprintf("%s is not latest, %s, upgrade to %s", out.CurrentVersion, out.Message, out.LatestVersion)
		log.Warn(m)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	//rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.vault-searcher.yaml)")
	verboseLevel = rootCmd.PersistentFlags().CountP("verbose", "v", "verbosity level -vvv")

	// configure auth related
	password = rootCmd.PersistentFlags().StringP("password", "s", "", "store password for future auth locally on your OS keyring")
	username = rootCmd.PersistentFlags().StringP("username", "u", "", "store username for future auth locally on your OS keyring")
	updateLocalCredentials = rootCmd.PersistentFlags().Bool("update-creds", false, "update credentials locally on your OS keyring")
	method = rootCmd.PersistentFlags().StringP("auth", "a", "ldap", "authentication method")

}

const (
	EnvVarPrefix             string = "SURF"
	EnvKeyVaultDefaultPrefix string = "VAULT_DEFAULT_PREFIX"
	EnvKeyVaultDefaultMount  string = "VAULT_DEFAULT_MOUNT"
	EnvKeyS3DefaultBucket    string = "S3_DEFAULT_MOUNT"
	EnvVersionCheckOptout    string = "VERSION_CHECK"
	EnvElasticsearchURL      string = "ELASTICSEARCH_URL"
	EnvElasticsearchUsername string = "ELASTICSEARCH_USERNAME"
	EnvElasticsearchPwd      string = "ELASTICSEARCH_PASSWORD"
	EnvElasticsearchToken    string = "ELASTICSEARCH_TOKEN"
	EnvElasticsearchIndexes  string = "ELASTICSEARCH_INDEXES"
	EnvLogzIOToken           string = "LOGZ_IO_TOKEN"
	EnvLogzIOURL             string = "LOGZ_IO_URL"
	EnvLogzIOSubAccountIDs   string = "LOGZ_IO_ACCOUNT_IDS"
)

var confEnvVars = []struct {
	Value       string
	Description string
}{
	{
		Value:       EnvKeyVaultDefaultMount,
		Description: "Mount to start the search from in Vault",
	},
	{
		Value:       EnvKeyVaultDefaultPrefix,
		Description: "Prefix to start the search from in Vault appended to mount",
	},
	{
		Value:       EnvVersionCheckOptout,
		Description: "if set true the tool will skip latest version check from github.com",
	},
	{
		Value:       EnvKeyS3DefaultBucket,
		Description: "if set this bucket will be searched by default",
	},
	{
		Value:       EnvElasticsearchURL,
		Description: "if set for elastic command will be used, will override standard ELASTICSEARCH_URL if set",
	},
	{
		Value:       EnvElasticsearchUsername,
		Description: "if set will be used for authentication with elasticsearch (must add password)",
	},
	{
		Value:       EnvElasticsearchPwd,
		Description: "if set will be used for authentication with elasticsearch (must add username)",
	},
	{
		Value:       EnvElasticsearchToken,
		Description: "if set will be used for authentication, if exist with name/password conflict will use token",
	},
	{
		Value:       EnvElasticsearchIndexes,
		Description: "command separated list of indexes to search for in elasticsearch",
	},
	{
		Value:       EnvLogzIOToken,
		Description: "logz.io token, must have permissions to search in sub-accounts and list accounts",
	},
	{
		Value:       EnvLogzIOURL,
		Description: "logz.io url, check https://docs.logz.io/user-guide/accounts/account-region.html for more info",
	},
	{
		Value:       EnvLogzIOSubAccountIDs,
		Description: "logz.io sub-account ids tp search in comma-separated",
	},
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".surf" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".surf")
	}

	viper.SetEnvPrefix(EnvVarPrefix)

	for _, v := range confEnvVars {
		viper.BindEnv(v.Value)
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
