package types

type CreateVercelProjectBody struct {
	Name                 string              `json:"name"`
	BuildCommand         string              `json:"buildCommand"`
	InstallCommand       string              `json:"installCommand"`
	FrameworkPreset      string              `json:"framework"`
	GitRepository        VercelGitRepo       `json:"gitRepository"`
	EnvironmentVariables []VercelEnvVariable `json:"environmentVariables"`
}

type VercelEnvVariable struct {
	Key       string `json:"key"`
	Target    string `json:"target"`
	Value     string `json:"value"`
	Type      string `json:"type"`
	GitBranch string `json:"gitBranch,omitempty"`
}

type VercelGitRepo struct {
	Repo string `json:"repo"`
	Type string `json:"type"`
}

type VercelDeploymentRequest struct {
	GitSource VercelGitSource `json:"gitSource"`
	Project   string          `json:"project"`
	Name      string          `json:"name,omitempty"`
	Target    string          `json:"target,omitempty"`
}

type VercelGitSource struct {
	Type     string `json:"type"`
	Repo     string `json:"repo"`
	RepoUuid string `json:"repoUuid"`
	Ref      string `json:"ref"`
}
