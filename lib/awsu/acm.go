package awsu

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/isan-rivkin/route53-cli/aws_utils"
	log "github.com/sirupsen/logrus"
)

type ACMFilter = func(c *acm.CertificateDetail) bool

type ACMResult struct {
	Certificates []*acm.CertificateDetail
}

type AcmAPI interface {
	ListAndFilter(parallel int, describe bool, filter ACMFilter) (*ACMResult, error)
}

type AcmClient struct {
	c *acm.ACM
}

func NewAcmClient(c *acm.ACM) AcmAPI {
	return &AcmClient{c: c}
}

func (a *AcmClient) client() *acm.ACM {
	return a.c
}

func (a *AcmClient) GetAll() ([]*acm.CertificateSummary, error) {
	result := []*acm.CertificateSummary{}

	input := &acm.ListCertificatesInput{}

	err := a.client().ListCertificatesPages(input,
		func(page *acm.ListCertificatesOutput, lastPage bool) bool {
			result = append(result, page.CertificateSummaryList...)
			return !lastPage
		})

	return result, err
}

func (a *AcmClient) ListAndFilter(parallel int, describe bool, filter ACMFilter) (*ACMResult, error) {

	result := &ACMResult{
		Certificates: []*acm.CertificateDetail{},
	}

	certsSummary, err := a.GetAll()

	if err != nil {
		return nil, err
	}

	for _, cs := range certsSummary {
		certArn := aws.StringValue(cs.CertificateArn)
		var cert *acm.CertificateDetail

		if describe {
			reqInput := &acm.DescribeCertificateInput{
				CertificateArn: aws.String(certArn),
			}

			out, err := a.client().DescribeCertificate(reqInput)
			if err != nil || out == nil {
				log.WithField("arn", aws.StringValue(reqInput.CertificateArn)).WithError(err).Error("failed describing ceritificate")
			}
			cert = out.Certificate
		} else {
			cert = &acm.CertificateDetail{CertificateArn: aws.String(certArn)}
		}

		if isMatch := filter(cert); isMatch {
			result.Certificates = append(result.Certificates, cert)
		}
	}

	return result, err
}

func GenerateACMWebURL(region, certId string) string {
	return fmt.Sprintf("https://console.aws.amazon.com/acm/home?region=%s#/certificates/%s", region, certId)
}
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
			//fmt.Println(page)
			for _, pc := range page.CertificateSummaryList {
				certArn := aws.StringValue(pc.CertificateArn)
				FilterCert(acmClient, certArn)
			}
			return pageNum <= 3
		})
	if err != nil {
		log.WithError(err).Error("failed searching in acm")
	}

}

func FilterCert(acmClient *acm.ACM, certArn string) {
	reqInput := &acm.DescribeCertificateInput{
		CertificateArn: aws.String(certArn),
	}

	out, err := acmClient.DescribeCertificate(reqInput)
	if err != nil || out == nil {
		log.WithField("arn", aws.StringValue(reqInput.CertificateArn)).WithError(err).Error("failed describing ceritificate")
	}
	cert := out.Certificate
	//alternativeNames := aws.StringValueSlice(cert.SubjectAlternativeNames)
	fmt.Println(cert)
	fmt.Println(aws.StringValueSlice(cert.InUseBy))
	splittedArn := strings.Split(certArn, "/")

	certId := splittedArn[len(splittedArn)-1]
	region := "us-east-1"
	url := GenerateACMWebURL(region, certId)
	fmt.Println(url)
	// fmt.Println("DOMAIN = ", aws.StringValue(cert.DomainName), aws.StringValue(cert.Status))
	// fmt.Println("===> ", alternativeNames)
}
