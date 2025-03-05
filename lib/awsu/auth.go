package awsu

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudcontrol"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/isan-rivkin/route53-cli/aws_utils"
	"gopkg.in/ini.v1"
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

func NewCloudFormation(in *AuthInput) (*cloudformation.Client, error) {
	conf, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(in.EffectiveRegion))
	if err != nil {
		return nil, fmt.Errorf("failed loading aws config %s", err.Error())
	}
	cf := cloudformation.NewFromConfig(conf)
	if cf == nil {
		return nil, fmt.Errorf("failed creating cloudformation client")
	}
	return cf, nil
}

type AWSProfileConfig struct {
	profileName    string
	profileConfigs map[string]string
}

func NewAWSProfileConfig(profileName string, configs map[string]string) *AWSProfileConfig {
	return &AWSProfileConfig{profileName: profileName, profileConfigs: configs}
}
func (c *AWSProfileConfig) ProfileName() string {
	return c.profileName
}
func (c *AWSProfileConfig) SetConfig(key, value string) {
	c.profileConfigs[key] = value
}
func (c *AWSProfileConfig) GetConfig(key string) (string, bool) {
	v, exists := c.profileConfigs[key]
	return v, exists
}

func (c *AWSProfileConfig) GetRegion() (string, bool) {
	return c.GetConfig("region")
}
func (c *AWSProfileConfig) GetName() string {
	return c.profileName
}

func GetAllProfilesContext(credsFilePath, configFilePath *string) (string, map[string]*AWSProfileConfig, bool,string, error) {
	// credentials in env precedence over everything
	noProfileAuth := os.Getenv("AWS_ACCESS_KEY_ID") != "" && os.Getenv("AWS_SECRET_ACCESS_KEY") != ""
	defaultProfile := "default"
	// check target profile
	if os.Getenv("AWS_PROFILE") != "" {
		defaultProfile = os.Getenv("AWS_PROFILE")
	}
	envRegion := os.Getenv("AWS_REGION")
	profiles, err := loadMergedProfiles(credsFilePath, configFilePath)
	if err != nil {
		return "", nil, noProfileAuth, "", err
	}
	return defaultProfile, profiles, noProfileAuth,envRegion, nil
}

func loadMergedProfiles(credsFilePath, configFilePath *string) (map[string]*AWSProfileConfig, error) {
	res := map[string]*AWSProfileConfig{}
	credsPath := config.DefaultSharedCredentialsFilename()
	if credsFilePath != nil {
		credsPath = *credsFilePath
	}
	configPath := config.DefaultSharedConfigFilename()
	if configFilePath != nil {
		configPath = *configFilePath
	}
	// Load profiles from both config and credentials files
	profiles := make(map[string]map[string]string)
	if err := loadProfiles(configPath, profiles, true); err != nil {
		return nil, fmt.Errorf("config file: %w", err)
	}
	if err := loadProfiles(credsPath, profiles, false); err != nil {
		return nil, fmt.Errorf("credentials file: %w", err)
	}
	for profileName, profileConfigs := range profiles {
		res[profileName] = NewAWSProfileConfig(profileName, profileConfigs)
	}
	return res, nil
}

// loadProfiles loads profiles from the specified AWS configuration file
func loadProfiles(filePath string, profiles map[string]map[string]string, isConfig bool) error {
	cfg, err := ini.Load(filePath)
	if err != nil {
		return fmt.Errorf("loading file %s: %w", filePath, err)
	}

	for _, section := range cfg.Sections() {
		if section.Name() == ini.DefaultSection {
			continue
		}

		profileName := section.Name()
		if isConfig && strings.HasPrefix(profileName, "profile ") {
			profileName = strings.TrimPrefix(profileName, "profile ")
		}

		if _, exists := profiles[profileName]; !exists {
			profiles[profileName] = make(map[string]string)
		}

		for _, key := range section.Keys() {
			profiles[profileName][key.Name()] = key.Value()
		}
	}
	return nil
}
