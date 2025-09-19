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
	@echo "For detailed Azurite testing instructions, see AZURITE_TESTING.md"

.PHONY: builder
builder:
	go install go.opentelemetry.io/collector/cmd/builder@v0.129.0

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

