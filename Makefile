.PHONY: build build-client build-server test integration-test lint vet fmt cross-compile docker-up docker-down terraform-init terraform-apply terraform-destroy clean

BINARY_CLIENT = lan-party
BINARY_SERVER = lanpartyd
BUILD_DIR = bin
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS = -X github.com/ffa/lan-party/internal/version.Version=$(VERSION)

build: build-server build-client

build-server:
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_SERVER) ./cmd/lanpartyd

build-client:
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_CLIENT) ./cmd/lan-party

test:
	go test -race -count=1 ./...

integration-test:
	go test -race -count=1 -tags=integration ./...

lint:
	golangci-lint run ./...

vet:
	go vet ./...

fmt:
	gofmt -w .
	goimports -w . 2>/dev/null || true

cross-compile:
	GOOS=linux   GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_CLIENT)-linux-amd64   ./cmd/lan-party
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_CLIENT)-windows-amd64.exe ./cmd/lan-party
	GOOS=linux   GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_SERVER)-linux-amd64   ./cmd/lanpartyd

docker-up:
	docker compose -f deploy/docker/docker-compose.yml up -d

docker-down:
	docker compose -f deploy/docker/docker-compose.yml down

terraform-init:
	cd deploy/terraform && terraform init

terraform-apply:
	cd deploy/terraform && terraform apply

terraform-destroy:
	cd deploy/terraform && terraform destroy

terraform-validate:
	cd deploy/terraform && terraform validate

clean:
	rm -rf $(BUILD_DIR)
