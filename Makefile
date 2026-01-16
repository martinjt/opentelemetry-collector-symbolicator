.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build                    - Build the OpenTelemetry collector"
	@echo "  build-docker            - Build Docker image"
	@echo "  run                     - Run the collector with config.yaml"
	@echo "  test                    - Run tests for all processors"
	@echo "  test-integration        - Run integration tests"
	@echo "  clean                   - Clean build artifacts"
	@echo ""
	@echo "Azurite (Azure Storage Emulator) targets:"
	@echo "  azurite-start           - Start Azurite container"
	@echo "  azurite-stop            - Stop Azurite container"
	@echo "  azurite-logs            - Show Azurite container logs"
	@echo "  azurite-list-container  - Show Azurite container status"
	@echo "  upload-test-files-azurite - Upload test files to Azurite"
	@echo "  azurite-list-files      - List files in Azurite containers"
	@echo "  azurite-clean           - Clean up Azurite containers and data"
	@echo ""
	@echo "S3Mock (AWS S3 Emulator) targets:"
	@echo "  s3mock-start            - Start S3Mock container"
	@echo "  s3mock-stop             - Stop S3Mock container"
	@echo "  s3mock-logs             - Show S3Mock container logs"
	@echo "  s3mock-list-container   - Show S3Mock container status"
	@echo "  upload-test-files-s3mock - Upload test files to S3Mock"
	@echo "  s3mock-list-files       - List files in S3Mock buckets"
	@echo "  s3mock-clean            - Clean up S3Mock containers and data"
	@echo ""
	@echo "For detailed S3Mock testing instructions, see S3MOCK_TESTING.md"

.PHONY: builder
builder:
	go install go.opentelemetry.io/collector/cmd/builder@v0.142.0

.PHONY: clean
clean:
	rm -rf otelcol-dev

.PHONY: build
build: builder
	builder --config builder-config.yaml

.PHONY: build-docker
build-docker:
	docker buildx build . -t collector-symbolicator-processor

.PHONY: run
run: build
	go run ./otelcol-dev --config config.yaml

.PHONY: test
test: build
	go test -tags integration ./sourcemapprocessor/ ./dsymprocessor ./proguardprocessor

.PHONY: test-integration
test-integration: build
	cd sourcemapprocessor && go test -tags=integration -v -run TestAzureWithAzurite

.PHONY: upload-test-files-azurite
upload-test-files-azurite:
	@echo "Uploading test files to Azurite..."
	@# Set Azure CLI to use Azurite
	@export AZURE_STORAGE_CONNECTION_STRING="DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;BlobEndpoint=http://localhost:10000/devstoreaccount1;" && \
	echo "Creating containers..." && \
	az storage container create --name source-maps --connection-string "$$AZURE_STORAGE_CONNECTION_STRING" || true && \
	az storage container create --name dsyms --connection-string "$$AZURE_STORAGE_CONNECTION_STRING" || true && \
	az storage container create --name proguard-mappings --connection-string "$$AZURE_STORAGE_CONNECTION_STRING" || true && \
	echo "Uploading basic source map files..." && \
	az storage blob upload --file test_assets/basic-mapping.js --container-name source-maps --name basic-mapping.js --connection-string "$$AZURE_STORAGE_CONNECTION_STRING" --overwrite && \
	az storage blob upload --file test_assets/basic-mapping.js.map --container-name source-maps --name basic-mapping.js.map --connection-string "$$AZURE_STORAGE_CONNECTION_STRING" --overwrite && \
	echo "Uploading dSYM files..." && \
	az storage blob upload-batch --source test_assets/6A8CB813-45F6-3652-AD33-778FD1EAB196.dSYM --destination dsyms --destination-path ios/symbols/6A8CB813-45F6-3652-AD33-778FD1EAB196.dSYM --connection-string "$$AZURE_STORAGE_CONNECTION_STRING" --overwrite && \
	echo "Test files uploaded successfully to Azurite!"

.PHONY: azurite-start
azurite-start:
	@echo "Starting Azurite container..."
	docker compose up -d azurite

.PHONY: azurite-stop
azurite-stop:
	@echo "Stopping Azurite container..."
	docker compose down

.PHONY: azurite-logs
azurite-logs:
	@echo "Showing Azurite logs..."
	docker compose logs -f azurite

.PHONY: azurite-list-container
azurite-list-container:
	@echo "Listing Azurite container status..."
	@docker ps -a --filter "name=azurite" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}\t{{.Image}}"

.PHONY: azurite-list-files
azurite-list-files:
	@echo "Listing files in Azurite containers..."
	@export AZURE_STORAGE_CONNECTION_STRING="DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;BlobEndpoint=http://localhost:10000/devstoreaccount1;" && \
	echo "=== Source Maps Container ===" && \
	az storage blob list --container-name source-maps --connection-string "$$AZURE_STORAGE_CONNECTION_STRING" --output table || echo "Container 'source-maps' not found or empty" && \
	echo "=== dSYMs Container ===" && \
	az storage blob list --container-name dsyms --connection-string "$$AZURE_STORAGE_CONNECTION_STRING" --output table || echo "Container 'dsyms' not found or empty" && \
	echo "=== ProGuard Mappings Container ===" && \
	az storage blob list --container-name proguard-mappings --connection-string "$$AZURE_STORAGE_CONNECTION_STRING" --output table || echo "Container 'proguard-mappings' not found or empty"

