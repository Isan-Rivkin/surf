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

	es "github.com/isan-rivkin/surf/lib/elastic"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	esToken *string
)

// esCmd represents the es command
var esCmd = &cobra.Command{
	Use:   "es",
	Short: "Search in Elasticsearch database",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("running es")
		isLogz := false
		confBuilder, err := initESConfWithAuth(*username, *password, *esToken, isLogz)
		if err != nil {
			log.WithError(err).Fatal("failed initiating configuration for elastic, please check auth details provided")
		}
		confBuilder.
	},
}

func initESConfWithAuth(uname, pwd, token string, isLogz bool) (*es.ConfigBuilder, error) {
	confBuilder := es.NewConf()
	if uname != "" && pwd != "" {
		return confBuilder.WithUserAuth(uname, pwd), nil
	}
	if !isLogz && token != "" {
		return confBuilder.WithBasicAuthToken(token), nil
	}
	if isLogz && token != "" {
		return confBuilder.WithHeader("X-API-TOKEN", token).WithHeader("Content-Type", "application/json"), nil
	}
	return nil, errors.New("no valid auth credentials provided")
}

func init() {
	esToken = esCmd.PersistentFlags().StringP("token", "t", "", "auth with token")
	rootCmd.AddCommand(esCmd)
}
