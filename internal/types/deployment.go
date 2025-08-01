package types

import "github.com/digitalocean/godo"

type DeploymentEnvironment struct {
	ResourceNames ResourceNames
	Frontend      FrontendEnvironment
	Backend       BackendEnvironment
}

type ResourceNames struct {
	S3Bucket         string
	DatabaseMain     string
	DatabaseHangfire string
	DOApp            string
	VercelProject    string
	FrontendURL      string
	BackendURL       string
}

type FrontendEnvironment struct {
	S3URL               string
	StrapiURL           string
	DefaultLanguage     string
	FrontURL            string
	ValidationTime      string
	AmazonEnv           string
	RevalidationToken   string
	ForcedLoginTimeInMs string
}

type BackendEnvironment struct {
	AppLevelVars       map[string]godo.AppVariableDefinition
	ComponentLevelVars map[string]godo.AppVariableDefinition
}

type DeploymentDefaults struct {
	JWT           JWTDefaults
	Frontend      FrontendDefaults
	S3Path        string
	GitRepository GitRepositoryDefaults
}

type JWTDefaults struct {
	Issuer            string
	Audience          string
	Key               string
	ExpirationMinutes string
}

type FrontendDefaults struct {
	DefaultLanguage     string
	ValidationTime      string
	AmazonEnv           string
	RevalidationToken   string
	ForcedLoginTimeInMs string
	BuildCommand        string
	InstallCommand      string
	FrameworkPreset     string
}

type GitRepositoryDefaults struct {
	BackendRepo  string
	FrontendRepo string
	Branch       string
}