.PHONY: azurite-clean
azurite-clean:
	@echo "Cleaning up Azurite containers and data..."
	@export AZURE_STORAGE_CONNECTION_STRING="DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;BlobEndpoint=http://localhost:10000/devstoreaccount1;" && \
	az storage container delete --name source-maps --connection-string "$$AZURE_STORAGE_CONNECTION_STRING" || true && \
	az storage container delete --name dsyms --connection-string "$$AZURE_STORAGE_CONNECTION_STRING" || true && \
	az storage container delete --name proguard-mappings --connection-string "$$AZURE_STORAGE_CONNECTION_STRING" || true && \
	echo "Containers cleaned up!"

# S3Mock targets
.PHONY: upload-test-files-s3mock
upload-test-files-s3mock:
	@echo "Uploading test files to S3Mock using curl..."
	@echo "Uploading basic source map files..." && \
	curl -X PUT --upload-file test_assets/basic-mapping.js http://localhost:9090/source-maps/basic-mapping.js && \
	curl -X PUT --upload-file test_assets/basic-mapping.js.map http://localhost:9090/source-maps/basic-mapping.js.map && \
	echo "Test files uploaded successfully to S3Mock!"

.PHONY: upload-test-files-s3mock-aws
upload-test-files-s3mock-aws:
	@echo "Uploading test files to S3Mock using AWS CLI..."
	@export AWS_ACCESS_KEY_ID=test && \
	export AWS_SECRET_ACCESS_KEY=test && \
	export AWS_DEFAULT_REGION=us-east-1 && \
	echo "Uploading basic source map files..." && \
	aws s3 cp test_assets/basic-mapping.js s3://source-maps/basic-mapping.js --endpoint-url=http://localhost:9090 && \
	aws s3 cp test_assets/basic-mapping.js.map s3://source-maps/basic-mapping.js.map --endpoint-url=http://localhost:9090 && \
	echo "Uploading dSYM files..." && \
	aws s3 cp test_assets/6A8CB813-45F6-3652-AD33-778FD1EAB196.dSYM s3://dsyms/ios/symbols/6A8CB813-45F6-3652-AD33-778FD1EAB196.dSYM --recursive --endpoint-url=http://localhost:9090 && \
	echo "Test files uploaded successfully to S3Mock!"

.PHONY: s3mock-start
s3mock-start:
	@echo "Starting S3Mock container..."
	docker compose up -d s3mock

.PHONY: s3mock-stop
s3mock-stop:
	@echo "Stopping S3Mock container..."
	docker compose down

.PHONY: s3mock-logs
s3mock-logs:
	@echo "Showing S3Mock logs..."
	docker compose logs -f s3mock

.PHONY: s3mock-list-container
s3mock-list-container:
	@echo "Listing S3Mock container status..."
	@docker ps -a --filter "name=s3mock" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}\t{{.Image}}"

.PHONY: s3mock-list-files
s3mock-list-files:
	@echo "Listing files in S3Mock buckets..."
	@echo "Note: S3Mock does not support listing bucket contents via HTTP API"
	@echo "Use 'make s3mock-list-files-aws' if you have AWS CLI installed"
	@echo "Or check the S3Mock logs to see uploaded files"

.PHONY: s3mock-list-files-aws
s3mock-list-files-aws:
	@echo "Listing files in S3Mock buckets using AWS CLI..."
	@export AWS_ACCESS_KEY_ID=test && \
	export AWS_SECRET_ACCESS_KEY=test && \
	export AWS_DEFAULT_REGION=us-east-1 && \
	echo "=== Source Maps Bucket ===" && \
	aws s3 ls s3://source-maps --endpoint-url=http://localhost:9090 || echo "Bucket 'source-maps' not found or empty" && \
	echo "=== dSYMs Bucket ===" && \
	aws s3 ls s3://dsyms --recursive --endpoint-url=http://localhost:9090 || echo "Bucket 'dsyms' not found or empty" && \
	echo "=== ProGuard Mappings Bucket ===" && \
	aws s3 ls s3://proguard-mappings --endpoint-url=http://localhost:9090 || echo "Bucket 'proguard-mappings' not found or empty"

.PHONY: s3mock-clean
s3mock-clean:
	@echo "Cleaning up S3Mock buckets..."
	@echo "Note: Restart S3Mock container to clean all data, or use 'make s3mock-clean-aws' if you have AWS CLI"
	@echo "Restarting S3Mock container to clean data..."
	@docker compose down s3mock || true
	@docker compose up -d s3mock
	@echo "S3Mock restarted with clean data!"

.PHONY: s3mock-clean-aws
s3mock-clean-aws:
	@echo "Cleaning up S3Mock buckets using AWS CLI..."
	@export AWS_ACCESS_KEY_ID=test && \
	export AWS_SECRET_ACCESS_KEY=test && \
	export AWS_DEFAULT_REGION=us-east-1 && \
	aws s3 rm s3://source-maps --recursive --endpoint-url=http://localhost:9090 || true && \
	aws s3 rm s3://dsyms --recursive --endpoint-url=http://localhost:9090 || true && \
	aws s3 rm s3://proguard-mappings --recursive --endpoint-url=http://localhost:9090 || true && \
	echo "Buckets cleaned up!"

