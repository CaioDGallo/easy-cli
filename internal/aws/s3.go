package aws

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/CaioDGallo/easy-cli/internal/interfaces"
	"github.com/CaioDGallo/easy-cli/internal/logger"
	"github.com/CaioDGallo/easy-cli/internal/retry"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/sirupsen/logrus"
)

var _ interfaces.CloudStorageProvider = (*S3Service)(nil)

type S3Service struct {
	client *s3.Client
}

func NewS3Service(region, accessKeyID, secretAccessKey string) (*S3Service, error) {
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKeyID,
			secretAccessKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &S3Service{
		client: s3.NewFromConfig(cfg),
	}, nil
}

func (s *S3Service) CreateBucket(ctx context.Context, bucketName string) error {
	log := logger.WithFields(logrus.Fields{
		"bucket":  bucketName,
		"service": "s3",
	})

	log.Info("Creating S3 bucket")

	retryConfig := retry.DefaultConfig()

	err := retry.Do(ctx, retryConfig, func() error {
		createBucketInput := s3.CreateBucketInput{
			Bucket: aws.String(bucketName),
		}

		_, err := s.client.CreateBucket(ctx, &createBucketInput)
		return err
	})
	if err != nil {
		log.WithError(err).Error("Failed to create S3 bucket")
		return fmt.Errorf("failed to create bucket %s: %w", bucketName, err)
	}

	log.Info("Configuring bucket encryption")
	if err := s.configureBucketEncryption(ctx, bucketName); err != nil {
		log.WithError(err).Error("Failed to configure bucket encryption")
		return fmt.Errorf("failed to configure bucket encryption: %w", err)
	}

	log.Info("Configuring bucket public access")
	if err := s.configureBucketPublicAccess(ctx, bucketName); err != nil {
		log.WithError(err).Error("Failed to configure bucket public access")
		return fmt.Errorf("failed to configure bucket public access: %w", err)
	}

	log.Info("Creating public folder")
	if err := s.createPublicFolder(ctx, bucketName); err != nil {
		log.WithError(err).Error("Failed to create public folder")
		return fmt.Errorf("failed to create public folder: %w", err)
	}

	log.Info("S3 bucket created successfully")
	return nil
}

func (s *S3Service) configureBucketEncryption(ctx context.Context, bucketName string) error {
	encryptionInput := &s3.PutBucketEncryptionInput{
		Bucket: aws.String(bucketName),
		ServerSideEncryptionConfiguration: &s3types.ServerSideEncryptionConfiguration{
			Rules: []s3types.ServerSideEncryptionRule{
				{
					ApplyServerSideEncryptionByDefault: &s3types.ServerSideEncryptionByDefault{
						SSEAlgorithm: s3types.ServerSideEncryptionAes256,
					},
					BucketKeyEnabled: aws.Bool(false),
				},
			},
		},
	}

	_, err := s.client.PutBucketEncryption(ctx, encryptionInput)
	return err
}

func (s *S3Service) configureBucketPublicAccess(ctx context.Context, bucketName string) error {
	publicAccessInput := &s3.PutPublicAccessBlockInput{
		Bucket: aws.String(bucketName),
		PublicAccessBlockConfiguration: &s3types.PublicAccessBlockConfiguration{
			BlockPublicAcls:       aws.Bool(false),
			IgnorePublicAcls:      aws.Bool(false),
			BlockPublicPolicy:     aws.Bool(false),
			RestrictPublicBuckets: aws.Bool(false),
		},
	}

	if _, err := s.client.PutPublicAccessBlock(ctx, publicAccessInput); err != nil {
		return err
	}

	bucketPolicy := fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Sid": "PublicRead",
				"Effect": "Allow",
				"Principal": "*",
				"Action": [
					"s3:GetObject",
					"s3:GetObjectVersion"
				],
				"Resource": "arn:aws:s3:::%s/public/*"
			}
		]
	}`, bucketName)

	policyInput := &s3.PutBucketPolicyInput{
		Bucket: aws.String(bucketName),
		Policy: aws.String(bucketPolicy),
	}

	_, err := s.client.PutBucketPolicy(ctx, policyInput)
	return err
}

func (s *S3Service) createPublicFolder(ctx context.Context, bucketName string) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String("public/"),
		Body:   strings.NewReader(""),
	})
	return err
}

func (s *S3Service) DeleteBucket(ctx context.Context, bucketName string) error {
	log := logger.WithFields(logrus.Fields{
		"bucket":  bucketName,
		"service": "s3",
		"action":  "delete",
	})

	log.Info("Starting S3 bucket deletion")

	if err := s.emptyBucket(ctx, bucketName); err != nil {
		if s.isBucketNotFoundError(err) {
			log.Info("Bucket does not exist, skipping deletion")
			return nil
		}
		log.WithError(err).Error("Failed to empty bucket")
		return fmt.Errorf("failed to empty bucket %s: %w", bucketName, err)
	}

	retryConfig := retry.DefaultConfig()
	err := retry.Do(ctx, retryConfig, func() error {
		_, err := s.client.DeleteBucket(ctx, &s3.DeleteBucketInput{
			Bucket: aws.String(bucketName),
		})
		return err
	})
	if err != nil {
		if s.isBucketNotFoundError(err) {
			log.Info("Bucket does not exist, deletion successful")
			return nil
		}
		log.WithError(err).Error("Failed to delete S3 bucket")
		return fmt.Errorf("failed to delete bucket %s: %w", bucketName, err)
	}

	log.Info("S3 bucket deleted successfully")
	return nil
}

func (s *S3Service) emptyBucket(ctx context.Context, bucketName string) error {
	log := logger.WithFields(logrus.Fields{
		"bucket":  bucketName,
		"service": "s3",
		"action":  "empty",
	})

	log.Info("Emptying S3 bucket")

	listInput := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	}

	for {
		listOutput, err := s.client.ListObjectsV2(ctx, listInput)
		if err != nil {
			return fmt.Errorf("failed to list objects in bucket: %w", err)
		}

		if len(listOutput.Contents) == 0 {
			break
		}

		objectIds := make([]s3types.ObjectIdentifier, len(listOutput.Contents))
		for i, obj := range listOutput.Contents {
			objectIds[i] = s3types.ObjectIdentifier{
				Key: obj.Key,
			}
		}

		deleteInput := &s3.DeleteObjectsInput{
			Bucket: aws.String(bucketName),
			Delete: &s3types.Delete{
				Objects: objectIds,
			},
		}

		_, err = s.client.DeleteObjects(ctx, deleteInput)
		if err != nil {
			return fmt.Errorf("failed to delete objects from bucket: %w", err)
		}

		log.Infof("Deleted %d objects from bucket", len(objectIds))

		if listOutput.IsTruncated == nil || !*listOutput.IsTruncated {
			break
		}
		listInput.ContinuationToken = listOutput.NextContinuationToken
	}

	log.Info("S3 bucket emptied successfully")
	return nil
}

func (s *S3Service) isBucketNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	var noBucket *s3types.NoSuchBucket
	if errors.As(err, &noBucket) {
		return true
	}

	var notFound *s3types.NotFound
	if errors.As(err, &notFound) {
		return true
	}

	return false
}
