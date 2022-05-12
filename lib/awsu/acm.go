package awsu

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/isan-rivkin/route53-cli/aws_utils"
	log "github.com/sirupsen/logrus"
)

type _acmAsyncRes struct {
	C   *acm.CertificateDetail
	Err error
}

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
	return &AcmClient{
		c: c,
	}
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

	filteredResult := &ACMResult{
		Certificates: []*acm.CertificateDetail{},
	}

	certsSummary, err := a.GetAll()

	if err != nil {
		return nil, err
	}

	pool := NewWorkerPool(parallel)
	asyncResults := make(chan *_acmAsyncRes, len(certsSummary))

	for _, cs := range certsSummary {

		certArn := aws.StringValue(cs.CertificateArn)

		if describe {
			pool.Submit(func() {
				reqInput := &acm.DescribeCertificateInput{
					CertificateArn: aws.String(certArn),
				}
				req, out := a.client().DescribeCertificateRequest(reqInput)

				var cert *acm.CertificateDetail
				err := req.Send()
				if err != nil || out == nil {
					log.WithField("arn", aws.StringValue(reqInput.CertificateArn)).WithError(err).Error("failed describing ceritificate")
				} else {
					cert = out.Certificate
				}
				asyncResults <- &_acmAsyncRes{C: cert, Err: err}
			})

		} else {
			cert := &acm.CertificateDetail{CertificateArn: aws.String(certArn)}
			result.Certificates = append(result.Certificates, cert)
		}
	}
	if describe {
		pool.RunAll()
		size := len(certsSummary)
		counter := 1
		for r := range asyncResults {
			if counter >= size {
				break
			}
			counter++

			if r.Err != nil {
				continue
			}
			if isMatch := filter(r.C); isMatch {
				filteredResult.Certificates = append(filteredResult.Certificates, r.C)
			}
		}
	} else {
		for _, cert := range result.Certificates {
			if isMatch := filter(cert); isMatch {
				filteredResult.Certificates = append(filteredResult.Certificates, cert)
			}
		}
	}

	return filteredResult, err
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
