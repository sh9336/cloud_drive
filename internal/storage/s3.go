// internal/storage/s3.go
package storage

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Service struct {
	client        *s3.Client
	presignClient *s3.Client // Client for presigned URLs with external endpoint
	bucketName    string
}

func NewS3Service(region, bucketName string, useIAMRole bool, accessKey, secretKey, endpoint, externalEndpoint string) (*S3Service, error) {
	var cfg aws.Config
	var err error

	if useIAMRole {
		// Production: Use IAM Role
		cfg, err = config.LoadDefaultConfig(context.Background(),
			config.WithRegion(region),
		)
	} else {
		// Development: Use access keys
		cfg, err = config.LoadDefaultConfig(context.Background(),
			config.WithRegion(region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				accessKey,
				secretKey,
				"",
			)),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create internal S3 client (for actual operations)
	var internalClient *s3.Client
	if endpoint != "" {
		// Use custom endpoint (MinIO for development)
		internalClient = s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(endpoint)
			o.UsePathStyle = true // Required for MinIO
		})
	} else {
		internalClient = s3.NewFromConfig(cfg)
	}

	// Create presigned URL client (for external endpoint)
	var presignedClient *s3.Client
	if externalEndpoint != "" && externalEndpoint != endpoint {
		// Different external endpoint - create separate client
		presignedClient = s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(externalEndpoint)
			o.UsePathStyle = true // Required for MinIO
		})
	} else {
		// Same endpoint or no external endpoint - use same client
		presignedClient = internalClient
	}

	return &S3Service{
		client:        internalClient,
		presignClient: presignedClient,
		bucketName:    bucketName,
	}, nil
}

// GeneratePresignedPutURL generates a presigned URL for uploading a file
func (s *S3Service) GeneratePresignedPutURL(ctx context.Context, key string, expiryDuration time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s.presignClient)

	request, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiryDuration
	})

	if err != nil {
		return "", fmt.Errorf("failed to generate presigned PUT URL: %w", err)
	}

	return request.URL, nil
}

// GeneratePresignedGetURL generates a presigned URL for downloading a file
func (s *S3Service) GeneratePresignedGetURL(ctx context.Context, key string, expiryDuration time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s.presignClient)

	request, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiryDuration
	})

	if err != nil {
		return "", fmt.Errorf("failed to generate presigned GET URL: %w", err)
	}

	return request.URL, nil
}

// DeleteObject deletes a file from S3
func (s *S3Service) DeleteObject(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})

	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	return nil
}

// HeadObject checks if an object exists
func (s *S3Service) HeadObject(ctx context.Context, key string) error {
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})

	return err
}
// PutObject uploads a file to S3
func (s *S3Service) PutObject(ctx context.Context, key string, body []byte, contentType string) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(body),
		ContentType: aws.String(contentType),
	})

	if err != nil {
		return fmt.Errorf("failed to put object: %w", err)
	}

	return nil
}

// GetBucketCors retrieves the CORS configuration for the bucket
func (s *S3Service) GetBucketCors(ctx context.Context) (*s3.GetBucketCorsOutput, error) {
	return s.client.GetBucketCors(ctx, &s3.GetBucketCorsInput{
		Bucket: aws.String(s.bucketName),
	})
}
