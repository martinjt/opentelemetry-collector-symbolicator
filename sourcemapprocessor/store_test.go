package sourcemapprocessor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestFileStore(t *testing.T) {
	ctx := context.Background()

	fs, err := newFileStore(ctx, zaptest.NewLogger(t), &LocalSourceMapConfiguration{Path: "../test_assets"})
	assert.NoError(t, err)

	source, sMap, err := fs.GetSourceMap(ctx, jsFile)

	assert.NoError(t, err)
	assert.NotEmpty(t, source)
	assert.NotEmpty(t, sMap)

	source, sMap, err = fs.GetSourceMap(ctx, noFile)
	assert.ErrorIs(t, err, errFailedToFindSourceFile)
}

func TestS3StoreConfiguration(t *testing.T) {
	ctx := context.Background()

	// Test with nil configuration
	_, err := newS3Store(ctx, zaptest.NewLogger(t), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no S3 configuration provided")
}

func TestGCSStoreConfiguration(t *testing.T) {
	ctx := context.Background()

	// Test with nil configuration
	_, err := newGCSStore(ctx, zaptest.NewLogger(t), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no GCS configuration provided")
}

func TestAzureStoreConfiguration(t *testing.T) {
	ctx := context.Background()

	// Test with nil configuration
	_, err := newAzureStore(ctx, zaptest.NewLogger(t), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no Azure configuration provided")

	// Test with valid configuration (will fail due to authentication, but validates config parsing)
	cfg := &AzureSourceMapConfiguration{
		AccountName:   "testaccount",
		ContainerName: "testcontainer",
		Prefix:        "test/prefix",
		Endpoint:      "http://localhost:10000/testaccount",
	}

	// This will fail due to authentication, but we can verify the configuration is processed
	store, err := newAzureStore(ctx, zaptest.NewLogger(t), cfg)
	if err != nil {
		// The error should be about Azure credentials, not configuration
		assert.Contains(t, err.Error(), "failed to create Azure credential")
	} else {
		// If no error, verify the store was created with correct prefix
		assert.NotNil(t, store)
		assert.Equal(t, cfg.Prefix, store.prefix)
	}
}
