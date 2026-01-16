//go:build integration
// +build integration

package sourcemapprocessor

import (
	"context"
	"testing"
	"time"

	"github.com/honeycombio/opentelemetry-collector-symbolicator/sourcemapprocessor/internal/metadata"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap/zaptest"
)

func TestAzureWithAzurite(t *testing.T) {
	ctx := context.Background()
	cfg := createDefaultConfig().(*Config)
	testTel := componenttest.NewTelemetry()
	tb, err := metadata.NewTelemetryBuilder(testTel.NewTelemetrySettings())
	defer tb.Shutdown()

	assert.NoError(t, err)

	attributes := attribute.NewSet(
		attribute.String("processor_type", "symbolicator"),
	)
	localAzuriteConfig := &AzureSourceMapConfiguration{
		ContainerName:    "source-maps",
		ConnectionString: "DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;BlobEndpoint=http://localhost:10000/devstoreaccount1",
	}
	cfg.SourceMapStoreKey = "azure_store"
	cfg.AzureSourceMapConfiguration = localAzuriteConfig

	store, err := newAzureStore(ctx, zaptest.NewLogger(t), localAzuriteConfig)
	assert.NoError(t, err)
	s, err := newBasicSymbolicator(ctx, 1*time.Second, 128, store, tb, attributes)

	assert.NoError(t, err)

	processorInstance := newSymbolicatorProcessor(ctx, cfg, processor.Settings{
		TelemetrySettings: component.TelemetrySettings{
			Logger: zaptest.NewLogger(t),
		},
	}, s, tb, attributes)

	// Create test data based on the provided JSON
	obfuscatedExceptionSpan := generateObfuscatedSpan()

	// Process the traces
	symbolisedTraces, err := processorInstance.processTraces(ctx, obfuscatedExceptionSpan)
	assert.NoError(t, err)

	symbolisedStackTrace, found := symbolisedTraces.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).Attributes().Get("exception.stacktrace")
	assert.True(t, found)

	assert.Equal(t, "Error: I blewed up!\n    at bar(basic-mapping-original.js:8:1)", symbolisedStackTrace.Str())
}

func TestS3WithS3Mock(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cfg := createDefaultConfig().(*Config)
	testTel := componenttest.NewTelemetry()
	tb, err := metadata.NewTelemetryBuilder(testTel.NewTelemetrySettings())
	defer tb.Shutdown()

	assert.NoError(t, err)

	attributes := attribute.NewSet(
		attribute.String("processor_type", "symbolicator"),
	)
	localS3MockConfig := &S3SourceMapConfiguration{
		Region:     "us-east-1",
		BucketName: "source-maps",
		Prefix:     "",
		Endpoint:   "http://localhost:9090",
	}
	cfg.SourceMapStoreKey = "s3_store"
	cfg.S3SourceMapConfiguration = localS3MockConfig

	// Set AWS credentials for S3Mock (these are dummy values)
	t.Setenv("AWS_ACCESS_KEY_ID", "test")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test")

	store, err := newS3Store(ctx, zaptest.NewLogger(t), localS3MockConfig)
	assert.NoError(t, err)
	s, err := newBasicSymbolicator(ctx, 1*time.Second, 128, store, tb, attributes)

	assert.NoError(t, err)

	processorInstance := newSymbolicatorProcessor(ctx, cfg, processor.Settings{
		TelemetrySettings: component.TelemetrySettings{
			Logger: zaptest.NewLogger(t),
		},
	}, s, tb, attributes)

	// Create test data based on the provided JSON
	obfuscatedExceptionSpan := generateObfuscatedSpan()

	// Process the traces
	symbolisedTraces, err := processorInstance.processTraces(ctx, obfuscatedExceptionSpan)
	assert.NoError(t, err)

	symbolisedStackTrace, found := symbolisedTraces.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).Attributes().Get("exception.stacktrace")
	assert.True(t, found)

	assert.Equal(t, "Error: I blewed up!\n    at bar(basic-mapping-original.js:8:1)", symbolisedStackTrace.Str())
}

func generateObfuscatedSpan() ptrace.Traces {
	td := ptrace.NewTraces()
	rs := td.ResourceSpans().AppendEmpty()
	ils := rs.ScopeSpans().AppendEmpty()

	// Set scope information
	ils.Scope().SetName("@honeycombio/instrumentation-global-errors")
	ils.Scope().SetVersion("1.0.2")

	span := ils.Spans().AppendEmpty()
	span.SetName("test-azure-span")
	span.SetTraceID([16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16})

	// Set exception attributes from the JSON data
	span.Attributes().PutStr("exception.type", "Error")
	span.Attributes().PutStr("exception.message", "I blewed up!")
	span.Attributes().PutStr("exception.stacktrace", "Error: I blewed up!\n    at i (http://localhost:4173/basic-mapping.js:0:34)")

	// Set structured stacktrace columns
	columnsSlice := span.Attributes().PutEmpty("exception.structured_stacktrace.columns").SetEmptySlice()
	columns := []int64{34}
	for _, col := range columns {
		columnsSlice.AppendEmpty().SetInt(col)
	}

	// Set structured stacktrace lines
	linesSlice := span.Attributes().PutEmpty("exception.structured_stacktrace.lines").SetEmptySlice()
	lines := []int64{0}
	for _, line := range lines {
		linesSlice.AppendEmpty().SetInt(line)
	}

	// Set structured stacktrace functions
	functionsSlice := span.Attributes().PutEmpty("exception.structured_stacktrace.functions").SetEmptySlice()
	functions := []string{"i"}
	for _, fn := range functions {
		functionsSlice.AppendEmpty().SetStr(fn)
	}

	// Set structured stacktrace URLs
	urlsSlice := span.Attributes().PutEmpty("exception.structured_stacktrace.urls").SetEmptySlice()
	urls := []string{
		"http://localhost:4173/basic-mapping.js",
	}
	for _, url := range urls {
		urlsSlice.AppendEmpty().SetStr(url)
	}

	// combine them all onto resourceSpans

	return td
}
