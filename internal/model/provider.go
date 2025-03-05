package model

import (
	"github.com/isan-rivkin/surf/lib/awsu"
)

// provider is an interface that represents a single platform + all context details
// For example: Provider=AWS, Context=Region=us-west-2, Account=1234567890
type Provider interface {
	Name() string
	CurrentContext() string
	Init() error
}

type MockProvider struct {
}

func (m *MockProvider) Name() string {
	return "MockAWS"
}
func (m *MockProvider) CurrentContext() string {
	return "MockAWS/Region=us-west-2, Account=1234567890"
}
func (m *MockProvider) Init() error {
	return nil
}

type AWSProvider struct {
	profiles       map[string]*awsu.AWSProfileConfig
	defaultProfile string
	// if true it means creds are used in env
	isAnnonymousEnvPCredsSet bool
	// if not empty it means region is set in env and overrides
	envRegion string
}

func (a *AWSProvider) Name() string {
	return "AWS"
}
func (a *AWSProvider) CurrentContext() string {
	return "Region=us-west-2, Account=1234567890"
}

func (a *AWSProvider) Init() error {
	defaultProfile, profiles, isAnnonymousEnvPCredsSet, envRegion, err := awsu.GetAllProfilesContext(nil, nil)
	if err != nil {
		return err
	}
	a.profiles = profiles
	a.defaultProfile = defaultProfile
	a.isAnnonymousEnvPCredsSet = isAnnonymousEnvPCredsSet
	a.envRegion = envRegion
	return nil
}
