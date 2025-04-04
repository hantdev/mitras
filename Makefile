# Define the Docker image name prefix for all Mitras services
MITRAS_DOCKER_IMAGE_NAME_PREFIX ?= hantdev1

# Output directory for compiled binaries
BUILD_DIR ?= build

# List of all available services in the system
SERVICES = auth users clients groups channels domains http coap ws cli mqtt certs journal

# Services that have API tests available
TEST_API_SERVICES = journal auth certs http clients users channels groups domains

# Create targets for API testing of each service
TEST_API = $(addprefix test_api_,$(TEST_API_SERVICES))

# Create targets for building Docker images of each service
DOCKERS = $(addprefix docker_,$(SERVICES))

# Create targets for building development Docker images of each service
DOCKERS_DEV = $(addprefix docker_dev_,$(SERVICES))

# Go build configuration
CGO_ENABLED ?= 0
GOARCH ?= arm64

# Version information for builds
VERSION ?= $(shell git describe --abbrev=0 --tags 2>/dev/null || echo 'unknown')
COMMIT ?= $(shell git rev-parse HEAD)
TIME ?= $(shell date +%F_%T)

# Extract repository owner and name from git origin URL
USER_REPO ?= $(shell git remote get-url origin | sed -E 's#.*[:/]([^/:]+)/([^/.]+)(\.git)?#\1_\2#')

# Define empty and space variables for string manipulation
empty:=
space:= $(empty) $(empty)

# Set Docker Compose project name based on repository info, following Docker naming guidelines
DOCKER_PROJECT ?= $(shell echo $(subst $(space),,$(USER_REPO)) | sed -E 's/[^a-zA-Z0-9]/_/g' | tr '[:upper:]' '[:lower:]')

# Supported Docker Compose commands and default command
DOCKER_COMPOSE_COMMANDS_SUPPORTED := up down config restart
DEFAULT_DOCKER_COMPOSE_COMMAND  := up

# Flag to check if gRPC mTLS certificate files exist
GRPC_MTLS_CERT_FILES_EXISTS = 0

# Version of mockery tool to use for generating mocks
MOCKERY_VERSION=v2.53.2

# Directories for generated protocol buffer code
PKG_PROTO_GEN_OUT_DIR=api/grpc
INTERNAL_PROTO_DIR=internal/proto
INTERNAL_PROTO_FILES := $(shell find $(INTERNAL_PROTO_DIR) -name "*.proto" | sed 's|$(INTERNAL_PROTO_DIR)/||')

# Set message broker type (defaults to nats if not specified)
ifneq ($(MITRAS_MESSAGE_BROKER_TYPE),)
    MITRAS_MESSAGE_BROKER_TYPE := $(MITRAS_MESSAGE_BROKER_TYPE)
else
    MITRAS_MESSAGE_BROKER_TYPE=nats
endif

# Set event store type (defaults to nats if not specified)
ifneq ($(MITRAS_ES_TYPE),)
    MITRAS_ES_TYPE := $(MITRAS_ES_TYPE)
else
    MITRAS_ES_TYPE=nats
endif

# Function to compile a service with proper build flags and tags
define compile_service
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) GOARM=$(GOARM) \
	go build -tags $(MITRAS_MESSAGE_BROKER_TYPE) --tags $(MITRAS_ES_TYPE) -ldflags "-s -w \
	-X 'github.com/hantdev/mitras.BuildTime=$(TIME)' \
	-X 'github.com/hantdev/mitras.Version=$(VERSION)' \
	-X 'github.com/hantdev/mitras.Commit=$(COMMIT)'" \
	-o ${BUILD_DIR}/$(1) cmd/$(1)/main.go
endef

# Function to build a Docker image for a service
define make_docker
	$(eval svc=$(subst docker_,,$(1)))

	docker build \
		--no-cache \
		--build-arg SVC=$(svc) \
		--build-arg GOARCH=$(GOARCH) \
		--build-arg GOARM=$(GOARM) \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg TIME=$(TIME) \
		--tag=$(MITRAS_DOCKER_IMAGE_NAME_PREFIX)/$(svc) \
		-f docker/Dockerfile .
endef

# Function to build a development Docker image for a service
define make_docker_dev
	$(eval svc=$(subst docker_dev_,,$(1)))

	docker build \
		--no-cache \
		--build-arg SVC=$(svc) \
		--tag=$(MITRAS_DOCKER_IMAGE_NAME_PREFIX)/$(svc) \
		-f docker/Dockerfile.dev ./build
endef

# Services categorized as addons
ADDON_SERVICES = journal certs

# External services that can be run as dependencies
EXTERNAL_SERVICES = vault prometheus

