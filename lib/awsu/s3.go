package awsu

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go/aws"
)

type S3API interface {
	ListAllBuckets() ([]types.Bucket, error)
	ListAllObjects(bucket, prefix string) ([]types.Object, error)
}

type S3Client struct {
	c *s3.Client
}

func NewS3Client(c *s3.Client) S3API {
	return &S3Client{
		c: c,
	}
}
func (s *S3Client) client() *s3.Client {
	return s.c
}

func (s *S3Client) ListAllBuckets() ([]types.Bucket, error) {
	in := &s3.ListBucketsInput{}
	resp, err := s.client().ListBuckets(context.TODO(), in)

	if err != nil {
		return nil, fmt.Errorf("failed listing buckets %s", err.Error())
	}

	return resp.Buckets, nil
}

func (s *S3Client) ListAllObjects(bucket, prefix string) ([]types.Object, error) {
	var allObjects []types.Object
	in := &s3.ListObjectsV2Input{
		Bucket:  aws.String(bucket),
		Prefix:  aws.String(prefix),
		MaxKeys: 1000,
	}

	for {
		resp, err := s.client().ListObjectsV2(context.TODO(), in)

		if err != nil {
			return nil, fmt.Errorf("failed listing objects %s", err.Error())
		}

		if resp == nil {
			return nil, fmt.Errorf("failed listing objects s3 respone is nil for some reason")
		}

		allObjects = append(allObjects, resp.Contents...)

		if resp.IsTruncated && resp.NextContinuationToken != nil {
			in.ContinuationToken = resp.NextContinuationToken
		} else {
			break
		}

	}

	return allObjects, nil
}

func GenerateS3WebURL(bucket, region, prefix string) string {
	return fmt.Sprintf("https://s3.console.aws.amazon.com/s3/object/%s?region=%s&prefix=%s", bucket, region, prefix)
}
