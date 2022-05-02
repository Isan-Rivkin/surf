package awsu

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/isan-rivkin/route53-cli/aws_utils"
)

type AuthInput struct {
	Provider client.ConfigProvider
	Configs  []*aws.Config
}

func NewSessionInput(profile, region string) (*AuthInput, error) {
	sess := aws_utils.GetEnvSession(profile)
	if sess == nil {
		return nil, fmt.Errorf("failed creating env sessions profile %s region %s", profile, region)
	}

	c := aws.NewConfig()
	if region != "" {
		c = c.WithRegion(region)
	}
	conf := []*aws.Config{c}
	return &AuthInput{Provider: sess, Configs: conf}, nil
}

func NewACM(in *AuthInput) (*acm.ACM, error) {
	a := acm.New(in.Provider, in.Configs...)
	if a == nil {
		return nil, fmt.Errorf("failed creating acm client")
	}
	return a, nil
}
