package digitalocean

import (
	"context"
	"fmt"
	"time"

	"github.com/CaioDGallo/easy-cli/internal/config"
	"github.com/CaioDGallo/easy-cli/internal/interfaces"
	"github.com/CaioDGallo/easy-cli/internal/logger"
	"github.com/CaioDGallo/easy-cli/internal/retry"
	"github.com/CaioDGallo/easy-cli/internal/types"
	"github.com/digitalocean/godo"
	"github.com/sirupsen/logrus"
)

var _ interfaces.AppHostingProvider = (*AppService)(nil)

type AppService struct {
	client *godo.Client
}

func NewAppService(token string) *AppService {
	return &AppService{
		client: godo.NewFromToken(token),
	}
}

func (a *AppService) CreateApp(ctx context.Context, client types.Client, envVars types.DigitalOceanEnvVars, cfg *config.Config) (string, error) {
	appLevelEnvs := make([]*godo.AppVariableDefinition, 0, len(envVars.AppEnvs))
	for _, value := range envVars.AppEnvs {
		valueCopy := value
		appLevelEnvs = append(appLevelEnvs, &valueCopy)
	}

	createAppReq := godo.AppCreateRequest{
		Spec: &godo.AppSpec{
			Name:   fmt.Sprintf("%s-%s", cfg.Application.NamePrefix, client.SanitizedClientName),
			Envs:   appLevelEnvs,
			Region: "sfo",
		},
	}

	app, _, err := a.client.Apps.Create(ctx, &createAppReq)
	if err != nil {
		return "", fmt.Errorf("failed to create DigitalOcean app: %w", err)
	}

	componentLevelEnvs := make([]*godo.AppVariableDefinition, 0, len(envVars.ComponentEnvs))
	for _, value := range envVars.ComponentEnvs {
		valueCopy := value
		componentLevelEnvs = append(componentLevelEnvs, &valueCopy)
	}

	component := &godo.AppServiceSpec{
		Name:           client.SanitizedClientName,
		SourceDir:      "/",
		DockerfilePath: "Dockerfile",
		Bitbucket: &godo.BitbucketSourceSpec{
			Repo:         cfg.Repository.Backend,
			Branch:       client.BackendBranch,
			DeployOnPush: true,
		},
		HTTPPort:         80,
		InstanceCount:    1,
		InstanceSizeSlug: "basic-xxs",
		Envs:             componentLevelEnvs,
	}

	app.Spec.Services = append(app.Spec.Services, component)
	updateRequest := &godo.AppUpdateRequest{Spec: app.Spec}

	if _, _, err := a.client.Apps.Update(ctx, app.ID, updateRequest); err != nil {
		return "", fmt.Errorf("failed to update DigitalOcean app with service: %w", err)
	}

	appURL, err := a.WaitForAppDeploymentAndGetURL(ctx, app.ID)
	if err != nil {
		return "", fmt.Errorf("failed to wait for app deployment and get URL: %w", err)
	}

	return appURL, nil
}

