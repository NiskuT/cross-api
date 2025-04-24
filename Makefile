# Directories
PROTO_DIR       := proto/protobuf-definition
OUT_DIR         := pkg/protobuf-generated

# Collect all .proto files
PROTOS          := $(shell find $(PROTO_DIR) -name "*.proto")

# Protoc plugin options
PROTOC_GEN_GO_OPTS         := paths=source_relative,Mcommons/commons.proto=gitlab.com/orkys/backend/gateway/pkg/protobuf-generated/commons
PROTOC_GEN_GO_GRPC_OPTS    := paths=source_relative,Mcommons/commons.proto=gitlab.com/orkys/backend/gateway/pkg/protobuf-generated/commons

.PHONY: all generate clean

all: generate doc

export:
	@echo "Exporting PATH variable..."
	@echo "Copying the following line and paste it in your shell:"
	export PATH="$${PATH}:$$(go env GOPATH)/bin"

generate:
	@echo "Generating protobuf and gRPC code..."
	mkdir -p $(OUT_DIR)
	protoc \
		-I=$(PROTO_DIR) \
		--go_out=$(OUT_DIR) \
		--go-grpc_out=$(OUT_DIR) \
		--go_opt=$(PROTOC_GEN_GO_OPTS) \
		--go-grpc_opt=$(PROTOC_GEN_GO_GRPC_OPTS) \
		$(PROTOS)
	@echo "Finished generating protobuf and gRPC code in $(OUT_DIR)"

clean:
	@echo "Cleaning generated files..."
	rm -rf $(OUT_DIR)
	@echo "Removed $(OUT_DIR)"
	@echo "Cleaning swagger documentation..."
	rm -rf docs
	@echo "Removed docs"

build:
	CGO_ENABLED=0 go build -o gateway cmd/gateway/main.go

start:
	export APP_ENV=local && go run cmd/gateway/main.go rest

doc:
	@echo "You need to install swaggo/swag using go install github.com/swaggo/swag/cmd/swag@latest"
	swag init --dir ./cmd/gateway,./internal/server --output ./docs --parseDependency --parseInternal

help:
	@echo "Please use 'make <target>' where <target> is one of"
	@echo "  export      to export the PATH variable"
	@echo "  generate    to generate protobuf and gRPC code"
	@echo "  clean       to remove generated files"
	@echo "  build       to build the gateway"
	@echo "  start       to start the gateway"
	@echo "  doc         to generate swagger documentation"
