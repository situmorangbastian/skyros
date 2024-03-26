# skyros

Simple Ecommerce Services

## Getting Started

### Project Structure

```bash
├── docs                        # API Docs
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

Documentation using swagger

Go to `docs` folder, and run this

```bash
go run main.go
```

and open this link to see the api documentation:

[API Docs](http://localhost:8080/swagger/index.html).

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
