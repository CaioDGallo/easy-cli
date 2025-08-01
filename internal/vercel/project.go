package vercel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/CaioDGallo/easy-cli/internal/config"
	"github.com/CaioDGallo/easy-cli/internal/interfaces"
	"github.com/CaioDGallo/easy-cli/internal/logger"
	"github.com/CaioDGallo/easy-cli/internal/resources"
	"github.com/CaioDGallo/easy-cli/internal/types"
	"github.com/sirupsen/logrus"
)

var _ interfaces.StaticHostingProvider = (*ProjectService)(nil)

type ProjectService struct {
	client *http.Client
	config config.VercelConfig
}

func NewProjectService(cfg config.VercelConfig) *ProjectService {
	return &ProjectService{
		client: &http.Client{},
		config: cfg,
	}
}

func (p *ProjectService) CreateProject(ctx context.Context, sanitizedClientName string, envVars []types.VercelEnvVariable, cfg *config.Config) (string, error) {
	defaults := resources.GetDeploymentDefaults(cfg)

	createProjectBody := &types.CreateVercelProjectBody{
		Name:            sanitizedClientName,
		BuildCommand:    defaults.Frontend.BuildCommand,
		InstallCommand:  defaults.Frontend.InstallCommand,
		FrameworkPreset: defaults.Frontend.FrameworkPreset,
		GitRepository: types.VercelGitRepo{
			Repo: defaults.GitRepository.FrontendRepo,
			Type: "bitbucket",
		},
		EnvironmentVariables: envVars,
	}

	jsonData, err := json.Marshal(createProjectBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := p.generateRequest(ctx, "POST", "/v11/projects", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to generate request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to create Vercel project: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Vercel API error (status %d): %s", resp.StatusCode, string(body))
	}

	domain, err := p.GetProjectDomain(ctx, sanitizedClientName)
	if err != nil {
		return "", fmt.Errorf("failed to get project domain: %w", err)
	}

	return domain, nil
}

func (p *ProjectService) generateRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	vercelAPIURL := fmt.Sprintf("https://api.vercel.com%s?teamId=%s", path, p.config.TeamID)

	req, err := http.NewRequestWithContext(ctx, method, vercelAPIURL, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+p.config.Token)
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func (p *ProjectService) DeleteProject(ctx context.Context, projectName string) error {
	log := logger.WithFields(logrus.Fields{
		"project": projectName,
		"service": "vercel",
		"action":  "delete",
	})

	log.Info("Starting Vercel project deletion")

	req, err := p.generateRequest(ctx, "DELETE", fmt.Sprintf("/v9/projects/%s", projectName), nil)
	if err != nil {
		log.WithError(err).Error("Failed to generate delete request")
		return fmt.Errorf("failed to generate delete request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		log.WithError(err).Error("Failed to delete Vercel project")
		return fmt.Errorf("failed to delete Vercel project: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		log.Info("Project does not exist, skipping deletion")
		return nil
	}

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		log.WithField("status", resp.StatusCode).WithField("response", string(body)).Error("Vercel API error during deletion")
		return fmt.Errorf("Vercel API error during deletion (status %d): %s", resp.StatusCode, string(body))
	}

	log.Info("Vercel project deleted successfully")
	return nil
}

func (p *ProjectService) CreateDeployment(ctx context.Context, client types.Client, cfg *config.Config) error {
	log := logger.WithFields(logrus.Fields{
		"project": client.SanitizedClientName,
		"service": "vercel",
		"action":  "deploy",
	})

	log.Info("Starting Vercel deployment creation")

	defaults := resources.GetDeploymentDefaults(cfg)

	deploymentRequest := &types.VercelDeploymentRequest{
		GitSource: types.VercelGitSource{
			Type:     "bitbucket",
			Repo:     defaults.GitRepository.FrontendRepo,
			RepoUuid: p.config.FrontendRepoUuid,
			Ref:      client.FrontendBranch,
		},
		Project: client.SanitizedClientName,
		Name:    fmt.Sprintf("%s-deployment", client.SanitizedClientName),
		Target:  "production",
	}

	jsonData, err := json.Marshal(deploymentRequest)
	if err != nil {
		log.WithError(err).Error("Failed to marshal deployment request")
		return fmt.Errorf("failed to marshal deployment request: %w", err)
	}

	req, err := p.generateRequest(ctx, "POST", "/v13/deployments", bytes.NewBuffer(jsonData))
	if err != nil {
		log.WithError(err).Error("Failed to generate deployment request")
		return fmt.Errorf("failed to generate deployment request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		log.WithError(err).Error("Failed to create Vercel deployment")
		return fmt.Errorf("failed to create Vercel deployment: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		log.WithField("status", resp.StatusCode).WithField("response", string(body)).Error("Vercel API error during deployment")
		return fmt.Errorf("Vercel API error during deployment (status %d): %s", resp.StatusCode, string(body))
	}

	log.Info("Vercel deployment created successfully")
	return nil
}

func (p *ProjectService) GetProjectDomain(ctx context.Context, projectName string) (string, error) {
	req, err := p.generateRequest(ctx, "GET", fmt.Sprintf("/v9/projects/%s/domains", projectName), nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get Vercel project domains: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Vercel API error (status %d): %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var domainsResponse struct {
		Domains []struct {
			Name               string `json:"name"`
			ApexName           string `json:"apexName"`
			ProjectID          string `json:"projectId"`
			Redirect           string `json:"redirect,omitempty"`
			RedirectStatusCode int    `json:"redirectStatusCode,omitempty"`
		} `json:"domains"`
	}

	if err := json.Unmarshal(body, &domainsResponse); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(domainsResponse.Domains) == 0 {
		return "", fmt.Errorf("no domains found for project")
	}

	var primaryDomain string
	for _, domain := range domainsResponse.Domains {
		if domain.Redirect == "" {
			primaryDomain = domain.Name
			break
		}
	}

	if primaryDomain == "" {
		primaryDomain = domainsResponse.Domains[0].Name
	}

	return fmt.Sprintf("https://%s", primaryDomain), nil
}

func (p *ProjectService) UpdateProjectEnvironmentVariables(ctx context.Context, projectName string, envVars []types.VercelEnvVariable) error {
	req, err := p.generateRequest(ctx, "GET", fmt.Sprintf("/v10/projects/%s/env", projectName), nil)
	if err != nil {
		return fmt.Errorf("failed to generate request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get Vercel project env vars: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Vercel API error getting env vars (status %d): %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	var existingEnvVars struct {
		Envs []struct {
			ID     string   `json:"id"`
			Key    string   `json:"key"`
			Value  string   `json:"value"`
			Target []string `json:"target"`
			Type   string   `json:"type"`
		} `json:"envs"`
	}

	if err := json.Unmarshal(body, &existingEnvVars); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	for _, envVar := range envVars {
		var existingID string
		for _, existing := range existingEnvVars.Envs {
			if existing.Key == envVar.Key {
				existingID = existing.ID
				break
			}
		}

		if existingID != "" {
			updateData := map[string]interface{}{
				"value":  envVar.Value,
				"target": []string{envVar.Target},
				"type":   envVar.Type,
			}

			jsonData, err := json.Marshal(updateData)
			if err != nil {
				return fmt.Errorf("failed to marshal env var update: %w", err)
			}

			req, err := p.generateRequest(ctx, "PATCH", fmt.Sprintf("/v9/projects/%s/env/%s", projectName, existingID), bytes.NewBuffer(jsonData))
			if err != nil {
				return fmt.Errorf("failed to generate update request: %w", err)
			}

			resp, err := p.client.Do(req)
			if err != nil {
				return fmt.Errorf("failed to update env var: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode >= 400 {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("Vercel API error updating env var (status %d): %s", resp.StatusCode, string(body))
			}
		} else {
			jsonData, err := json.Marshal(envVar)
			if err != nil {
				return fmt.Errorf("failed to marshal env var: %w", err)
			}

			req, err := p.generateRequest(ctx, "POST", fmt.Sprintf("/v10/projects/%s/env?upsert=true", projectName), bytes.NewBuffer(jsonData))
			if err != nil {
				return fmt.Errorf("failed to generate create request: %w", err)
			}

			resp, err := p.client.Do(req)
			if err != nil {
				return fmt.Errorf("failed to create env var: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode >= 400 {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("Vercel API error creating env var (status %d): %s", resp.StatusCode, string(body))
			}
		}
	}

	return nil
}
