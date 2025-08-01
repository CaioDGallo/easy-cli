package interfaces

import (
	"context"

	"github.com/CaioDGallo/easy-cli/internal/config"
	"github.com/CaioDGallo/easy-cli/internal/types"
)

type CloudStorageProvider interface {
	CreateBucket(ctx context.Context, bucketName string) error
	DeleteBucket(ctx context.Context, bucketName string) error
}

type DatabaseProvider interface {
	CreateClientDatabase(sanitizedClientName string) error
	DeleteClientDatabases(mainDBName, hangfireDBName string) error
}

type AppHostingProvider interface {
	CreateApp(ctx context.Context, client types.Client, envVars types.DigitalOceanEnvVars, cfg *config.Config) (string, error)
	UpdateAppEnvironmentVariables(ctx context.Context, appName string, envVars types.DigitalOceanEnvVars, cfg *config.Config) error
	DeleteApp(ctx context.Context, appName string, cfg *config.Config) error
}

type StaticHostingProvider interface {
	CreateProject(ctx context.Context, sanitizedClientName string, envVars []types.VercelEnvVariable, cfg *config.Config) (string, error)
	DeleteProject(ctx context.Context, projectName string) error
	CreateDeployment(ctx context.Context, client types.Client, cfg *config.Config) error
	UpdateProjectEnvironmentVariables(ctx context.Context, projectName string, envVars []types.VercelEnvVariable) error
}
