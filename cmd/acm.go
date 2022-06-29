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
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/isan-rivkin/surf/lib/awsu"
	search "github.com/isan-rivkin/surf/lib/search/vaultsearch"
	"github.com/isan-rivkin/surf/printer"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	awsRegion                  string
	filterQuery                string
	acmFilterDomains           *bool
	acmFilterID                *bool
	acmFilterAttachedResources *bool
	acmFilterAllOptions        *bool
)

// acmCmd represents the acm command
var acmCmd = &cobra.Command{
	Use:   "acm",
	Short: "search in ACM certificates on AWS",
	Long: `Options to search:

	- Domain Based (default)

	surf acm -q my-domain.com

	- Based on Resources using the certificate 

	surf acm -q some-load-balancer-arn --filter-used-by

	- Certificate ID 

	surf acm -q some-acm-id --filter-id
`,
	Run: func(cmd *cobra.Command, args []string) {
		s := &printer.SpinnerApi{}
		t := printer.NewTablePrinter()
		tui := printer.NewPrinter[printer.Loader, printer.Table](s, t)

		auth, err := awsu.NewSessionInput(awsProfile, awsRegion)

		if err != nil {
			log.Panicf("failed creating session in AWS %s", err.Error())
		}

		acmClient, err := awsu.NewACM(auth)

		if err != nil {
			log.Panicf("failed creating ACM client %s", err.Error())
		}

		api := awsu.NewAcmClient(acmClient)
		parallel := 20
		m := search.NewDefaultRegexMatcher()

		tui.GetLoader().Start("searching acm", "", "green")

		result, err := api.ListAndFilter(parallel, true, func(c *acm.CertificateDetail) bool {
			if *acmFilterAllOptions {
				*acmFilterAttachedResources = true
				*acmFilterDomains = true
				*acmFilterID = true
			}

			if *acmFilterDomains {
				domains := aws.StringValueSlice(c.SubjectAlternativeNames)
				for _, d := range domains {
					if isMatch, _ := m.IsMatch(filterQuery, d); isMatch {
						return true
					}
				}
			}
			if *acmFilterAttachedResources {
				usedBy := aws.StringValueSlice(c.InUseBy)
				for _, arn := range usedBy {
					if isMatch, _ := m.IsMatch(filterQuery, arn); isMatch {
						return true
					}
				}
			}
			if *acmFilterID {
				if isMatch, _ := m.IsMatch(filterQuery, aws.StringValue(c.CertificateArn)); isMatch {
					return true
				}
			}
			return false
		})

		tui.GetLoader().Stop()
		certs := result.Certificates
		sort.SliceStable(certs, func(i, j int) bool {
			c1 := certs[i]
			c2 := certs[j]

			c1Create := aws.TimeValue(c1.CreatedAt)
			c2Create := aws.TimeValue(c2.CreatedAt)

			return c2Create.After(c1Create)

		})

		for _, c := range result.Certificates {

			arn := aws.StringValue(c.CertificateArn)
			splitted := strings.Split(arn, "/")
			id := splitted[len(splitted)-1]
			url := awsu.GenerateACMWebURL(auth.EffectiveRegion, id)
			status := aws.StringValue(c.Status)
			domain := aws.StringValue(c.DomainName)
			inUseBy := aws.StringValueSlice(c.InUseBy)
			created := aws.TimeValue(c.CreatedAt)
			notAfter := aws.TimeValue(c.NotAfter)

			// date expiration

			expireDays := notAfter.Sub(time.Now()).Hours() / 24

			// status pretty output consolidation
			validationMethodsMapper := map[string]bool{}
			validationStatusMapper := map[string]bool{}
			validationMethods := ""
			validationStatus := ""
			if c.DomainValidationOptions != nil {
				for _, o := range c.DomainValidationOptions {
					m := aws.StringValue(o.ValidationMethod)
					validationMethodsMapper[m] = true

					s := aws.StringValue(o.ValidationStatus)
					validationStatusMapper[s] = true
				}
			}

			if len(validationStatusMapper) > 1 {
				validationStatus = "Partial"
			} else {
				for s := range validationStatusMapper {
					validationStatus = s
				}
			}

			for m := range validationMethodsMapper {
				validationMethods += m + " |"
			}

			labelsOrder := []string{"Domain", "URL", "Status"}

			certInfo := map[string]string{
				"Domain": domain,
				"URL":    url,
				"Status": status,
			}

			if getLogLevelFromVerbosity() >= log.DebugLevel {
				labelsOrder = append(labelsOrder, []string{"Created", "Expire In", "Validation"}...)
				certInfo["Created"] = created.String()
				certInfo["Expire In"] = fmt.Sprintf("%d", int(expireDays))
				certInfo["Validation"] = fmt.Sprintf("%s [%s]", validationMethods, validationStatus)
				for i, arn := range inUseBy {
					useByLabel := fmt.Sprintf("Used By %d", i)
					certInfo[useByLabel] = arn
					labelsOrder = append(labelsOrder, useByLabel)
				}
			}

			tui.GetTable().PrintInfoBox(certInfo, labelsOrder)
		}
	},
}

func init() {
	rootCmd.AddCommand(acmCmd)

	acmCmd.PersistentFlags().StringVarP(&awsProfile, "profile", "p", getDefaultProfileEnvVar(), "~/.aws/credentials chosen account")
	acmCmd.PersistentFlags().StringVarP(&awsRegion, "region", "r", "", "~/.aws/config default region if empty")
	acmCmd.PersistentFlags().StringVarP(&filterQuery, "query", "q", "", "filter query regex supported")

	acmFilterDomains = acmCmd.PersistentFlags().Bool("filter-domains", true, "compare query input against all subject names i.e domains")
	acmFilterID = acmCmd.PersistentFlags().Bool("filter-id", false, "compare query input against all acm arn's")
	acmFilterAttachedResources = acmCmd.PersistentFlags().Bool("filter-used-by", false, "compare query input against arn's using the acm certificate i.e load balancer")
	acmFilterAllOptions = acmCmd.PersistentFlags().Bool("filter-all", false, "if true the query will filter against all the filter options")

	acmCmd.MarkPersistentFlagRequired("query")
}
