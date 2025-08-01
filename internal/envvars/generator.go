package envvars

import (
	"fmt"

	"github.com/CaioDGallo/easy-cli/internal/config"
	"github.com/CaioDGallo/easy-cli/internal/resources"
	"github.com/CaioDGallo/easy-cli/internal/types"
	"github.com/digitalocean/godo"
)

func GenerateDeploymentEnvironment(client types.Client, cfg *config.Config) (types.DeploymentEnvironment, error) {
	resourceNames := resources.GenerateResourceNames(client.SanitizedClientName, cfg)
	defaults := resources.GetDeploymentDefaults(cfg)

	if err := resources.ValidateResourceNames(resourceNames); err != nil {
		return types.DeploymentEnvironment{}, fmt.Errorf("invalid resource names: %w", err)
	}

	frontend := generateFrontendEnvironment(resourceNames, defaults, client)
	backend := generateBackendEnvironment(resourceNames, defaults, client)

	return types.DeploymentEnvironment{
		ResourceNames: resourceNames,
		Frontend:      frontend,
		Backend:       backend,
	}, nil
}

func generateFrontendEnvironment(resourceNames types.ResourceNames, defaults types.DeploymentDefaults, client types.Client) types.FrontendEnvironment {
	backendURL := resourceNames.BackendURL
	if client.BackendInfo.URL != "" {
		backendURL = client.BackendInfo.URL
	}

	frontendURL := resourceNames.FrontendURL
	if client.FrontendInfo.URL != "" {
		frontendURL = client.FrontendInfo.URL
	}

	return types.FrontendEnvironment{
		S3URL:               fmt.Sprintf("https://%s.s3.amazonaws.com", resourceNames.S3Bucket),
		StrapiURL:           backendURL,
		DefaultLanguage:     defaults.Frontend.DefaultLanguage,
		FrontURL:            frontendURL,
		ValidationTime:      defaults.Frontend.ValidationTime,
		AmazonEnv:           defaults.Frontend.AmazonEnv,
		RevalidationToken:   defaults.Frontend.RevalidationToken,
		ForcedLoginTimeInMs: defaults.Frontend.ForcedLoginTimeInMs,
	}
}

func generateBackendEnvironment(resourceNames types.ResourceNames, defaults types.DeploymentDefaults, client types.Client) types.BackendEnvironment {
	appLevelEnvVars := map[string]godo.AppVariableDefinition{}
	componentLevelEnvVars := map[string]godo.AppVariableDefinition{}

	appLevelVars := map[string]string{
		"Security_JWT_Issuer":            defaults.JWT.Issuer,
		"Security_JWT_Audience":          defaults.JWT.Audience,
		"Security_JWT_Key":               defaults.JWT.Key,
		"Security_JWT_ExpirationMinutes": defaults.JWT.ExpirationMinutes,
	}

	for key, value := range appLevelVars {
		appLevelEnvVars[key] = godo.AppVariableDefinition{
			Key:   key,
			Value: value,
			Scope: godo.AppVariableScope_RunAndBuildTime,
			Type:  godo.AppVariableType_General,
		}
	}

	frontendURL := resourceNames.FrontendURL
	if client.FrontendInfo.URL != "" {
		frontendURL = client.FrontendInfo.URL
	}

	backendURL := resourceNames.BackendURL
	if client.BackendInfo.URL != "" {
		backendURL = client.BackendInfo.URL
	}

	componentLevelVars := map[string]string{
		"ConnectionStrings__NextGenDBContext": fmt.Sprintf("Server=%s;Database=%s;Username=%s;Password=%s;IncludeErrorDetail=true",
			client.DatabaseHost, resourceNames.DatabaseMain, client.DatabaseUser, client.BackendInfo.DatabasePassword),
		"ConnectionStrings__Hangfire": fmt.Sprintf("Server=%s;Database=%s;Username=%s;Password=%s",
			client.DatabaseHost, resourceNames.DatabaseHangfire, client.DatabaseUser, client.BackendInfo.DatabasePassword),
		"SMTP_Server":          client.SMTPInfo.Server,
		"SMTP_Port":            client.SMTPInfo.Port,
		"SMTP_Username":        client.SMTPInfo.Username,
		"SMTP_Password":        client.SMTPInfo.Password,
		"SMTP_DoNotReplyName":  client.SMTPInfo.DoNotReplyName,
		"SMTP_DoNotReplyEmail": client.SMTPInfo.DoNotReplyEmail,
		"SMTP_DevEmail":        client.SMTPInfo.DevEmail,
		"Paths_MediaPath":      fmt.Sprintf("https://%s.s3.amazonaws.com", resourceNames.S3Bucket),
		"Paths_FrontEndPath":   frontendURL,
		"Paths_BackendPath":    backendURL,
		"AWS_S3_BUCKET":        resourceNames.S3Bucket,
		"AWS_BUCKET_NAME":      resourceNames.S3Bucket,
		"AWS_S3_PATH":          defaults.S3Path,
	}

	for key, value := range componentLevelVars {
		componentLevelEnvVars[key] = godo.AppVariableDefinition{
			Key:   key,
			Value: value,
			Scope: godo.AppVariableScope_RunAndBuildTime,
			Type:  godo.AppVariableType_General,
		}
	}

	return types.BackendEnvironment{
		AppLevelVars:       appLevelEnvVars,
		ComponentLevelVars: componentLevelEnvVars,
	}
}

func GenerateVercelEnvironmentVariables(client types.Client, cfg *config.Config) ([]types.VercelEnvVariable, error) {
	resourceNames := resources.GenerateResourceNames(client.SanitizedClientName, cfg)
	defaults := resources.GetDeploymentDefaults(cfg)

	if err := resources.ValidateResourceNames(resourceNames); err != nil {
		return nil, fmt.Errorf("invalid resource names: %w", err)
	}

	frontend := generateFrontendEnvironment(resourceNames, defaults, client)

	return []types.VercelEnvVariable{
		{
			Key:    "NEXT_PUBLIC_S3_URL",
			Target: "production",
			Value:  frontend.S3URL,
			Type:   "plain",
		},
		{
			Key:    "NEXT_PUBLIC_STRAPI",
			Target: "production",
			Value:  frontend.StrapiURL,
			Type:   "plain",
		},
		{
			Key:    "NEXT_PUBLIC_DEFAULT_LANGUAGE",
			Target: "production",
			Value:  frontend.DefaultLanguage,
			Type:   "plain",
		},
		{
			Key:    "NEXT_PUBLIC_FRONT",
			Target: "production",
			Value:  frontend.FrontURL,
			Type:   "plain",
		},
		{
			Key:    "NEXT_PUBLIC_VALIDATION",
			Target: "production",
			Value:  frontend.ValidationTime,
			Type:   "encrypted",
		},
		{
			Key:    "NEXT_PUBLIC_AMAZON_ENV",
			Target: "production",
			Value:  frontend.AmazonEnv,
			Type:   "plain",
		},
		{
			Key:    "NEXT_PUBLIC_REVALIDATION_TOKEN",
			Target: "production",
			Value:  frontend.RevalidationToken,
			Type:   "encrypted",
		},
		{
			Key:    "NEXT_PUBLIC_FORCED_LOGIN_TIME_IN_MS",
			Target: "production",
			Value:  frontend.ForcedLoginTimeInMs,
			Type:   "plain",
		},
	}, nil
}
