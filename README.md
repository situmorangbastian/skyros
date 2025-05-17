# skyros

This project demonstrates a Go-based microservices architecture using **gRPC** and **REST**

## Table of Contents

- [Project Structure](#project-structure)
- [Documentation](#documentation)
- [Getting Started](#getting-started)

## Project Structure

```bash
├── docs                        # API Docs
├── gatewayservice              # Service for API Gateway
├── orderservice                # Service for handle order
├── postgres                    # Init Database for microservice
├── productservice              # Service for handle product
├── proto                       # Package for Grpc Library
└── userservice                 # Service for handle user
```

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