# Parse command line arguments for run targets
ifneq ($(filter run%,$(firstword $(MAKECMDGOALS))),)
  temp_args := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  DOCKER_COMPOSE_COMMAND := $(if $(filter $(DOCKER_COMPOSE_COMMANDS_SUPPORTED),$(temp_args)), $(filter $(DOCKER_COMPOSE_COMMANDS_SUPPORTED),$(temp_args)), $(DEFAULT_DOCKER_COMPOSE_COMMAND))
  $(eval $(DOCKER_COMPOSE_COMMAND):;@)
endif

# Parse command line arguments for run_addons targets
ifneq ($(filter run_addons%,$(firstword $(MAKECMDGOALS))),)
  temp_args := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  RUN_ADDON_ARGS :=  $(if $(filter-out $(DOCKER_COMPOSE_COMMANDS_SUPPORTED),$(temp_args)), $(filter-out $(DOCKER_COMPOSE_COMMANDS_SUPPORTED),$(temp_args)),$(ADDON_SERVICES) $(EXTERNAL_SERVICES))
  $(eval $(RUN_ADDON_ARGS):;@)
endif

# Check if gRPC mTLS certificate files exist
ifneq ("$(wildcard docker/ssl/certs/*-grpc-*)","")
GRPC_MTLS_CERT_FILES_EXISTS = 1
else
GRPC_MTLS_CERT_FILES_EXISTS = 0
endif

# Filter out addon services from the list of services to be built
FILTERED_SERVICES = $(filter-out $(RUN_ADDON_ARGS), $(SERVICES))

# Default target: build all services
all: $(SERVICES)

# Declare phony targets (targets that don't create files)
.PHONY: all $(SERVICES) dockers dockers_dev latest release run run_addons grpc_mtls_certs check_mtls check_certs test_api mocks

# Remove build artifacts
clean:
	rm -rf ${BUILD_DIR}

# Clean Docker resources
cleandocker:
	# Stops containers and removes containers, networks, volumes, and images created by up
	docker compose -f docker/docker-compose.yaml -p $(DOCKER_PROJECT) down --rmi all -v --remove-orphans

# Conditionally remove unused volumes if pv is specified
ifdef pv
	# Remove unused volumes
	docker volume ls -f name=$(MITRAS_DOCKER_IMAGE_NAME_PREFIX) -f dangling=true -q | xargs -r docker volume rm
endif

