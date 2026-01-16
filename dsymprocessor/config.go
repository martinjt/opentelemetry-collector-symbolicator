package dsymprocessor

import "time"

// Config defines configuration for the symbolicator processor.
type Config struct {
	// SymbolicatorFailureAttributeKey is the attribute key that will be set to
	// true if the symbolicator fails to fully symbolicate a stack trace.
	SymbolicatorFailureAttributeKey string `mapstructure:"symbolicator_failure_attribute_key"`

	// SymbolicatorErrorAttributeKey is the attribute key that contains the error
	// message if the symbolicator fails to fully symbolicate a stack trace.
	SymbolicatorErrorAttributeKey string `mapstructure:"symbolicator_error_attribute_key"`

	// StackTraceAttributeKey is the attribute key that contains an explicitly thrown stack trace.
	StackTraceAttributeKey string `mapstructure:"stack_trace_attribute_key"`

	// MetricKitStackTraceAttributeKey is the attribute key that contains the metrickit
	// stack trace.
	MetricKitStackTraceAttributeKey string `mapstructure:"metrickit_stack_trace_attribute_key"`

	// OutputMetricKitStackTraceAttributeKey is the attribute key that contains the
	// symbolicated metrickit stack trace.
	OutputMetricKitStackTraceAttributeKey string `mapstructure:"output_metrickit_stack_trace_attribute_key"`

	// OutputMetricKitStackTraceAttributeKey is the attribute key that contains the
	// inferred  metrickit stack trace.
	OutputMetricKitExceptionTypeAttributeKey string `mapstructure:"output_metrickit_exception_type_attribute_key"`

	// OutputMetricKitStackTraceAttributeKey is the attribute key that contains the
	// symbolicated metrickit stack trace.
	OutputMetricKitExceptionMessageAttributeKey string `mapstructure:"output_metrickit_exception_message_attribute_key"`

	// preserveStackTrace is a config option that determines whether to keep the
	// original stack trace in the output.
	PreserveStackTrace bool `mapstructure:"preserve_stack_trace"`

	// OriginalStackTraceAttributeKey is the attribute key that preserves the original stack
	// trace.
	OriginalStackTraceAttributeKey string `mapstructure:"original_stack_trace_attribute_key"`

	// BuildUUIDAttributeKey is the attribute key that contains the build UUID of the current app.
	BuildUUIDAttributeKey string `mapstructure:"build_uuid_attribute_key"`

	// AppExecutableAttributeKey is the attribute key that contains the current app's executable name.
	AppExecutableAttributeKey string `mapstructure:"app_executable_attribute_key"`

	DSYMStoreKey string `mapstructure:"dsym_store"`

	// LocalDSYMConfiguration is the configuration for sourcing source maps on a local volume.
	LocalDSYMConfiguration *LocalDSYMConfiguration `mapstructure:"local_dsyms"`

	// S3DSYMConfiguration is the configuration for sourcing source maps from S3.
	S3DSYMConfiguration *S3DSYMConfiguration `mapstructure:"s3_dsyms"`

	// GCSDSYMConfiguration is the configuration for sourcing source maps from GCS.
	GCSDSYMConfiguration *GCSDSYMConfiguration `mapstructure:"gcs_dsyms"`

	// AzureDSYMConfiguration is the configuration for sourcing dSYMs from Azure Blob Storage.
	AzureDSYMConfiguration *AzureDSYMConfiguration `mapstructure:"azure_dsyms"`

	// Timeout is the maximum time to wait for a response from the symbolicator.
	Timeout time.Duration `mapstructure:"timeout"`

	// CacheSize is the maximum number of dSYMs to cache.
	DSYMCacheSize int `mapstructure:"dsym_cache_size"`

	// LanguageAttributeKey is the attribute key that contains the programming language
	// or SDK language of the telemetry signal (e.g., "telemetry.sdk.language").
	// This is used to determine if this processor should handle the signal.
	LanguageAttributeKey string `mapstructure:"language_attribute_key"`

	// AllowedLanguages is a list of language values that this processor will handle.
	// If the signal's language attribute matches any value in this list, the processor will run.
	// If empty (default), the processor will process all signals regardless of language.
	AllowedLanguages []string `mapstructure:"allowed_languages"`
}

type LocalDSYMConfiguration struct {
	// Path is a file path to where the dSYMs are stored on disk.
	Path string `mapstructure:"path"`
}

type S3DSYMConfiguration struct {
	// Region is the AWS region where the S3 bucket is located.
	Region string `mapstructure:"region"`
	// BucketName is the name of the S3 bucket.
	BucketName string `mapstructure:"bucket"`
	// Prefix is the prefix to use when looking for dSYMs.
	Prefix string `mapstructure:"prefix"`
	// Endpoint is the S3 endpoint URL. If not specified, defaults to AWS S3.
	// For S3Mock testing, use: http://localhost:9090
	Endpoint string `mapstructure:"endpoint"`
}

type GCSDSYMConfiguration struct {
	// BucketName is the name of the GCS bucket.
	BucketName string `mapstructure:"bucket"`
	// Prefix is the prefix to use when looking for dSYMs.
	Prefix string `mapstructure:"prefix"`
}

type AzureDSYMConfiguration struct {
	// AccountName is the Azure storage account name.
	AccountName string `mapstructure:"account_name"`
	// ContainerName is the name of the Azure Blob Storage container.
	ContainerName string `mapstructure:"container"`
	// Prefix is the prefix to use when looking for dSYMs.
	Prefix string `mapstructure:"prefix"`
	// Endpoint is the Azure Blob Storage endpoint URL. If not specified, defaults to https://{account_name}.blob.core.windows.net/
	// For Azurite testing, use: http://localhost:10000/{account_name}
	Endpoint string `mapstructure:"endpoint"`
	// ConnectionString is the Azure Blob Storage connection string. If specified, takes precedence over account name and endpoint.
	ConnectionString string `mapstructure:"connection_string"`
}

// Validate checks the configuration for any issues.
func (c *Config) Validate() error {
	return nil
}
