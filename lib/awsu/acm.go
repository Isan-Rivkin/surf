package awsu

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/isan-rivkin/surf/lib/common"
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

	pool := common.NewWorkerPool(parallel)

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
					log.WithField("arn", certArn).
						WithError(err).
						Error("failed describing cert")
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