# Install compiled binaries to GOBIN directory
install:
	for file in $(BUILD_DIR)/*; do \
		cp $$file $(GOBIN)/mitras-`basename $$file`; \
	done

# Generate mock objects for testing
mocks:
	@which mockery > /dev/null || go install github.com/vektra/mockery/v2@$(MOCKERY_VERSION)
	@unset MOCKERY_VERSION && go generate ./...
	mockery --config ./tools/config/.mockery.yaml

# Run tests with coverage for specified directories
DIRS = consumers readers postgres internal
test: mocks
	mkdir -p coverage
	@for dir in $(DIRS); do \
        go test -v --race -count 1 -tags test -coverprofile=coverage/$$dir.out $$(go list ./... | grep $$dir | grep -v 'cmd'); \
    done
	go test -v --race -count 1 -tags test -coverprofile=coverage/coverage.out $$(go list ./... | grep -v 'consumers\|readers\|postgres\|internal\|cmd\|middleware')

# Function to run API tests for a service
define test_api_service
	$(eval svc=$(subst test_api_,,$(1)))
	@which st > /dev/null || (echo "schemathesis not found, please install it from https://github.com/schemathesis/schemathesis#getting-started" && exit 1)

	@if [ -z "$(USER_TOKEN)" ]; then \
		echo "USER_TOKEN is not set"; \
		echo "Please set it to a valid token"; \
		exit 1; \
	fi

	@if [ "$(svc)" = "http" ] && [ -z "$(CLIENT_SECRET)" ]; then \
		echo "CLIENT_SECRET is not set"; \
		echo "Please set it to a valid secret"; \
		exit 1; \
	fi

	@if [ "$(svc)" = "http" ]; then \
		st run api/openapi/$(svc).yaml \
		--checks all \
		--base-url $(2) \
		--header "Authorization: Client $(CLIENT_SECRET)" \
		--contrib-openapi-formats-uuid \
		--hypothesis-suppress-health-check=filter_too_much \
		--stateful=links; \
	else \
		st run api/openapi/$(svc).yaml \
		--checks all \
		--base-url $(2) \
		--header "Authorization: Bearer $(USER_TOKEN)" \
		--contrib-openapi-formats-uuid \
		--hypothesis-suppress-health-check=filter_too_much \
		--stateful=links; \
	fi
endef

# Define API test URLs for each service
test_api_users: TEST_API_URL := http://localhost:9002
test_api_clients: TEST_API_URL := http://localhost:9006
test_api_domains: TEST_API_URL := http://localhost:9003
test_api_channels: TEST_API_URL := http://localhost:9005
test_api_groups: TEST_API_URL := http://localhost:9004
test_api_http: TEST_API_URL := http://localhost:8008
test_api_auth: TEST_API_URL := http://localhost:9001
test_api_certs: TEST_API_URL := http://localhost:9019
test_api_journal: TEST_API_URL := http://localhost:9021

# Rule to run API tests for each service
$(TEST_API):
	$(call test_api_service,$(@),$(TEST_API_URL))

# Generate code from Protocol Buffer definitions
proto:
	protoc -I. --go_out=. --go_opt=paths=source_relative pkg/messaging/*.proto
	mkdir -p $(PKG_PROTO_GEN_OUT_DIR)
	protoc -I $(INTERNAL_PROTO_DIR) --go_out=$(PKG_PROTO_GEN_OUT_DIR) --go_opt=paths=source_relative --go-grpc_out=$(PKG_PROTO_GEN_OUT_DIR) --go-grpc_opt=paths=source_relative $(INTERNAL_PROTO_FILES)

# Rules to build individual services
$(FILTERED_SERVICES):
	$(call compile_service,$(@))

# Rules to build Docker images for each service
$(DOCKERS):
	$(call make_docker,$(@),$(GOARCH))

# Rules to build development Docker images for each service
$(DOCKERS_DEV):
	$(call make_docker_dev,$(@))

# Build all Docker images
dockers: $(DOCKERS)

# Build all development Docker images
dockers_dev: $(DOCKERS_DEV)

# Function to push Docker images with a specific tag
define docker_push
	for svc in $(SERVICES); do \
		docker push $(MITRAS_DOCKER_IMAGE_NAME_PREFIX)/$$svc:$(1); \
	done
endef

# Generate a changelog since the last tag
changelog:
	git log $(shell git describe --tags --abbrev=0)..HEAD --pretty=format:"- %s"

# Build and tag Docker images with 'latest' tag
latest: dockers
	$(call docker_push,latest)

# Build and tag Docker images with the version tag from git
release:
	$(eval version = $(shell git describe --abbrev=0 --tags))
	git checkout $(version)
	$(MAKE) dockers
	for svc in $(SERVICES); do \
		docker tag $(MITRAS_DOCKER_IMAGE_NAME_PREFIX)/$$svc $(MITRAS_DOCKER_IMAGE_NAME_PREFIX)/$$svc:$(version); \
	done
	$(call docker_push,$(version))

# Run development environment
rundev:
	cd scripts && ./run.sh

# Generate gRPC mTLS certificates
grpc_mtls_certs:
	$(MAKE) -C docker/ssl auth_grpc_certs clients_grpc_certs

# Check if TLS is enabled for gRPC
check_tls:
ifeq ($(GRPC_TLS),true)
	@unset GRPC_MTLS
	@echo "gRPC TLS is enabled"
	GRPC_MTLS=
else
	@unset GRPC_TLS
	GRPC_TLS=
endif

# Check if mTLS is enabled for gRPC
check_mtls:
ifeq ($(GRPC_MTLS),true)
	@unset GRPC_TLS
	@echo "gRPC MTLS is enabled"
	GRPC_TLS=
else
	@unset GRPC_MTLS
	GRPC_MTLS=
endif

# Check and generate certificates if needed
check_certs: check_mtls check_tls
ifeq ($(GRPC_MTLS_CERT_FILES_EXISTS),0)
ifeq ($(filter true,$(GRPC_MTLS) $(GRPC_TLS)),true)
ifeq ($(filter $(DEFAULT_DOCKER_COMPOSE_COMMAND),$(DOCKER_COMPOSE_COMMAND)),$(DEFAULT_DOCKER_COMPOSE_COMMAND))
	$(MAKE) -C docker/ssl auth_grpc_certs clients_grpc_certs
endif
endif
endif

# Run the system using Docker Compose
run: check_certs
	docker compose -f docker/docker-compose.yaml --env-file docker/.env -p $(DOCKER_PROJECT) $(DOCKER_COMPOSE_COMMAND) $(args)

# Run addon services using Docker Compose
run_addons: check_certs
	$(foreach SVC,$(RUN_ADDON_ARGS),$(if $(filter $(SVC),$(ADDON_SERVICES) $(EXTERNAL_SERVICES)),,$(error Invalid Service $(SVC))))
	@for SVC in $(RUN_ADDON_ARGS); do \
		MITRAS_ADDONS_CERTS_PATH_PREFIX="../."  docker compose -f docker/addons/$$SVC/docker-compose.yaml -p $(DOCKER_PROJECT) --env-file ./docker/.env $(DOCKER_COMPOSE_COMMAND) $(args) & \
	done

# Run the system in live development mode
run_live: check_certs
	GOPATH=$(go env GOPATH) docker compose  -f docker/docker-compose.yaml -f docker/docker-compose-live.yaml --env-file docker/.env -p $(DOCKER_PROJECT) $(DOCKER_COMPOSE_COMMAND) $(args)