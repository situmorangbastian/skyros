# skyros

A portfolio project demonstrating a **Go microservices** architecture with **gRPC** inter-service communication and an **API Gateway** exposing a unified REST interface to clients. Built to showcase backend engineering practices including service decomposition, JWT authentication, database-per-service pattern, and containerised local development

## 🧩 Key Features

- **API Gateway** — single entry point, JWT authentication, request routing via gRPC
- **User Service** — registration, login, JWT issuance
- **Product Service** — product catalog management
- **Order Service** — order creation and management
- **Database per service** — isolated PostgreSQL databases per microservice
- **Containerised** — full Docker Compose setup with health checks and dependency ordering

## 🛠 Tech Stack

| Layer | Technology |
| --- | --- |
| Language | Go 1.26 |
| API Gateway | gRPC-Gateway |
| Inter-service comm | gRPC / Protocol Buffers |
| Database | PostgreSQL 16 |
| Authentication | JWT |
| Containerisation | Docker / Docker Compose |
| Migration | golang-migrate |

## 🏗 Architecture

![Architecture](docs/assets/architecture.png)

## Request flow — place an order

![Request flow](docs/assets/request-flow.png)

## Project Structure

```text
├── gatewayservice        # REST API gateway, routes to microservices via gRPC
├── userservice           # User registration, login, JWT auth
├── productservice        # Product catalog CRUD
├── orderservice          # Order management, calls user + product service
├── proto                 # Shared Protobuf definitions
├── serviceutils          # Shared utilities across services
├── postgres              # PostgreSQL init scripts
└── docs                  # API documentation (ReDoc)
```

## Getting Started

### Prerequisites

- [Docker](https://www.docker.com/) and [Docker Compose](https://docs.docker.com/compose/)
- [Go 1.26+](https://golang.org/)
- [golang-migrate](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)
- [protoc](https://grpc.io/docs/protoc-installation/) (only needed to regenerate proto files)
- [golangci-lint](https://golangci-lint.run/usage/install/) (for linting)

### Setup

#### 1. Clone the repo

```bash
git clone https://github.com/situmorangbastian/skyros.git
cd skyros
```

#### 2. Configure environment**

```bash
cp .env.example .env
# edit .env if needed — defaults work out of the box
```

#### 3. Start all services**

```bash
make service-up
```

#### 4. Verify services are running**

```bash
docker compose ps
```

The gateway is available at `http://localhost:4000`.

### Stopping

```bash
make service-down
```

## API Documentation

Interactive API docs powered by ReDoc:

```bash
cd docs
cp .env.example .env
go run main.go
```

## Environment Variables

See [`.env.example`](.env.example) for all available configuration options. Key variables:

| Variable | Description |
| --- | --- |
| `POSTGRES_USER` / `POSTGRES_PASSWORD` | Database credentials |
| `USER_SECRET_KEY` | JWT signing key for user service |
| `APP_ENV` | `development` or `production` |
| `ENABLE_GATEWAY_GRPC` | Enable gRPC gateway passthrough |

## Available Make Commands

| Command | Description |
| --- | --- |
| `make service-up` | Start all services via Docker Compose |
| `make service-down` | Stop all services |
| `make build` | Compile all services |
| `make test` | Run tests with race detector and coverage |
| `make lint` | Run golangci-lint |
| `make tidy` | Tidy and verify go modules |
| `make migrate-up` | Run all database migrations |
| `make migrate-down` | Rollback all database migrations |

## 🎯 Project Goals

- Demonstrate a modular, production-aware microservice architecture in Go
- Showcase gRPC for efficient, type-safe inter-service communication
- Use an API Gateway pattern to unify and secure public-facing API access
- Serve as a reference architecture for real-world distributed backend systems

## Documentation

Documentation using ReDoc

```bash
cd docs
cp .env.example .env
go run main.go
```

### Running

```bash
cp .env.example .env
# edit .env with your values, then:
make service-up
```
