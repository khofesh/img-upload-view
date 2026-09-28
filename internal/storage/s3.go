package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/rs/zerolog"
)

type Options struct {
	Endpoint       string
	PublicEndpoint string
	Region         string
	Bucket         string
	AccessKey      string
	SecretKey      string
	UsePathStyle   bool
	PresignPutTTL  time.Duration
	PresignGetTTL  time.Duration
	CORSOrigins    []string
}

type S3Store struct {
	bucket        string
	client        *s3.Client
	presignClient *s3.PresignClient
	presignPutTTL time.Duration
	presignGetTTL time.Duration
	corsOrigins   []string
	logger        *zerolog.Logger
}

func NewS3Store(opts Options, logger *zerolog.Logger) *S3Store {
	newClient := func(endpoint string) *s3.Client {
		return s3.New(s3.Options{
			Region:       opts.Region,
			BaseEndpoint: aws.String(endpoint),
			UsePathStyle: opts.UsePathStyle,
			Credentials:  credentials.NewStaticCredentialsProvider(opts.AccessKey, opts.SecretKey, ""),
		})
	}

	publicEndpoint := opts.PublicEndpoint
	if publicEndpoint == "" {
		publicEndpoint = opts.Endpoint
	}

	return &S3Store{
		bucket:        opts.Bucket,
		client:        newClient(opts.Endpoint),
		presignClient: s3.NewPresignClient(newClient(publicEndpoint)),
		presignPutTTL: opts.PresignPutTTL,
		presignGetTTL: opts.PresignGetTTL,
		corsOrigins:   opts.CORSOrigins,
		logger:        logger,
	}
}

func (s *S3Store) PresignPut(ctx context.Context, key, contentType string, size int64) (PresignedRequest, error) {
	req, err := s.presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(size),
	}, s3.WithPresignExpires(s.presignPutTTL))
	if err != nil {
		return PresignedRequest{}, fmt.Errorf("unable to presign put object: %v", err)
	}

	return PresignedRequest{
		URL:       req.URL,
		Method:    req.Method,
		Headers:   map[string]string{"Content-Type": contentType},
		ExpiresAt: time.Now().Add(s.presignPutTTL),
	}, nil
}

func (s *S3Store) PresignGet(ctx context.Context, key string) (string, error) {
	req, err := s.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(s.presignGetTTL))
	if err != nil {
		return "", fmt.Errorf("unable to presign get object: %v", err)
	}

	return req.URL, nil
}

func (s *S3Store) Head(ctx context.Context, key string) (ObjectInfo, error) {
	out, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if isNotFound(err) {
			return ObjectInfo{}, ErrNotFound
		}
		return ObjectInfo{}, fmt.Errorf("unable to head object: %v", err)
	}

	return ObjectInfo{
		Size:        aws.ToInt64(out.ContentLength),
		ContentType: aws.ToString(out.ContentType),
	}, nil
}

func (s *S3Store) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil && !isNotFound(err) {
		return fmt.Errorf("unable to delete object: %v", err)
	}

	return nil
}

func (s *S3Store) EnsureBucket(ctx context.Context) error {
	_, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(s.bucket)})
	if err == nil {
		return s.ensureCORS(ctx)
	}
	if !isNotFound(err) {
		return fmt.Errorf("unable to head bucket: %v", err)
	}

	_, err = s.client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(s.bucket)})
	if err != nil {
		return fmt.Errorf("unable to create bucket: %v", err)
	}

	s.logger.Info().Str("bucket", s.bucket).Msg("created object storage bucket")
	return s.ensureCORS(ctx)
}

func (s *S3Store) ensureCORS(ctx context.Context) error {
	if len(s.corsOrigins) == 0 {
		return nil
	}

	rules := make([]types.CORSRule, 0, 1)
	rules = append(rules, types.CORSRule{
		AllowedOrigins: s.corsOrigins,
		AllowedMethods: []string{"GET", "PUT"},
		AllowedHeaders: []string{"*"},
		MaxAgeSeconds:  aws.Int32(3000),
	})

	_, err := s.client.PutBucketCors(ctx, &s3.PutBucketCorsInput{
		Bucket: aws.String(s.bucket),
		CORSConfiguration: &types.CORSConfiguration{
			CORSRules: rules,
		},
	})
	if err != nil {
		// MinIO configures CORS server-side and returns NotImplemented here.
		s.logger.Warn().Err(err).Msg("unable to set bucket CORS (continuing)")
	}

	return nil
}

func isNotFound(err error) bool {
	var notFound *types.NotFound
	if errors.As(err, &notFound) {
		return true
	}

	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "NotFound", "NoSuchKey", "NoSuchBucket":
			return true
		}
	}

	return false
}
