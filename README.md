# skyros

This project demonstrates a Go-based microservices architecture using **gRPC** and **REST**

## Table of Contents

- [Project Structure](#project-structure)
- [Features](#features)
- [Technologies Used](#technologies-used)
- [Documentation](#documentation)
- [Getting Started](#getting-started)

## Project Structure

```bash
├── docs                        # API Docs
├── postgres                    # Init Database for microservice
├── orderservice                # Service for handle order
├── productservice              # Service for handle product
├── gatewayservice              # Service for API Gateway
├── skyrosgrpc                  # Package for Grpc Library
├── userservice                 # Service for handle user
├── docker-compose.yml          # Docker compose file for all service
├── Dockerfile-orderservice     # Docker file for orderservice
├── Dockerfile-productservice   # Docker file for productservice
└── Dockerfile-userservice      # Docker file for userservice
```

## Features

- **gRPC Microservices**: Each service is defined using protobuf and communicates via gRPC.
- **Modular Structure**: Separation of concerns with independent REST and gRPC entry points.

## Technologies Used

- [Go](https://golang.org/)
- [gRPC](https://grpc.io/)
- [Protobuf](https://developers.google.com/protocol-buffers) (Service Definitions)
- [Mockery](https://github.com/vektra/mockery) (Mocks for Unit Testing)

## Documentation

Documentation using ReDoc

```bash
cd docs
go run main.go
```

and open this link to see the api documentation:

[API Docs](http://localhost:8080/docs).

## Getting Started

### Database Migrations

Using CLI version of <https://github.com/golang-migrate/migrate>

- [Installation](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)

### Running

To start and stop the Services, run:

```bash
make service-up
make service-down
```
