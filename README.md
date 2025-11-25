# go-auth-service
Golang Auth Microservice Enterprise Grade

## Architecture
The service is built using a modular architecture, separating concerns into 
different packages for better maintainability and scalability. Folder components 
include:
- cmd/
  - api/
    - main.go -- Entry point for the API server
- internal/
  - config/:
    - config.go -- Configuration loading, config struct definitions
  - http/:
    - server.go -- HTTP server setup and route definitions
    - middleware/ -- Middleware implementations (e.g., logging, authentication)
      - auth_middleware.go
      - logging.go
      - ....
    - handler/
      - auth_handler.go -- Handlers for authentication endpoints
      - user_handler.go -- Handlers for user management endpoints
      - ....
    - domain/
      - user.go -- User entity and related methods
      - token.go -- Token entity and related methods
    - repository/
      - user_repository.go -- interface and implementation for user data access
      - session_repository.go -- refresh tokens, device sessions, etc
      - migration/
        - user.go -- Database migration for user table
        - session.go -- Database migration for session table
    - service/
      - auth_service.go -- Core business rules: login, refresh, logout
      - user_service.go -- User CRUD, profile, etc.
      - password_service.go -- Password hashing and validation
    - security/ -- cryptographic operations
      - jwt_manager.go -- JWT token generation and validation
      - password_hasher.go -- Password hashing utilities
      - ....
    - store/
      - db.go -- Database connection and setup
      - redis.go -- Redis connection for session management
- pkg/
  - logger/
    - logger.go -- Pretty logger setup using zap
  - metrics/
    - prometheus.go -- Prometheus metrics setup