package awsu

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/isan-rivkin/route53-cli/aws_utils"
	log "github.com/sirupsen/logrus"
)

func Test() {
	profile := "default"
	region := "us-east-1"
	sess := aws_utils.GetEnvSession(profile)
	acmClient := acm.New(sess, aws.NewConfig().WithRegion(region))
	if acmClient == nil {
		log.Error("ACM client is nil")
		return
	}
	// search certificates based on: domain
	// Example iterating over at most 3 pages of a ListCertificates operation.
	pageNum := 0
	input := &acm.ListCertificatesInput{}
	err := acmClient.ListCertificatesPages(input,
		func(page *acm.ListCertificatesOutput, lastPage bool) bool {
			pageNum++
			fmt.Println(page)
			return pageNum <= 3
		})
	if err != nil {
		log.WithError(err).Error("failed searching in acm")
	}

}
