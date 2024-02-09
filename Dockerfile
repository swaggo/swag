# Dockerfile References: https://docs.docker.com/engine/reference/builder/

# Start from the latest golang base image
FROM --platform=$BUILDPLATFORM golang:1.21-alpine as builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Configure go compiler target platform
ARG TARGETOS
ARG TARGETARCH
ENV GOARCH=$TARGETARCH \
    GOOS=$TARGETOS

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -v -a -installsuffix cgo -o swag cmd/swag/main.go


######## Start a new stage from scratch #######
FROM --platform=$TARGETPLATFORM scratch

WORKDIR /code/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/swag /bin/swag

ENTRYPOINT ["/bin/swag"]
