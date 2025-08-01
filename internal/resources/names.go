package resources

import (
	"fmt"
	"strings"

	"github.com/CaioDGallo/easy-cli/internal/config"
	"github.com/CaioDGallo/easy-cli/internal/types"
)

func GenerateResourceNames(sanitizedClientName string, cfg *config.Config) types.ResourceNames {
	return types.ResourceNames{
		S3Bucket:         generateS3BucketName(sanitizedClientName, cfg),
		DatabaseMain:     sanitizedClientName,
		DatabaseHangfire: fmt.Sprintf("%s-hf", sanitizedClientName),
		DOApp:            generateDOAppName(sanitizedClientName, cfg),
		VercelProject:    sanitizedClientName,
		FrontendURL:      generateFrontendURL(sanitizedClientName, cfg),
		BackendURL:       "",
	}
}

func generateS3BucketName(sanitizedClientName string, cfg *config.Config) string {
	return fmt.Sprintf("%s-%s", cfg.Application.NamePrefix, sanitizedClientName)
}

func generateDOAppName(sanitizedClientName string, cfg *config.Config) string {
	return fmt.Sprintf("%s-%s", cfg.Application.NamePrefix, sanitizedClientName)
}

func generateFrontendURL(sanitizedClientName string, cfg *config.Config) string {
	return fmt.Sprintf("https://%s-%s.vercel.app", sanitizedClientName, cfg.Application.NamePrefix)
}

func ValidateResourceNames(names types.ResourceNames) error {
	if err := validateS3BucketName(names.S3Bucket); err != nil {
		return fmt.Errorf("invalid S3 bucket name: %w", err)
	}

	if err := validateDatabaseName(names.DatabaseMain); err != nil {
		return fmt.Errorf("invalid main database name: %w", err)
	}

	if err := validateDatabaseName(names.DatabaseHangfire); err != nil {
		return fmt.Errorf("invalid hangfire database name: %w", err)
	}

	return nil
}

func validateS3BucketName(bucketName string) error {
	if len(bucketName) < 3 || len(bucketName) > 63 {
		return fmt.Errorf("bucket name must be between 3 and 63 characters")
	}

	if strings.Contains(bucketName, "_") {
		return fmt.Errorf("bucket name cannot contain underscores")
	}

	if strings.Contains(bucketName, " ") {
		return fmt.Errorf("bucket name cannot contain spaces")
	}

	if bucketName != strings.ToLower(bucketName) {
		return fmt.Errorf("bucket name must be lowercase")
	}

	return nil
}

func validateDatabaseName(dbName string) error {
	if len(dbName) == 0 {
		return fmt.Errorf("database name cannot be empty")
	}

	if len(dbName) > 63 {
		return fmt.Errorf("database name cannot exceed 63 characters")
	}

	return nil
}