func (a *AppService) GetAppURL(ctx context.Context, appID string) (string, error) {
	log := logger.WithFields(logrus.Fields{
		"app_id": appID,
		"action": "get_app_url",
	})

	retryConfig := retry.Config{
		MaxAttempts: 10,
		Delay:       5 * time.Second,
		Backoff:     2 * time.Second,
	}

	var appURL string
	err := retry.Do(ctx, retryConfig, func() error {
		app, _, err := a.client.Apps.Get(ctx, appID)
		if err != nil {
			return fmt.Errorf("failed to get app: %w", err)
		}

		liveURL := app.GetLiveURL()
		if liveURL == "" {
			log.Debug("App URL not available yet, retrying")
			return fmt.Errorf("app URL not available yet")
		}

		appURL = liveURL
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to get app URL after retries: %w", err)
	}

	return appURL, nil
}

func (a *AppService) UpdateAppEnvironmentVariables(ctx context.Context, appName string, envVars types.DigitalOceanEnvVars, cfg *config.Config) error {
	apps, _, err := a.client.Apps.List(ctx, &godo.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list apps: %w", err)
	}

	var targetApp *godo.App
	for _, app := range apps {
		if app.Spec.Name == fmt.Sprintf("%s-%s", cfg.Application.NamePrefix, appName) {
			targetApp = app
			break
		}
	}

	if targetApp == nil {
		return fmt.Errorf("app not found: %s-%s", cfg.Application.NamePrefix, appName)
	}

	targetApp.Spec.Envs = make([]*godo.AppVariableDefinition, 0, len(envVars.AppEnvs))
	for _, value := range envVars.AppEnvs {
		valueCopy := value
		targetApp.Spec.Envs = append(targetApp.Spec.Envs, &valueCopy)
	}

	if len(targetApp.Spec.Services) > 0 {
		targetApp.Spec.Services[0].Envs = make([]*godo.AppVariableDefinition, 0, len(envVars.ComponentEnvs))
		for _, value := range envVars.ComponentEnvs {
			valueCopy := value
			targetApp.Spec.Services[0].Envs = append(targetApp.Spec.Services[0].Envs, &valueCopy)
		}

		if targetApp.Spec.Services[0].DockerfilePath == "" {
			targetApp.Spec.Services[0].DockerfilePath = "Dockerfile"
		}
	}

	updateRequest := &godo.AppUpdateRequest{Spec: targetApp.Spec}
	if _, _, err := a.client.Apps.Update(ctx, targetApp.ID, updateRequest); err != nil {
		return fmt.Errorf("failed to update app environment variables: %w", err)
	}

	return nil
}

func (a *AppService) WaitForAppDeploymentAndGetURL(ctx context.Context, appID string) (string, error) {
	log := logger.WithFields(logrus.Fields{
		"app_id": appID,
		"action": "wait_for_deployment",
	})

	log.Info("Waiting for DigitalOcean app deployment to complete")

	retryConfig := retry.Config{
		MaxAttempts: 45,
		Delay:       20 * time.Second,
		Backoff:     0,
	}

	var appURL string
	err := retry.Do(ctx, retryConfig, func() error {
		app, _, err := a.client.Apps.Get(ctx, appID)
		if err != nil {
			return fmt.Errorf("failed to get app: %w", err)
		}

		if app.InProgressDeployment == nil && app.LastDeploymentActiveAt.IsZero() {
			log.Debug("App deployment not started yet")
			return fmt.Errorf("app deployment not started yet")
		}

		if app.InProgressDeployment != nil && app.InProgressDeployment.ID != "" {
			deployment, _, err := a.client.Apps.GetDeployment(ctx, appID, app.InProgressDeployment.ID)
			if err != nil {
				log.WithError(err).Warn("Failed to get deployment details, will retry")
				return fmt.Errorf("failed to get deployment: %w", err)
			}

			phase := deployment.GetPhase()
			log.WithField("phase", phase).Debug("Deployment phase status")

			if phase == "ERROR" || phase == "CANCELED" {
				log.WithField("phase", phase).Error("Deployment failed")
				return fmt.Errorf("deployment failed with phase: %s", phase)
			}

			if phase != "ACTIVE" {
				log.WithField("phase", phase).Info("Deployment still in progress")
				return fmt.Errorf("deployment still in progress, phase: %s", phase)
			}
		}

		liveURL := app.GetLiveURL()
		if liveURL == "" {
			log.Debug("App URL not available yet")
			return fmt.Errorf("app URL not available yet")
		}

		appURL = liveURL
		log.WithField("url", appURL).Info("App deployment completed successfully")
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("app deployment did not complete within timeout: %w", err)
	}

	return appURL, nil
}

func (a *AppService) DeleteApp(ctx context.Context, appName string, cfg *config.Config) error {
	apps, _, err := a.client.Apps.List(ctx, &godo.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list apps: %w", err)
	}

	var targetApp *godo.App
	for _, app := range apps {
		if app.Spec.Name == fmt.Sprintf("%s-%s", cfg.Application.NamePrefix, appName) {
			targetApp = app
			break
		}
	}

	if targetApp == nil {
		return nil
	}

	if _, err := a.client.Apps.Delete(ctx, targetApp.ID); err != nil {
		return fmt.Errorf("failed to delete DigitalOcean app: %w", err)
	}

	return nil
}
