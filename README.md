# Go Moon API Server

[![Go CI Pipeline](https://github.com/your-username/go-moon/actions/workflows/ci.yml/badge.svg)](https://github.com/your-username/go-moon/actions/workflows/ci.yml)

A RESTful API server built with Go, Gin, GORM, PostgreSQL, and Redis.

## Features
- User Authentication (Register, Login, Logout, Refresh Token)
- JWT for session management
- Password Hashing (bcrypt)
- User Profile Management (Change Password, Update Profile)
- TODO List Management (CRUD operations)
- Dockerized setup for development and deployment
- Makefile for common development tasks
- Live reloading with Air
- CI Pipeline with GitHub Actions for automated builds, linting, and unit tests.
- Deployment configuration for Render (`render.yaml`).

## System Prerequisites
- Go (version specified in `go.mod`, e.g., 1.22)
- Docker
- Docker Compose

## Local Development Prerequisites (Optional Tools)
For an enhanced local development experience, you can install the following tools:
- **Air:** For live reloading of the application during development.
  ```bash
  go install github.com/cosmtrek/air@latest
  ```
- **golangci-lint:** For running linters.
  ```bash
  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
  ```
You can also use the Makefile to install these:
```bash
make install-tools
```

## Setup & Running Locally (Without Docker)

1.  **Environment Variables:**
    Create a `.env` file by copying `.env.example`:
    ```bash
    cp .env.example .env
    ```
    Update `.env` with your local PostgreSQL and Redis connection details.
    For example:
    ```env
    POSTGRES_DSN=postgresql://youruser:yourpass@localhost:5432/yourdbname?sslmode=disable&TimeZone=UTC
    REDIS_ADDR=localhost:6379
    JWT_SECRET=yourlocaljwtsecret
    # ... other variables
    ```

2.  **Run PostgreSQL and Redis:**
    Ensure you have local instances of PostgreSQL and Redis running and accessible.

3.  **Run the Application:**
    *   Using `go run`:
        ```bash
        go run cmd/server/main.go
        ```
    *   Using `make run-dev` (which does the same):
        ```bash
        make run-dev
        ```
    *   For live reloading with Air (if installed):
        ```bash
        make air
        # or directly:
        # air
        ```
    The server will start on the port defined by `SERVER_PORT` in your `.env` file (default `8080`).

## Setup & Running with Docker

1.  **Environment Variables:**
    Create a `.env` file in the project root by copying `.env.example`:
    ```bash
    cp .env.example .env
    ```
    Review and update the variables in `.env` if needed. The defaults in `.env.example` are configured to work with the `docker-compose.yml` setup (e.g., `POSTGRES_HOST=db`, `REDIS_HOST=redis`). The `JWT_SECRET` should ideally be changed.

2.  **Build and Run Services:**
    Use Docker Compose to build the application image and start all services (app, PostgreSQL, Redis):
    ```bash
    make docker-build
    make docker-up
    # Or combined:
    # docker-compose up --build -d
    ```

3.  **Accessing the Application:**
    Once the services are running, the API server will be accessible at `http://localhost:8080` (or the `SERVER_PORT` you configured on the host in your `.env` file, which `docker-compose.yml` forwards).

4.  **Viewing Logs (App Service):**
    ```bash
    make docker-logs
    # Or directly:
    # docker-compose logs -f app
    ```

5.  **Stopping Services:**
    ```bash
    make docker-down
    # Or directly:
    # docker-compose down
    ```
    To stop and remove volumes (deletes database and cache data):
    ```bash
    docker-compose down -v
    ```

## Makefile Targets
Run `make help` to see a list of available targets, including:
- `build`: Build the Go application binary.
- `run`: Run the built application.
- `run-dev`: Run the application using `go run`.
- `test`: Run all unit and integration tests.
- `test-unit`: Run only unit tests.
- `test-integration`: Run only integration tests (ensure Docker services are up or local DB/Redis are configured and running).
- `lint`: Run golangci-lint.
- `fmt`: Format Go code using `go fmt`.
- `air`: Run with live reload.
- `install-tools`: Install `air` and `golangci-lint`.
- Docker commands: `docker-build`, `docker-up`, `docker-down`, `docker-logs`.

## CI/CD Pipeline
This project uses GitHub Actions for its CI pipeline. The workflow is defined in `.github/workflows/ci.yml`.
The pipeline currently includes:
- Setting up Go.
- Checking out code.
- Installing linters.
- Verifying Go modules (`go mod tidy`).
- Running `go vet` and `golangci-lint`.
- Running unit tests (with code coverage).
- Building the application binary.

Future enhancements could include:
- Running integration tests by setting up PostgreSQL and Redis services within the CI environment (an example job is commented out in `ci.yml`).
- Automated deployment steps (CD) to services like Render.
- Uploading code coverage reports to services like Codecov.

## Deployment to Render

This project includes a `render.yaml` file, which defines the infrastructure as code for deploying to [Render](https://render.com/).

**Steps to Deploy using the Blueprint:**

1.  **Sign up/Log in to Render.**
2.  **Create a New Blueprint Instance:**
    *   On the Render Dashboard, click "New +" -> "Blueprint".
    *   Connect the GitHub repository for this project.
    *   Render will automatically detect the `render.yaml` file in the root of your repository.
3.  **Configure and Deploy:**
    *   Render will display the services defined in the blueprint (web service, PostgreSQL database, Redis cache).
    *   You might need to confirm names or regions. Render will automatically handle generating secrets (like `JWT_SECRET`) and connection strings for the database and Redis, injecting them into the application service as environment variables.
    *   Click "Create New Services" (or similar) to provision the infrastructure and deploy your application.
4.  **Automatic Deploys:**
    *   By default (`autoDeploy: true` in `render.yaml`), Render will automatically redeploy your application whenever you push changes to the connected branch (e.g., `main`).

The `scripts/deploy.sh` script in this repository provides further informational guidance and is not an executable deployment script itself.

## API Endpoints
(TODO: Add a summary of API endpoints here, perhaps linking to Postman collection or Swagger docs if available)

- `POST /users/register`
- `POST /users/login`
- `POST /auth/refresh`
- `POST /auth/logout` (Protected)
- `POST /users/password` (Protected)
- `PUT /users/me` (Protected)
- `GET /users/me/profile` (Protected)
- `POST /todos` (Protected)
- `GET /todos` (Protected)
- `GET /todos/:todoID` (Protected)
- `PUT /todos/:todoID` (Protected)
- `DELETE /todos/:todoID` (Protected)
- `GET /healthz` (Health check)

## Running Tests

### Unit Tests
```bash
make test-unit
# or
# go test -v -race -coverprofile=coverage-unit.out $(go list ./... | grep -v /tests/integration)
```

### Integration Tests
The integration tests require Docker Compose services to be running, or a local PostgreSQL and Redis instance configured similarly to how `tests/integration/main_test.go` expects (via `.env` or environment variables). The tests will clear data from the configured test database. **DO NOT run integration tests against a production database.**

1.  Start dependent services (if using Docker for tests):
    ```bash
    make docker-up
    ```
2.  Run tests:
    ```bash
    make test-integration
    # Or from project root:
    # go test -v -race ./tests/integration/...
    ```
3.  Stop services after testing:
    ```bash
    make docker-down
    ```

(TODO: Add more details on configuration, specific environment variables for tests, etc.)
