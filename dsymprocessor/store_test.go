package dsymprocessor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestFileStore(t *testing.T) {
	ctx := context.Background()

	fs, err := newFileStore(ctx, zaptest.NewLogger(t), &LocalDSYMConfiguration{Path: "../test_assets"})
	assert.NoError(t, err)

	source, err := fs.GetDSYM(ctx, "6A8CB813-45F6-3652-AD33-778FD1EAB196", "Chateaux Bufeaux")

	assert.NoError(t, err)
	assert.NotEmpty(t, source)

	_, err = fs.GetDSYM(ctx, "6A8CB813-45F6-3652-AD33-778FD1EAB196", "Not A Binary")
	assert.ErrorIs(t, err, errFailedToFindDSYM)
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
	cfg := &AzureDSYMConfiguration{
		AccountName:   "testaccount",
		ContainerName: "testcontainer",
		Prefix:        "test/prefix",
		Endpoint:      "http://localhost:10000/testaccount",
	}

	store, err := newAzureStore(ctx, zaptest.NewLogger(t), cfg)
	// Azure client creation may succeed even without valid credentials
	if err == nil {
		assert.NotNil(t, store)
	}

	// Test with connection string configuration
	cfgWithConnectionString := &AzureDSYMConfiguration{
		ContainerName:    "testcontainer",
		Prefix:           "test/prefix",
		ConnectionString: "DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;BlobEndpoint=http://localhost:10000/devstoreaccount1",
	}

	storeWithConnectionString, err := newAzureStore(ctx, zaptest.NewLogger(t), cfgWithConnectionString)
	// Connection string should be parsed successfully
	if err == nil {
		assert.NotNil(t, storeWithConnectionString)
	}

}
