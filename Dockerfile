# Stage 1: Build the Go application
FROM golang:1.22-alpine AS build

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download && go mod verify

# Copy the source code into the container
COPY . .

# Build the Go app
# CGO_ENABLED=0 for static linking (no C libraries needed)
# GOOS=linux to build for Linux environment
# -o /app/main specifies the output file name and location
RUN CGO_ENABLED=0 GOOS=linux go build -v -o /app/main cmd/server/main.go

# Stage 2: Setup the runtime environment
FROM alpine:latest

WORKDIR /app

# Copy the Pre-built binary file from the previous stage
COPY --from=build /app/main /app/main

# Optional: Copy .env.example as .env.
# This is for convenience if you want default env vars baked into the image (not recommended for secrets).
# Better to supply environment variables at runtime via docker-compose or orchestrator.
# If you do this, ensure .env does not contain actual secrets.
# COPY .env.example .env

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
ENTRYPOINT ["/app/main"]
