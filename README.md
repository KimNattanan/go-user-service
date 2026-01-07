<p align="center">
  <a href='#'>
    <img 
      src="https://github.com/user-attachments/assets/4bdad212-b6e8-444e-aefb-560796c3c56d"
      alt="https://iconbu.com/illust/4214"
      width="200"
    />
  </a>
</p>

# go-user-service

A Go-based user authentication service built using Clean Architecture.\
It supports Google OAuth2, secure access/refresh token rotation, and uses PostgreSQL + Redis for persistence and session management.

## Features

- Clean Architecture with clear separation of concerns
- Google OAuth2 authentication (login & signup)
- Access/Refresh token flow with rotation and proper invalidation
- Secure token storage & validation
- REST API built with Gorilla Mux
- PostgreSQL for persistent user data
- Redis for managing refresh tokens and sessions
- Modular, testable codebase
- Built-in Swagger documentation

## Prerequisites

- Go 1.24+
- Docker & Docker Compose

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

3. Configure environment variables

    Copy `.env.example`, rename it to `.env.development`, then configure it.

4. Start the databases using Docker Compose:

    ```sh
    docker-compose up -d
    ```

5. Run the application:

    ```sh
    go run ./cmd/app
    ```

6. Test:

   ```sh
   go test ./pkg/routes
   ```

See Swagger UI at: `localhost:8000/swagger/index.html`

## Project Structure

```
.
├── cmd/app/main.go
├── docs/
├── internal
│   ├── app
│   │   ├── app.go
│   │   └── server.go
│   ├── dto
│   │   ├── preference.go
│   │   └── user.go
│   ├── entity
│   │   ├── preference.go
│   │   ├── session.go
│   │   └── user.go
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
│   ├── config/
│   ├── database/
│   ├── httpserver/
│   ├── redisclient/
│   ├── routes
│   │   ├── notfound_route.go
│   │   ├── private_routes.go
│   │   ├── public_routes.go
│   │   └── public_routes_test.go
│   └── token/
│── .env.example
│── .gitignore
│── docker-compose.yml
│── go.mod
│── LICENSE
└── README.md
```

## Endpoints

| Endpoint | Method | Description 
|-|-|-|
| /api/v1/auth/google/login | GET | Redirects to Google OAuth provider
| /api/v1/auth/google/callback | GET | Handles Google OAuth callback
| /api/v1/auth/logout | POST | Logout user
| /api/v1/me | GET | Get user
| /api/v1/me | PATCH | Update user info
| /api/v1/me | DELETE | Delete user
| /api/v1/me/preferences | GET | Get user's preferences
| /api/v1/me/preferences | PATCH | Update user's preferences
| /api/v1/users | GET | Find all users
| /api/v1/users/{id} | GET | Find user by userID

## License

This project is licensed under the MIT License.\
See the `LICENSE` file for details.
