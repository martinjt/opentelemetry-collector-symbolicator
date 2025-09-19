package dsymprocessor

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"
)

var (
	errFailedToFindDSYM = fmt.Errorf("failed to find dSYM file")
)

type store struct {
	fetch  func(ctx context.Context, key string) ([]byte, error)
	logger *zap.Logger
	prefix string
}

func (s *store) GetDSYM(ctx context.Context, debugId, binaryName string) ([]byte, error) {
	path := filepath.Join(s.prefix, fmt.Sprintf("%s.dSYM", debugId), "Contents", "Resources", "DWARF", binaryName)
	dsymBytes, err := s.fetch(ctx, path)

	if err != nil {
		return nil, fmt.Errorf("%w: %s", errFailedToFindDSYM, path)
	}

	return dsymBytes, nil

}

func newFileStore(_ context.Context, logger *zap.Logger, cfg *LocalDSYMConfiguration) (*store, error) {
	if cfg == nil {
		return nil, fmt.Errorf("no file configuration provided")
	}

	return &store{
		fetch: func(ctx context.Context, key string) ([]byte, error) {
			return os.ReadFile(key)
		},
		logger: logger,
		prefix: cfg.Path,
	}, nil
}

func newS3Store(ctx context.Context, logger *zap.Logger, cfg *S3DSYMConfiguration) (*store, error) {
	if cfg == nil {
		return nil, fmt.Errorf("no S3 configuration provided")
	}

	options := make([]func(*config.LoadOptions) error, 0)

	if cfg.Region != "" {
		options = append(options, config.WithRegion(cfg.Region))
	}

	awsConfig, err := config.LoadDefaultConfig(ctx, options...)

	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(awsConfig)

	return &store{
		fetch: func(ctx context.Context, key string) ([]byte, error) {
			key = strings.TrimPrefix(key, "/")

			result, err := client.GetObject(ctx, &s3.GetObjectInput{
				Bucket: aws.String(cfg.BucketName),
				Key:    aws.String(key),
			})

			if err != nil {
				return nil, err
			}

			defer result.Body.Close()

			return io.ReadAll(result.Body)
		},
		logger: logger,
		prefix: cfg.Prefix,
	}, nil
}

func newGCSStore(ctx context.Context, logger *zap.Logger, cfg *GCSDSYMConfiguration) (*store, error) {
	if cfg == nil {
		return nil, fmt.Errorf("no GCS configuration provided")
	}

	client, err := storage.NewClient(ctx)

	if err != nil {
		return nil, err
	}

	bucket := client.Bucket(cfg.BucketName)

	return &store{
		fetch: func(ctx context.Context, key string) ([]byte, error) {
			// GCS keys can't start with a slash
			key = strings.TrimPrefix(key, "/")

			r, err := bucket.Object(key).NewReader(ctx)

			if err != nil {
				return nil, err
			}

			defer r.Close()

			return io.ReadAll(r)
		},
		logger: logger,
		prefix: cfg.Prefix,
	}, nil
}

func newAzureStore(ctx context.Context, logger *zap.Logger, cfg *AzureDSYMConfiguration) (*store, error) {
	if cfg == nil {
		return nil, fmt.Errorf("no Azure configuration provided")
	}

	// Determine the endpoint URL
	var url string
	if cfg.Endpoint != "" {
		url = cfg.Endpoint
	} else {
		url = fmt.Sprintf("https://%s.blob.core.windows.net/", cfg.AccountName)
	}

	// Create Azure Blob Storage client
	var client *azblob.Client
	var err error

	// Try connection string first (for Azurite), then fall back to credential-based auth
	if cfg.ConnectionString != "" {
		client, err = azblob.NewClientFromConnectionString(cfg.ConnectionString, nil)
	} else {
		// Create a default Azure credential
		credential, credErr := azidentity.NewDefaultAzureCredential(nil)
		if credErr != nil {
			return nil, fmt.Errorf("failed to create Azure credential: %w", credErr)
		}
		client, err = azblob.NewClient(url, credential, &azblob.ClientOptions{})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create Azure Blob Storage client: %w", err)
	}

	return &store{
		fetch: func(ctx context.Context, key string) ([]byte, error) {
			// Azure Blob Storage keys can't start with a slash
			key = strings.TrimPrefix(key, "/")

			// Download the blob
			response, err := client.DownloadStream(ctx, cfg.ContainerName, key, nil)
			if err != nil {
				return nil, err
			}

			defer response.Body.Close()

			return io.ReadAll(response.Body)
		},
		logger: logger,
		prefix: cfg.Prefix,
	}, nil
}
