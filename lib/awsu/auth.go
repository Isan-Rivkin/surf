package awsu

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudcontrol"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/isan-rivkin/route53-cli/aws_utils"
)

type AuthInput struct {
	Provider         client.ConfigProvider
	Configs          []*aws.Config
	EffectiveRegion  string
	EffectiveProfile string
}

type AWSSessionInput struct {
	Profile string
	Region  string
}

func NewSessionInputMatrix(inputs []*AWSSessionInput) ([]*AuthInput, error) {
	var out []*AuthInput
	for _, input := range inputs {
		auth, err := NewSessionInput(input.Profile, input.Region)
		if err != nil {
			return nil, err
		}
		out = append(out, auth)
	}
	return out, nil
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

	return &AuthInput{Provider: sess, Configs: conf, EffectiveRegion: effectiveRegion, EffectiveProfile: profile}, nil
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

func NewDDB(in *AuthInput) (*dynamodb.DynamoDB, error) {
	ddb := dynamodb.New(in.Provider, in.Configs...)
	if ddb == nil {
		return nil, fmt.Errorf("failed initiating dynamodb instance")
	}
	return ddb, nil
}

func NewCloudControl(in *AuthInput) (*cloudcontrol.Client, error) {
	conf, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(in.EffectiveRegion))
	if err != nil {
		return nil, fmt.Errorf("failed loading aws config %s", err.Error())
	}
	cc := cloudcontrol.NewFromConfig(conf)
	if cc == nil {
		return nil, fmt.Errorf("failed creating cloudcontrol client")
	}
	return cc, nil
}
