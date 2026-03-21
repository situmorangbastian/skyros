# skyros

A production-ready e-commerce backend built with **Go microservices**, **gRPC** inter-service communication, and an **API Gateway** exposing a unified REST interface to clients.

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

```bash
make service-up     # Start all services via Docker Compose
make service-down   # Stop all services
```

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
