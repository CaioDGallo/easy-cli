package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type Config struct {
	Database    DatabaseConfig
	Vercel      VercelConfig
	AWS         AWSConfig
	DO          DOConfig
	SMTP        SMTPConfig
	Repository  RepositoryConfig
	Application ApplicationConfig
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

type VercelConfig struct {
	Token            string
	TeamID           string
	FrontendRepoUuid string
}

type AWSConfig struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
}

type DOConfig struct {
	Token string
}

type SMTPConfig struct {
	Server           string
	Username         string
	DoNotReplyEmail  string
	DevEmail         string
}

type RepositoryConfig struct {
	Backend  string
	Frontend string
}

type ApplicationConfig struct {
	NamePrefix string
}

func Load() (*Config, error) {
	envFile := findEnvFile()
	if envFile == "" {
		return nil, fmt.Errorf("no environment file found. Please create one at ~/.easy-cli.env or run the installer to create a template")
	}

	if err := godotenv.Load(envFile); err != nil {
		return nil, fmt.Errorf("error loading .env file from %s: %w", envFile, err)
	}

	config := &Config{
		Database: DatabaseConfig{
			Host:     getEnvOrDefault("DB_HOST", "your-database-host.rds.amazonaws.com"),
			Port:     5432,
			User:     getEnvOrDefault("DB_USER", "postgres"),
			Password: os.Getenv("DB_PASSWORD"),
			DBName:   getEnvOrDefault("DB_NAME", "postgres"),
		},
		Vercel: VercelConfig{
			Token:            os.Getenv("VERCEL_TOKEN"),
			TeamID:           os.Getenv("VERCEL_TEAM_ID"),
			FrontendRepoUuid: os.Getenv("VERCEL_FRONTEND_REPO_UUID"),
		},
		AWS: AWSConfig{
			Region:          getEnvOrDefault("AWS_REGION", "us-east-1"),
			AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
			SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		},
		DO: DOConfig{
			Token: os.Getenv("DO_TOKEN"),
		},
		SMTP: SMTPConfig{
			Server:          getEnvOrDefault("SMTP_SERVER", "your-smtp-server.com"),
			Username:        getEnvOrDefault("SMTP_USERNAME", "your-smtp-username@yourdomain.com"),
			DoNotReplyEmail: getEnvOrDefault("SMTP_DO_NOT_REPLY_EMAIL", "noreply@yourdomain.com"),
			DevEmail:        getEnvOrDefault("SMTP_DEV_EMAIL", "developer@yourdomain.com"),
		},
		Repository: RepositoryConfig{
			Backend:  getEnvOrDefault("BACKEND_REPO", "your-org/your-backend-repo"),
			Frontend: getEnvOrDefault("FRONTEND_REPO", "your-org/your-frontend-repo"),
		},
		Application: ApplicationConfig{
			NamePrefix: getEnvOrDefault("APP_NAME_PREFIX", "your-app-prefix"),
		},
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

func (c *Config) Validate() error {
	if c.Database.Password == "" {
		return fmt.Errorf("DB_PASSWORD environment variable is required")
	}
	if c.Vercel.Token == "" {
		return fmt.Errorf("VERCEL_TOKEN environment variable is required")
	}
	if c.Vercel.TeamID == "" {
		return fmt.Errorf("VERCEL_TEAM_ID environment variable is required")
	}
	if c.Vercel.FrontendRepoUuid == "" {
		return fmt.Errorf("VERCEL_FRONTEND_REPO_UUID environment variable is required")
	}
	if c.DO.Token == "" {
		return fmt.Errorf("DO_TOKEN environment variable is required")
	}
	if c.AWS.AccessKeyID == "" {
		return fmt.Errorf("AWS_ACCESS_KEY_ID environment variable is required")
	}
	if c.AWS.SecretAccessKey == "" {
		return fmt.Errorf("AWS_SECRET_ACCESS_KEY environment variable is required")
	}
	// Note: SMTP, Repository, and Application configs are optional and have defaults
	return nil
}

func findEnvFile() string {
	// Priority order for .env file locations:
	// 1. ~/.easy-cli.env (user's home directory)
	// 2. Same directory as the binary
	// 3. Current working directory (.env)

	locations := []string{}

	// 1. User home directory
	if homeDir, err := os.UserHomeDir(); err == nil {
		locations = append(locations, filepath.Join(homeDir, ".easy-cli.env"))
	}

	// 2. Same directory as binary
	if execPath, err := os.Executable(); err == nil {
		binaryDir := filepath.Dir(execPath)
		locations = append(locations, filepath.Join(binaryDir, ".env"))
	}

	// 3. Current working directory
	if cwd, err := os.Getwd(); err == nil {
		locations = append(locations, filepath.Join(cwd, ".env"))
	}

	// Check each location and return the first one that exists
	for _, location := range locations {
		if _, err := os.Stat(location); err == nil {
			return location
		}
	}

	return "" // No .env file found
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
