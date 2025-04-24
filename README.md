# Gateway repository

## How to build

### Submodules initialization

`git submodule update --init`

### Install required packages

```bash
sudo apt  install golang-go
go install github.com/swaggo/swag/cmd/swag@latest
```

### Export path

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

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
docker build -t registry.github.com/NiskuT/cross-api:latest .
```

### Push the container to the registry

```bash
docker push registry.github.com/NiskuT/cross-api:latest
```
