# Risk-Based Audit Backend

A microservice-based backend for Risk-Based Internal Audit system built with Go, PostgreSQL, Redis, and Docker.

## 🏗️ Architecture

This backend follows a clean architecture pattern with clear separation of concerns:

```
backend/
├── cmd/                    # Service entry points
│   ├── auth/              # Authentication service
│   ├── audit/             # Audit management service
│   ├── risk/              # Risk assessment service
│   ├── dashboard/         # Dashboard analytics service
│   └── gateway/           # API Gateway
├── internal/              # Private application code
│   ├── auth/             # Auth service implementation
│   │   ├── controllers/  # HTTP handlers
│   │   ├── services/     # Business logic
│   │   ├── repositories/ # Data access layer
│   │   ├── routes/       # Route definitions
│   │   ├── registry/     # Dependency injection
│   │   └── models/       # Domain models & DTOs
│   ├── audit/            # Audit service (same structure)
│   ├── risk/             # Risk service (same structure)
│   └── dashboard/        # Dashboard service (same structure)
├── pkg/                  # Reusable packages
│   ├── config/          # Configuration management
│   ├── database/        # Database connections
│   ├── redis/           # Redis client
│   ├── logger/          # Logging utilities
│   ├── middleware/      # HTTP middleware
│   ├── validations/     # Request validation
│   ├── utils/           # Helper functions
│   ├── response/        # HTTP response helpers
│   └── errors/          # Custom error types
├── configs/             # Configuration files
├── migrations/          # Database migrations
├── docker/             # Docker configurations
├── k8s/                # Kubernetes manifests
└── proto/              # gRPC protobuf definitions
```

## 🚀 Technology Stack

- **Language**: Go 1.21
- **Web Framework**: Gin
- **Database**: PostgreSQL with GORM
- **Cache**: Redis
- **Authentication**: JWT (golang-jwt/jwt)
- **Password Hashing**: bcrypt
- **Configuration**: Viper
- **Logging**: Zap
- **Validation**: go-playground/validator
- **Deployment**: Docker & Docker Compose

## 📋 Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- PostgreSQL 16
- Redis 7

## 🔧 Installation

1. **Clone the repository**
   ```bash
   cd backend
   ```

2. **Install dependencies**
   ```bash
   make deps
   ```

3. **Configure environment**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. **Start services with Docker**
   ```bash
   make docker-up
   ```

Or run locally:

```bash
# Start PostgreSQL and Redis
docker-compose up -d postgres redis

# Run the auth service
make run
```

## 🏃 Running Services

### Using Make commands

```bash
# Build all services
make build

# Run auth service
make run

# Run tests
make test

# Format code
make fmt

# Run linter
make lint
```

### Using Docker Compose

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

## 📡 API Endpoints

### Auth Service (`:8001`)

#### Public Routes
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/register` - User registration

#### Protected Routes (require JWT)
- `POST /api/v1/auth/logout` - User logout
- `POST /api/v1/auth/change-password` - Change password
- `GET /api/v1/auth/me` - Get current user info

### User Management (`:8001`)

All routes require JWT authentication:

- `POST /api/v1/users` - Create user
- `GET /api/v1/users` - List users (with pagination)
- `GET /api/v1/users/:id` - Get user by ID
- `PUT /api/v1/users/:id` - Update user
- `DELETE /api/v1/users/:id` - Delete user

### Health Check

All services have a health check endpoint:
- `GET /health` - Service health status

## 🔐 Authentication

The API uses JWT (JSON Web Token) for authentication. Include the token in the Authorization header:

```
Authorization: Bearer <your-jwt-token>
```

## 📝 Example Requests

### Login

```bash
curl -X POST http://localhost:8001/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "password123"
  }'
```

### Create User (Protected)

```bash
curl -X POST http://localhost:8001/api/v1/users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your-token>" \
  -d '{
    "username": "johndoe",
    "email": "john@example.com",
    "password": "password123",
    "full_name": "John Doe",
    "department": "IT",
    "roles": ["auditor"]
  }'
```

### List Users (Protected)

```bash
curl -X GET http://localhost:8001/api/v1/users?page=1&page_size=10 \
  -H "Authorization: Bearer <your-token>"
```

## 🗄️ Database Schema

### Users Table
- `id` (UUID, Primary Key)
- `username` (VARCHAR, Unique)
- `email` (VARCHAR, Unique)
- `password` (VARCHAR, Hashed)
- `full_name` (VARCHAR)
- `phone` (VARCHAR)
- `department` (VARCHAR)
- `is_active` (BOOLEAN)
- `created_at` (TIMESTAMP)
- `updated_at` (TIMESTAMP)

### Roles Table
- `id` (UUID, Primary Key)
- `name` (VARCHAR, Unique)
- `description` (TEXT)
- `created_at` (TIMESTAMP)
- `updated_at` (TIMESTAMP)

### Permissions Table
- `id` (UUID, Primary Key)
- `name` (VARCHAR, Unique)
- `resource` (VARCHAR)
- `action` (VARCHAR)
- `description` (TEXT)
- `created_at` (TIMESTAMP)
- `updated_at` (TIMESTAMP)

## 🧪 Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific service tests
go test ./internal/auth/...
```

## 🐳 Docker

### Build Images

```bash
make docker-build
```

### Run Services

```bash
make docker-up
```

### View Logs

```bash
make docker-logs
```

### Stop Services

```bash
make docker-down
```

## 🔧 Configuration

Configuration is loaded from `configs/config.yaml`:

```yaml
server:
  port: 8001
  mode: release
  read_timeout: 60

database:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  dbname: rb_audit
  sslmode: disable

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0

jwt:
  secret: your-jwt-secret
  expiry_hour: 24

log:
  level: info
  format: json
```

## 📦 Dependency Injection

Each service uses a registry pattern for dependency injection:

```go
registry := registry.NewRegistry(cfg, db, rdb, validator, engine)
registry.RegisterRoutes()
```

The registry handles:
- Repository initialization
- Service initialization
- Controller initialization
- Route registration
- Middleware setup

## 🛠️ Development

### Adding a New Service

1. Create the service directory structure under `internal/`
2. Copy the layer structure from `internal/auth/`
3. Implement your models, repositories, services, controllers
4. Create a registry for dependency injection
5. Add a main entry point under `cmd/`
6. Update Docker Compose if needed

### Adding a New Endpoint

1. Define the request/response DTOs in `models/dto.go`
2. Add the repository method in `repositories/`
3. Add the service logic in `services/`
4. Add the controller handler in `controllers/`
5. Register the route in `routes/routes.go`

## 📊 Performance Requirements

Based on requirements.MD, the system targets:

- Batch upload processing: 3 seconds for 10,000 records
- Search query response: 5 seconds for 1 million records
- Dashboard load: 2 seconds under 100 concurrent users
- Email notifications: 6 seconds maximum
- Report generation: 60 seconds for 200 pages

## 🔒 Security

- Password hashing with bcrypt
- JWT-based authentication
- Role-based access control (RBAC)
- CORS support
- Request validation
- SQL injection prevention (via GORM)
- Data encryption at rest (PostgreSQL)

## 📝 License

This project is proprietary software.

## 🤝 Contributing

1. Create a feature branch
2. Make your changes
3. Run tests and linters
4. Submit a pull request

## Local Infrastructure 
psql -U $(whoami) //check postgres
psql -d postgres //connect to db with postgres as entry point
psql -U postgres -c "CREATE DATABASE auth_service;"
