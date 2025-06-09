.PHONY: all generate clean

all: generate doc

export:
	@echo "Exporting PATH variable..."
	@echo "Copying the following line and paste it in your shell:"
	export PATH="$${PATH}:$$(go env GOPATH)/bin"

clean:
	@echo "Cleaning generated files..."
	rm -rf $(OUT_DIR)
	@echo "Removed $(OUT_DIR)"
	@echo "Cleaning swagger documentation..."
	rm -rf docs
	@echo "Removed docs"

build:
	CGO_ENABLED=0 go build -o api cmd/api/main.go

production:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o api cmd/api/main.go

start:
	export APP_ENV=local && go run cmd/api/main.go rest

doc:
	@echo "You need to install swaggo/swag using go install github.com/swaggo/swag/cmd/swag@latest"
	swag init --dir ./cmd/api,./internal/server --output ./docs --parseDependency --parseInternal

help:
	@echo "Please use 'make <target>' where <target> is one of"
	@echo "  export      to export the PATH variable"
	@echo "  clean       to remove generated files"
	@echo "  build       to build the gateway"
	@echo "  start       to start the gateway"
	@echo "  doc         to generate swagger documentation"
