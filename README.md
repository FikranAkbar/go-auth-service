# Go Auth Service
Production-ready authentication microservice built with Go, following Clean Architecture principles and industry best practices.

## ✨ Features

### 🔐 Authentication
- **User Registration** with email verification (2-step process)
- **Email Verification** with JWT tokens (24h expiry)
- **Resend Verification Email** for unverified accounts
- **Login** with email/password authentication
- **Logout** with refresh token invalidation
- **Refresh Token** mechanism for seamless token renewal

### 🛡️ Security
- **Bcrypt Password Hashing** (configurable cost factor)
- **JWT Authentication** with HS256 algorithm
- **Access Tokens** (15 minutes expiry) - stateless
- **Refresh Tokens** (7 days expiry) - stored in Redis
- **Token Type Validation** (access, refresh, verification)
- **Email Verification Required** before login
- **Transaction-based Registration** with automatic rollback on failure

### 📧 Email Integration
- **SMTP Email Service** using Gomail v2
- **Verification Email** with clickable links
- **Gmail Integration** with app-specific passwords

### 💾 Data Storage
- **PostgreSQL** for persistent data
- **Redis** for refresh token session management
- **Database Migrations** using Goose
- **Transaction Support** for data integrity

## 🏗️ Architecture

This service follows **Clean Architecture** principles with clear separation of concerns:

```
go-auth-service/
├── cmd/
│   └── api/
│       └── main.go                    # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go                  # Configuration management
│   ├── domain/                        # Business entities & interfaces
│   │   ├── auth/
│   │   │   ├── entity.go             # Auth DTOs (requests/responses)
│   │   │   └── service.go            # Auth service interface
│   │   ├── email/
│   │   │   └── service.go            # Email service interface
│   │   ├── repository/
│   │   │   └── token_repository.go   # Token repository interface
│   │   ├── security/
│   │   │   ├── jwt_manager.go        # JWT manager interface
│   │   │   └── password_hasher.go    # Password hasher interface
│   │   └── user/
│   │       ├── entity.go             # User entity
│   │       ├── repository.go         # User repository interface
│   │       └── service.go            # User service interface
│   ├── handler/                       # HTTP request handlers
│   │   ├── auth_handler.go           # Auth endpoints
│   │   ├── health_handler.go         # Health check
│   │   └── user_handler.go           # User endpoints
│   ├── http/
│   │   ├── server.go                 # HTTP server setup
│   │   ├── middleware/
│   │   │   ├── auth_middleware.go    # JWT authentication
│   │   │   └── logging.go            # Request logging
│   │   └── router/
│   │       └── init_router.go        # Route definitions
│   ├── infra/                        # Infrastructure setup
│   │   ├── db.go                     # PostgreSQL connection
│   │   ├── email.go                  # Email client setup
│   │   └── redis.go                  # Redis connection
│   ├── repository/                   # Data access layer
│   │   ├── token_repository.go       # Redis token storage
│   │   └── user_repository.go        # PostgreSQL user data
│   ├── security/                     # Security utilities
│   │   ├── jwt_manager.go            # JWT generation/validation
│   │   └── password_hasher.go        # Bcrypt password hashing
│   └── service/                      # Business logic
│       ├── auth_service.go           # Authentication logic
│       ├── email_service.go          # Email sending logic
│       └── user_service.go           # User management logic
├── pkg/                              # Shared utilities
│   ├── constants/                    # Application constants
│   ├── errors/                       # Error handling
│   ├── logger/                       # Structured logging (zerolog)
│   ├── metrics/                      # Prometheus metrics
│   └── response/                     # HTTP response helpers
├── migrations/                       # Database migrations (Goose)
│   └── 20251126130037_create_users_table.sql
├── api-contract/                     # API testing (Bruno)
├── docs/                             # Documentation
└── config.yaml                       # Configuration file
```

## 🛠️ Tech Stack

- **Language**: Go 1.25
- **Web Framework**: Chi Router v5
- **Database**: PostgreSQL (lib/pq driver)
- **Cache**: Redis v9
- **Authentication**: JWT (golang-jwt/jwt v5)
- **Password**: bcrypt (golang.org/x/crypto)
- **Email**: Gomail v2
- **Logging**: Zerolog
- **Migrations**: Goose
- **API Testing**: Bruno
- **Configuration**: YAML

## 📡 API Endpoints

### Health
- `GET /health` - Health check endpoint

### Authentication
- `POST /api/auth/register` - Register new user (sends verification email)
- `GET /api/auth/verify-email?token={token}` - Verify email address
- `POST /api/auth/resend-verification` - Resend verification email
- `POST /api/auth/login` - Login with email/password
- `POST /api/auth/refresh-token` - Refresh access token
- `POST /api/auth/logout` - Logout (requires authentication)

### Response Format

All endpoints return JSON responses with the following structure:

**Success Response:**
```json
{
  "success": true,
  "message": "Operation successful",
  "data": {
    // Response data here
  }
}
```

**Error Response:**
```json
{
  "success": false,
  "message": "Error message",
  "error": "Detailed error description"
}
```

