# syntax=docker/dockerfile:1

# Build the application from source
FROM golang:1.19 AS build-stage

WORKDIR /app

COPY ../../go.mod ./

RUN go mod download

# Copy the entire application code into the container
COPY ../ ./

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main

# Run tests in the container
FROM build-stage AS run-test-stage
RUN go test -v ./...

# Deploy the application binary into a lean image
FROM gcr.io/distroless/base-debian11 AS build-release-stage

WORKDIR /

# Copy the built binary from the build stage
COPY --from=build-stage /app/main /app/main

# Expose the necessary port
EXPOSE 8080

# Run as non-root user
USER nonroot:nonroot

# Set the entrypoint to run the Go application
ENTRYPOINT ["/app/main"]
