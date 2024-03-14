# skyros

Simple Ecommerce Services

## Getting Started

### Project Structure

```bash
├── init                        # Init Migration for Database
├── orderservice                # Service for handle order
├── productservice              # Service for handle product
├── reverseproxyservice         # Service for API Gateway
├── skyrosgrpc                  # Package for Grpc Library
├── userservice                 # Service for handle user
├── docker-compose.yml          # Docker compose file for all service
├── Dockerfile-orderservice     # Docker file for orderservice
├── Dockerfile-productservice   # Docker file for productservice
└── Dockerfile-userservice      # Docker file for userservice
```

### Documentation

On this link [documentation](https://app.swaggerhub.com/apis-docs/situmorangbastian/skyros/1.0.0).

### Database Migrations

Using CLI version of <https://github.com/golang-migrate/migrate>

* [Installation](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)

### Running

Make sure to set the .env file (see: .env.example).

To start and stop the Services, run:

```bash
make service-up
make service-down
```

Run migration.

```bash
make migrate-up
```

### Dummy User

User Seller

```bash
"email": "user-seller@example.com",
"password": "password"
```

User Buyer

```bash
"email": "user-buyer@example.com",
"password": "password"
```
