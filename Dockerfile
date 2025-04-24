##############################
# Build Stage
##############################
FROM golang:1.23.3-alpine AS builder
WORKDIR /app

# Install necessary build tools
RUN apk add --no-cache make protobuf

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest  && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest && \
    go install github.com/swaggo/swag/cmd/swag@latest

# Cache dependencies first
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Run protobuf and documentation generation.
RUN make

# Build the binary.
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app cmd/gateway/main.go

##############################
# Final Stage
##############################
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

# Copy the binary from the build stage.
COPY --from=builder /app/app .

# Expose the port that your gateway uses (9000 in this case).
EXPOSE 9000

# Run the binary.
CMD ["./app", "rest"]
