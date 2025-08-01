package resources

import (
	"github.com/CaioDGallo/easy-cli/internal/config"
	"github.com/CaioDGallo/easy-cli/internal/types"
)

func GetDeploymentDefaults(cfg *config.Config) types.DeploymentDefaults {
	return types.DeploymentDefaults{
		JWT: types.JWTDefaults{
			Issuer:            "http://localhost",
			Audience:          "http://localhost",
			Key:               "SomeLongJWTKeyHere",
			ExpirationMinutes: "20",
		},
		Frontend: types.FrontendDefaults{
			DefaultLanguage:     "en",
			ValidationTime:      "907200",
			AmazonEnv:           "prod",
			RevalidationToken:   "GP7goH9yvlNh2419midDws16hANf6Mvz",
			ForcedLoginTimeInMs: "5000",
			BuildCommand:        "sh vercel-script.sh && npm run codegen && npm run build",
			InstallCommand:      "npm install",
			FrameworkPreset:     "nextjs",
		},
		S3Path: "public",
		GitRepository: types.GitRepositoryDefaults{
			BackendRepo:  cfg.Repository.Backend,
			FrontendRepo: cfg.Repository.Frontend,
			Branch:       "master",
		},
	}
}
