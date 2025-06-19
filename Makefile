GO_MODULE := apac-ai-api

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## run: run API application
.PHONY: run/api
run/api:
	go run ./cmd/api -config-path="./config.yaml" -is-local=true

## run: run cli
.PHONY: run/cli
run/cli:
	go run ./cmd/cli -config-path="./config.yaml" -is-local=true


# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## audit: tidy dependencies and format, vet and test all code
.PHONY: audit
audit: tidy
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	staticcheck ./...
	@echo 'Running tests...'
	go test -race -vet=off ./...

## vendor: tidy and vendor dependencies
.PHONY: vendor 
vendor:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy 
	go mod verify 
	@echo 'Vendoring dependencies...'
	go mod vendor

.PHONY: tidy 
tidy:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy 
	go mod verify 

# ==================================================================================== #
# BUILD
# ==================================================================================== #

## build/api: build the cmd/api application
.PHONY: build/api
build/api:
	@echo 'Building cmd/api...'
	go build -ldflags="-s -w" -gcflags=all=-l -o=./bin/api ./cmd/api
	GOOS=linux go build -ldflags="-s -w" -gcflags=all=-l -o=./bin/linux/api ./cmd/api

## build/cli: build the cmd/api application
.PHONY: build/cli
build/cli:
	@echo 'Building cmd/cli...'
	go build -ldflags="-s -w" -gcflags=all=-l -o=./bin/cli ./cmd/cli
	GOOS=linux go build -ldflags="-s -w" -gcflags=all=-l -o=./bin/linux/cli ./cmd/cli



# ==================================================================================== #
# BUILD DEBUG
# ==================================================================================== #

## build/debug/api: build the api with debugging flags enabled
.PHONY: build/debug/api
build/debug/api:
	@echo 'Building cmd/api...'
	go build -gcflags=all="-N -l" -o=./bin/api-debug ./cmd/api

## build/debug/cli: build the api with debugging flags enabled
.PHONY: build/debug/cli
build/debug/cli:
	@echo 'Building cmd/cli...'
	go build -gcflags=all="-N -l" -o=./bin/cli-debug ./cmd/cli