**Example Login Response:**
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "user": {
      "id": 1,
      "email": "user@example.com",
      "username": "johndoe",
      "is_active": true,
      "created_at": "2026-01-11T10:30:00Z"
    },
    "access_token": "eyJhbGc...",
    "refresh_token": "eyJhbGc..."
  }
}
```

## 🚀 Getting Started

### Prerequisites
- Go 1.25+
- PostgreSQL
- Redis
- SMTP credentials (Gmail recommended)

### Installation

1. Clone the repository
```bash
git clone <repository-url>
cd go-auth-service
```

2. Copy configuration
```bash
cp config.yaml.example config.yaml
cp .env.example .env
```

3. Start required services (PostgreSQL & Redis)
```bash
make docker-up
```

4. Update `config.yaml` with your settings:
   - Database credentials
   - Redis connection
   - JWT secret key (generate with: `openssl rand -base64 64`)
   - SMTP credentials (for Gmail, use App Password, not regular password)

5. Update `.env` file:
```bash
DB_URL=postgres://admin:secret123@localhost:5432/goauthdb?sslmode=disable
```

6. Install dependencies
```bash
go mod download
```

7. Run database migrations
```bash
make migrate-up
```

8. Run the application
```bash
make run
# or with hot reload
make dev
```

The API will be available at `http://localhost:8080`

### Gmail SMTP Setup

For Gmail SMTP, you need to use an **App Password**:

1. Enable 2-Factor Authentication on your Google Account
2. Go to [Google App Passwords](https://myaccount.google.com/apppasswords)
3. Generate a new App Password for "Mail"
4. Copy the 16-character password (remove spaces)
5. Update `config.yaml`:
```yaml
email:
  smtp_username: your-email@gmail.com
  smtp_password: your-16-char-app-password  # No spaces!
```

## 🔧 Makefile Commands

```bash
make help                # Show all available commands

# Database Migrations
make migrate-up          # Apply all pending migrations
make migrate-down        # Rollback the last migration
make migrate-status      # Check migration status
make migrate-reset       # Rollback all migrations
make migrate-create name=migration_name  # Create new migration

# Development
make run                 # Run the application
make dev                 # Run with hot reload (air)
make build               # Build binary to bin/api

# Testing
make test                # Run all tests
make test-coverage       # Run tests with coverage report

# Docker
make docker-up           # Start PostgreSQL & Redis containers
make docker-down         # Stop containers
make docker-logs         # View container logs
```

## 🧪 Testing

API contract tests are available in the `api-contract/` directory using Bruno.

## 📝 Configuration

Configuration is managed via `config.yaml`:

```yaml
server:
  port: :8080
  read_timeout_duration: 10s
  write_timeout_duration: 10s

database:
  type: postgres
  host: localhost
  port: 5432
  username: admin
  password: secret123
  name: goauthdb

redis:
  host: localhost
  port: 6379

jwt:
  secret_key: your-secret-key
  access_token_expiry: 15m
  refresh_token_expiry: 168h  # 7 days
  bcrypt_cost: 10

email:
  smtp_host: smtp.gmail.com
  smtp_port: 587
  smtp_username: your-email@gmail.com
  smtp_password: your-app-password
```

## 📚 Documentation

Additional documentation available in `docs/`:
- `API_ENDPOINTS.md` - Detailed API endpoint documentation
- `DEPLOYMENT.md` - Deployment guides
- Other technical documentation

## 🔑 Key Design Decisions

1. **Clean Architecture**: Domain-driven design with clear boundaries
2. **Interface-based**: All major components use interfaces for testability
3. **Transaction Safety**: Registration uses DB transactions with automatic rollback
4. **Email Verification**: Required before login for security
5. **Stateless Access Tokens**: JWTs validated by signature only
6. **Stateful Refresh Tokens**: Stored in Redis for revocation capability
7. **Security First**: Bcrypt for passwords, HTTPS required in production

## 🚦 Quick Start Examples

### 1. Register a User
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "username": "johndoe",
    "password": "SecurePass123!"
  }'
```

### 2. Verify Email
Check your email for the verification link, or extract the token and:
```bash
curl -X GET "http://localhost:8080/api/auth/verify-email?token=YOUR_TOKEN"
```

### 3. Login
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!"
  }'
```

Save the `access_token` and `refresh_token` from the response.

### 4. Refresh Access Token
```bash
curl -X POST http://localhost:8080/api/auth/refresh-token \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "YOUR_REFRESH_TOKEN"
  }'
```

### 5. Logout
```bash
curl -X POST http://localhost:8080/api/auth/logout \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

This project is licensed under the MIT License.

## 🐛 Troubleshooting

### Common Issues

**Port 8080 already in use:**
```bash
# Find process using port 8080
lsof -i :8080
# Kill the process
kill -9 <PID>
```

**Database connection failed:**
- Ensure PostgreSQL is running: `make docker-up`
- Check credentials in `config.yaml` and `.env`

**Email not sending:**
- Use Gmail App Password, not regular password
- Ensure 2FA is enabled on Google Account
- Check SMTP settings in `config.yaml`

**Redis connection failed:**
- Ensure Redis is running: `make docker-up`
- Check Redis host/port in `config.yaml`

