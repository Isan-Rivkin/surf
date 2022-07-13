package awsu

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/isan-rivkin/route53-cli/aws_utils"
)

type AuthInput struct {
	Provider        client.ConfigProvider
	Configs         []*aws.Config
	EffectiveRegion string
}

func NewSessionInput(profile, region string) (*AuthInput, error) {
	sess := aws_utils.GetEnvSession(profile)
	if sess == nil {
		return nil, fmt.Errorf("failed creating env sessions profile %s region %s", profile, region)
	}
	effectiveRegion := aws.StringValue(sess.Config.Region)

	c := aws.NewConfig()
	if region != "" {
		c = c.WithRegion(region)
		effectiveRegion = region
	}
	conf := []*aws.Config{c}

	return &AuthInput{Provider: sess, Configs: conf, EffectiveRegion: effectiveRegion}, nil
}

func NewACM(in *AuthInput) (*acm.ACM, error) {
	a := acm.New(in.Provider, in.Configs...)
	if a == nil {
		return nil, fmt.Errorf("failed creating acm client")
	}
	return a, nil
}

func NewS3(in *AuthInput) (*s3.Client, error) {
	conf, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(in.EffectiveRegion))

	if err != nil {
		return nil, fmt.Errorf("failed loading aws config %s", err.Error())
	}

	s := s3.NewFromConfig(conf)

	if s == nil {
		return nil, fmt.Errorf("failed creating s3 client")
	}
	return s, nil
}
