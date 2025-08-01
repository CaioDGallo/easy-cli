package cmd

import (
	"context"
	"fmt"

	"github.com/CaioDGallo/easy-cli/internal/aws"
	"github.com/CaioDGallo/easy-cli/internal/config"
	"github.com/CaioDGallo/easy-cli/internal/database"
	"github.com/CaioDGallo/easy-cli/internal/digitalocean"
	"github.com/CaioDGallo/easy-cli/internal/envvars"
	"github.com/CaioDGallo/easy-cli/internal/logger"
	"github.com/CaioDGallo/easy-cli/internal/rollback"
	"github.com/CaioDGallo/easy-cli/internal/types"
	"github.com/CaioDGallo/easy-cli/internal/utils"
	"github.com/CaioDGallo/easy-cli/internal/validation"
	"github.com/CaioDGallo/easy-cli/internal/vercel"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var freshInstallCmd = &cobra.Command{
	Use:   "fresh-install",
	Short: "Command used to install a new project",
	Long:  `This command is used to set up a new project from scratch, specifying client parameters and configs.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			logger.Fatalf("Failed to load configuration: %v", err)
		}

		clientName := cmd.Flag("client-name").Value.String()
		smtpServer := cmd.Flag("smtp-server").Value.String()
		smtpPort := cmd.Flag("smtp-port").Value.String()
		smtpUsername := cmd.Flag("smtp-username").Value.String()
		smtpPassword := cmd.Flag("smtp-password").Value.String()
		smtpDoNotReplyName := cmd.Flag("smtp-donotreplyname").Value.String()
		smtpDoNotReplyEmail := cmd.Flag("smtp-donotreplyemail").Value.String()
		smtpDevEmail := cmd.Flag("smtp-devemail").Value.String()
		backendBranch := cmd.Flag("backend-branch").Value.String()
		frontendBranch := cmd.Flag("frontend-branch").Value.String()

		log := logger.WithFields(logrus.Fields{
			"client":  clientName,
			"command": "fresh-install",
		})

		client := types.Client{
			Name:                clientName,
			SanitizedClientName: utils.SanitizeClientName(clientName),
			DatabaseHost:        cfg.Database.Host,
			DatabaseUser:        cfg.Database.User,
			BackendBranch:       backendBranch,
			FrontendBranch:      frontendBranch,
			SMTPInfo: types.SMTPInfo{
				Server:          smtpServer,
				Username:        smtpUsername,
				Port:            smtpPort,
				Password:        smtpPassword,
				DoNotReplyName:  smtpDoNotReplyName,
				DoNotReplyEmail: smtpDoNotReplyEmail,
				DevEmail:        smtpDevEmail,
			},
			BackendInfo: types.BackendInfo{
				URL:              "",
				DatabasePassword: cfg.Database.Password,
			},
			FrontendInfo: types.FrontendInfo{
				URL: "",
			},
		}

		log.Info("Validating client configuration")
		if err := validation.ValidateClient(client); err != nil {
			log.WithError(err).Error("Client validation failed")
			logger.Fatalf("Client validation failed: %v", err)
		}

		log.Info("Starting fresh install process")
		if err := freshInstall(client, cfg); err != nil {
			log.WithError(err).Error("Fresh install failed")
			logger.Fatalf("Fresh install failed: %v", err)
		}
		log.Info("Fresh install completed successfully")
	},
}

func init() {
	rootCmd.AddCommand(freshInstallCmd)

	freshInstallCmd.Flags().StringP("client-name", "c", "", "The name of the client for this setup")
	freshInstallCmd.MarkFlagRequired("client-name")

	// Load config to get SMTP defaults
	cfg, _ := config.Load()
	smtpServer := "your-smtp-server.com"
	smtpUsername := "your-smtp-username@yourdomain.com"
	smtpDoNotReplyEmail := "noreply@yourdomain.com"
	smtpDevEmail := "developer@yourdomain.com"
	
	if cfg != nil {
		smtpServer = cfg.SMTP.Server
		smtpUsername = cfg.SMTP.Username
		smtpDoNotReplyEmail = cfg.SMTP.DoNotReplyEmail
		smtpDevEmail = cfg.SMTP.DevEmail
	}

	freshInstallCmd.Flags().StringP("smtp-server", "s", smtpServer, "The SMTP server for the client of this setup")
	freshInstallCmd.Flags().StringP("smtp-port", "P", "587", "The SMTP port for the client of this setup")
	freshInstallCmd.Flags().StringP("smtp-username", "u", smtpUsername, "The SMTP username for the client of this setup")
	freshInstallCmd.Flags().StringP("smtp-password", "p", "your-smtp-password", "The SMTP password for the client of this setup")
	freshInstallCmd.Flags().StringP("smtp-donotreplyname", "r", "Do Not Reply", "The SMTP DoNotReplyName for the client of this setup")
	freshInstallCmd.Flags().StringP("smtp-donotreplyemail", "m", smtpDoNotReplyEmail, "The SMTP DoNotReplyEmail for the client of this setup")
	freshInstallCmd.Flags().StringP("smtp-devemail", "e", smtpDevEmail, "The SMTP devemail for the client of this setup")
	
	freshInstallCmd.Flags().StringP("backend-branch", "b", "master", "The git branch to use for the backend deployment")
	freshInstallCmd.Flags().StringP("frontend-branch", "f", "master", "The git branch to use for the frontend deployment")
}

func freshInstall(client types.Client, cfg *config.Config) error {
	ctx := context.Background()

	log := logger.WithFields(logrus.Fields{
		"client":         client.Name,
		"sanitized_name": client.SanitizedClientName,
	})

	rollbackMgr := rollback.NewManager()

	log.Info("Generating deployment environment")
	deploymentEnv, err := envvars.GenerateDeploymentEnvironment(client, cfg)
	if err != nil {
		log.WithError(err).Error("Failed to generate deployment environment")
		return fmt.Errorf("failed to generate deployment environment: %w", err)
	}

	bucketName := deploymentEnv.ResourceNames.S3Bucket

	log.Info("Creating S3 service")
	s3Service, err := aws.NewS3Service(cfg.AWS.Region, cfg.AWS.AccessKeyID, cfg.AWS.SecretAccessKey)
	if err != nil {
		log.WithError(err).Error("Failed to create S3 service")
		return fmt.Errorf("failed to create S3 service: %w", err)
	}

	if err := s3Service.CreateBucket(ctx, bucketName); err != nil {
		log.WithError(err).Error("Failed to setup AWS S3")
		return fmt.Errorf("failed to setup AWS S3: %w", err)
	}
	rollbackMgr.AddAction("S3 bucket cleanup", func(ctx context.Context) error {
		log.Info("Rolling back S3 bucket creation")
		if err := s3Service.DeleteBucket(ctx, bucketName); err != nil {
			log.WithError(err).Error("Failed to rollback S3 bucket")
			return fmt.Errorf("failed to delete S3 bucket during rollback: %w", err)
		}
		log.Info("S3 bucket rollback completed")
		return nil
	})

	log.Info("Creating database service")
	dbService := database.NewPostgresService(cfg.Database)
	if err := dbService.CreateClientDatabases(deploymentEnv.ResourceNames.DatabaseMain, deploymentEnv.ResourceNames.DatabaseHangfire); err != nil {
		log.WithError(err).Error("Failed to setup database")
		if rollbackErr := rollbackMgr.ExecuteRollback(ctx); rollbackErr != nil {
			log.WithError(rollbackErr).Error("Rollback failed")
		}
		return fmt.Errorf("failed to setup database: %w", err)
	}
	rollbackMgr.AddAction("Database cleanup", func(ctx context.Context) error {
		log.Info("Rolling back database creation")
		if err := dbService.DeleteClientDatabases(deploymentEnv.ResourceNames.DatabaseMain, deploymentEnv.ResourceNames.DatabaseHangfire); err != nil {
			log.WithError(err).Error("Failed to rollback databases")
			return fmt.Errorf("failed to delete databases during rollback: %w", err)
		}
		log.Info("Database rollback completed")
		return nil
	})

	log.Info("Creating DigitalOcean service")
	doService := digitalocean.NewAppService(cfg.DO.Token)
	backendEnvVars := types.DigitalOceanEnvVars{
		AppEnvs:       deploymentEnv.Backend.AppLevelVars,
		ComponentEnvs: deploymentEnv.Backend.ComponentLevelVars,
	}
	backendURL, err := doService.CreateApp(ctx, client, backendEnvVars, cfg)
	if err != nil {
		log.WithError(err).Error("Failed to setup DigitalOcean")
		if rollbackErr := rollbackMgr.ExecuteRollback(ctx); rollbackErr != nil {
			log.WithError(rollbackErr).Error("Rollback failed")
		}
		return fmt.Errorf("failed to setup DigitalOcean: %w", err)
	}
	log.WithField("backend_url", backendURL).Info("DigitalOcean app created successfully")
	rollbackMgr.AddAction("DigitalOcean app cleanup", func(ctx context.Context) error {
		log.Info("Rolling back DigitalOcean app creation")
		if err := doService.DeleteApp(ctx, client.SanitizedClientName, cfg); err != nil {
			log.WithError(err).Error("Failed to rollback DigitalOcean app")
			return fmt.Errorf("failed to delete DigitalOcean app during rollback: %w", err)
		}
		log.Info("DigitalOcean app rollback completed")
		return nil
	})

	log.Info("Creating Vercel service")
	vercelService := vercel.NewProjectService(cfg.Vercel)

	log.Info("Updating client with backend URL")
	client.BackendInfo.URL = backendURL

	log.Info("Generating Vercel environment variables")
	frontendEnvVars, err := envvars.GenerateVercelEnvironmentVariables(client, cfg)
	if err != nil {
		log.WithError(err).Error("Failed to generate Vercel environment variables")
		if rollbackErr := rollbackMgr.ExecuteRollback(ctx); rollbackErr != nil {
			log.WithError(rollbackErr).Error("Rollback failed")
		}
		return fmt.Errorf("failed to generate Vercel environment variables: %w", err)
	}

	frontendURL, err := vercelService.CreateProject(ctx, client.SanitizedClientName, frontendEnvVars, cfg)
	if err != nil {
		log.WithError(err).Error("Failed to setup Vercel")
		if rollbackErr := rollbackMgr.ExecuteRollback(ctx); rollbackErr != nil {
			log.WithError(rollbackErr).Error("Rollback failed")
		}
		return fmt.Errorf("failed to setup Vercel: %w", err)
	}
	log.WithField("frontend_url", frontendURL).Info("Vercel project created successfully")
	rollbackMgr.AddAction("Vercel project cleanup", func(ctx context.Context) error {
		log.Info("Rolling back Vercel project creation")
		if err := vercelService.DeleteProject(ctx, client.SanitizedClientName); err != nil {
			log.WithError(err).Error("Failed to rollback Vercel project")
			return fmt.Errorf("failed to delete Vercel project during rollback: %w", err)
		}
		log.Info("Vercel project rollback completed")
		return nil
	})

	log.Info("Updating Vercel environment variables with actual frontend URL")
	client.FrontendInfo.URL = frontendURL

	updatedFrontendEnvVars, err := envvars.GenerateVercelEnvironmentVariables(client, cfg)
	if err != nil {
		log.WithError(err).Error("Failed to generate updated Vercel environment variables")
		return fmt.Errorf("failed to generate updated Vercel environment variables: %w", err)
	}

	if err := vercelService.UpdateProjectEnvironmentVariables(ctx, client.SanitizedClientName, updatedFrontendEnvVars); err != nil {
		log.WithError(err).Error("Failed to update Vercel environment variables")
		return fmt.Errorf("failed to update Vercel environment variables: %w", err)
	}

	log.Info("Creating initial Vercel deployment")
	if err := vercelService.CreateDeployment(ctx, client, cfg); err != nil {
		log.WithError(err).Error("Failed to create Vercel deployment")
		return fmt.Errorf("failed to create Vercel deployment: %w", err)
	}

	log.Info("Updating DigitalOcean app with frontend URL")

	updatedBackendEnv, err := envvars.GenerateDeploymentEnvironment(client, cfg)
	if err != nil {
		log.WithError(err).Error("Failed to generate updated deployment environment")
		return fmt.Errorf("failed to generate updated deployment environment: %w", err)
	}

	updatedBackendEnvVars := types.DigitalOceanEnvVars{
		AppEnvs:       updatedBackendEnv.Backend.AppLevelVars,
		ComponentEnvs: updatedBackendEnv.Backend.ComponentLevelVars,
	}

	if err := doService.UpdateAppEnvironmentVariables(ctx, client.SanitizedClientName, updatedBackendEnvVars, cfg); err != nil {
		log.WithError(err).Error("Failed to update DigitalOcean app environment variables")
		return fmt.Errorf("failed to update DigitalOcean app environment variables: %w", err)
	}

	log.WithFields(logrus.Fields{
		"backend_url":  backendURL,
		"frontend_url": frontendURL,
	}).Info("Fresh install completed successfully with bidirectional URL configuration")

	return nil
}
