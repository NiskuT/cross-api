# Gateway repository

## How to build

### Submodules initialization

`git submodule update --init`

### Install required packages

```bash
sudo apt  install golang-go
sudo apt install -y protobuf-compiler
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/swaggo/swag/cmd/swag@latest
```

### Export path

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

### Verify the install:

```bash
protoc-gen-go --version
protoc-gen-go-grpc --version
```

### Generate protobuf for Golang
Then run `make generate`

### Generate documentation
Run `make doc`

### Start the application
Run `make start`

### Access the documentation

Follow [this link](http://localhost:9000/swagger/index.html)

## Container

### Login to the registry

```bash
docker login registry.gitlab.com
```

### Build the container

From the root of this project. Run:

```bash
docker build -t registry.gitlab.com/orkys/backend/gateway:latest .
```

### Push the container to the registry

```bash
docker push registry.gitlab.com/orkys/backend/gateway:latest
```
