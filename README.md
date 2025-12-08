<p align="center">
  <img 
    src="https://github.com/user-attachments/assets/4bdad212-b6e8-444e-aefb-560796c3c56d"
    alt="https://iconbu.com/illust/4214"
    width="200"
    style="pointer-events: none;"
  />
</p>

# go-user-service

A Go-based user authentication service built with Clean Architecture, featuring Google OAuth2 login/signup, secure access/refresh token flows with rotation, and a clean modular structure.
The service exposes RESTful APIs using Gorilla Mux, and uses PostgreSQL and Redis for data persistence and token/session management.

## Features

- Clean Architecture with clear separation of concerns (domain, use case, interfaces, infrastructure)
- Google OAuth2 authentication (login & signup)
- Access/Refresh token flow with rotation and proper invalidation
- Secure token storage & validation
- REST API built with Gorilla Mux
- PostgreSQL for persistent user data
- Redis for managing refresh tokens, blacklisting, and sessions
- Modular, testable codebase

## Getting Started

1. Clone the repository:

```sh
git clone https://github.com/KimNattanan/go-user-service.git
cd go-user-service
```

2. Install Go module dependencies:

```sh
go mod tidy
```

3. Copy the environment file `.env.example`, rename it to `.env.development`, and configure it.

4. Start the databases using Docker Compose:

```sh
docker-compose up -d
```

5. Run the application:

```sh
go run ./cmd/app
```

## Project Structure

```
.
├── cmd/app/main.go
├── internal
│   ├── app
│   │   ├── app.go
│   │   └── server.go
│   ├── dto
│   │   ├── preference.go
│   │   └── user.go
│   ├── entity
│   │   ├── preference.go
│   │   ├── user.go
│   │   └── session.go
│   ├── handler
│   │   └── rest
│   │       ├── preference.go
│   │       └── user.go
│   ├── middleware
│   │   ├── auth.go
│   │   └── cors.go
│   ├── repo
│   │   ├── preference
│   │   │   └── preference.go
│   │   ├── session
│   │   │   └── session.go
│   │   ├── user
│   │   │   └── user.go
│   │   └── interface.go
│   └── usecase
│       ├── preference
│       │   └── preference.go
│       ├── session
│       │   └── session.go
│       ├── user
│       │   └── user.go
│       └── interface.go
├── pkg
│   ├── apperror/
│   ├── database/
│   ├── redisclient/
│   ├── routes
│   │   ├── private_routes.go
│   │   └── public_routes.go
│   └── token
├── .env.example
├── .gitignore
├── docker-compose.yml
├── go.mod
├── LICENSE
└── README.md
```
