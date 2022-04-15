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

	"github.com/isan-rivkin/surf/lib/awsu"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	recordInput *string
)

// r53Cmd represents the r53 command
var r53Cmd = &cobra.Command{
	Use:   "r53",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("r53 called")
		query := "my.website.com"
		debug := true
		mute := true
		skipNSVerification := false
		recurse := true
		recurseDepth := 3
		in, err := awsu.NewR53Input(query, "default", debug, mute, skipNSVerification, recurse, recurseDepth)

		if err != nil {
			log.WithError(err).Fatal("failed creating r53 input")
		}
		_, err = awsu.SearchRoute53(in)

		if err != nil {
			log.WithError(err).Fatal("failed searching r53")
		}

	},
}

func init() {
	rootCmd.AddCommand(r53Cmd)
	recordInput = r53Cmd.PersistentFlags().StringP("domain", "q", "", "target domain to find in R53, wildcard supported")
	r53Cmd.MarkPersistentFlagRequired("domain")
}
