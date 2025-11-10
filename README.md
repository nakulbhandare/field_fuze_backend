
# FieldFuze Backend API

[![Go Version](https://img.shields.io/badge/Go-1.23.2-blue.svg)](https://golang.org/)
[![Gin Framework](https://img.shields.io/badge/Framework-Gin-green.svg)](https://github.com/gin-gonic/gin)
[![DynamoDB](https://img.shields.io/badge/Database-DynamoDB-orange.svg)](https://aws.amazon.com/dynamodb/)

A robust, scalable backend API built with Go, Gin framework, and AWS DynamoDB. Features comprehensive user management, role-based access control, infrastructure automation, and real-time monitoring.

## 🚀 Quick Start

```bash
# Clone the repository
git clone <repository-url>
cd fieldfuze-backend

# Install dependencies
go mod download

# Configure environment
cp config.json.example config.json
# Edit config.json with your settings

# Run the application
go run main.go
```

## 📋 Documentation Navigation

This documentation is organized into comprehensive sections covering each layer of the application:

### 🏗️ Architecture & Structure
- **[Project Structure](docs/PROJECT_STRUCTURE.md)** - Overall project organization and file layout
- **[Architecture Overview](docs/ARCHITECTURE.md)** - System design and component relationships

### 🎯 Core Components
- **[Controllers](docs/CONTROLLERS.md)** - HTTP request handlers and routing
- **[Services](docs/SERVICES.md)** - Business logic and service layer patterns
- **[Models](docs/MODELS.md)** - Data structures and database models
- **[Middleware](docs/MIDDLEWARE.md)** - Authentication, CORS, logging, and request processing
- **[Repository](docs/REPOSITORY.md)** - Data access layer and database operations

### 🔧 Infrastructure & Tools
- **[Worker System](docs/WORKERS.md)** - Background jobs and infrastructure automation
- **[Utilities](docs/UTILITIES.md)** - Helper functions, logging, and common tools
- **[Database Layer](docs/DATABASE.md)** - DynamoDB integration and data management

### 📚 API & Development
- **[API Documentation](docs/API.md)** - Complete API reference and examples
- **[Development Guide](docs/DEVELOPMENT.md)** - Setup, testing, and contribution guidelines
- **[EC2 Deployment](DEPLOYMENT.md)** - Traditional EC2 deployment guide
- **[ECS Deployment](DEPLOYMENT-ECS.md)** - Containerized ECS deployment guide

## ✨ Key Features

- **🔐 Advanced Authentication**: JWT-based auth with role-based access control
- **👥 User Management**: Complete user lifecycle with role assignments
- **🛡️ Security**: Multi-layer security with permission-based resource access
- **🏗️ Infrastructure Automation**: Automated DynamoDB table and index management
- **📊 Monitoring**: Real-time worker status and health monitoring
- **📖 API Documentation**: Interactive Swagger documentation
- **🔄 Background Workers**: Automated infrastructure maintenance
- **📝 Comprehensive Logging**: Structured logging with configurable levels

## 🏗️ Architecture Overview

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Controllers   │    │    Services     │    │   Repository    │
│                 │────│                 │────│                 │
│ • User          │    │ • User          │    │ • User          │
│ • Role          │    │ • Role          │    │ • Role          │
│ • Infrastructure│    │ • Infrastructure│    │ • Database      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Middleware    │    │     Models      │    │   DynamoDB      │
│                 │    │                 │    │                 │
│ • Auth          │    │ • User          │    │ • Tables        │
│ • CORS          │    │ • Role          │    │ • Indexes       │
│ • Logging       │    │ • Config        │    │ • Operations    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 🛠️ Technology Stack

- **Backend**: Go 1.23.2
- **Framework**: Gin Web Framework
- **Database**: AWS DynamoDB
- **Authentication**: JWT (golang-jwt/jwt/v5)
- **Documentation**: Swagger/OpenAPI
- **Logging**: Sirupsen Logrus
- **Validation**: Go Playground Validator
- **Configuration**: Viper
- **Scheduling**: Robfig Cron

## 🚦 Getting Started

### Prerequisites

- Go 1.23.2 or higher
- AWS Account with DynamoDB access
- AWS CLI configured (optional)

### Installation

1. **Clone and Setup**
   ```bash
   git clone <repository-url>
   cd fieldfuze-backend
   go mod download
   ```

2. **Configuration**
   ```bash
   # Copy example config
   cp config.json.example config.json
   
   # Edit with your settings
   vim config.json
   ```

3. **Run Application**
   ```bash
   go run main.go
   ```

4. **Access API Documentation**
   ```
   http://localhost:8081/swagger
   ```

## 📊 API Endpoints Overview

### Authentication
- `POST /api/v1/auth/user/register` - User registration
- `POST /api/v1/auth/user/login` - User login
- `POST /api/v1/auth/user/logout` - User logout
- `POST /api/v1/auth/user/validate` - Token validation

### User Management
- `GET /api/v1/auth/user/:id` - Get user details
- `GET /api/v1/auth/user/list` - List users with pagination
- `PATCH /api/v1/auth/user/update/:id` - Update user

### Role Management
- `GET /api/v1/auth/user/role` - List roles
- `POST /api/v1/auth/user/role` - Create role
- `GET /api/v1/auth/user/role/:id` - Get role details
- `PUT /api/v1/auth/user/role/:id` - Update role
- `DELETE /api/v1/auth/user/role/:id` - Delete role

### Infrastructure
- `GET /api/v1/infrastructure/worker/status` - Worker status
- `GET /api/v1/infrastructure/worker/health` - Health check
- `POST /api/v1/infrastructure/worker/restart` - Restart worker

## 🤝 Contributing

Please read our [Development Guide](docs/DEVELOPMENT.md) for details on our code of conduct, development process, and how to submit pull requests.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 📞 Support

For support and questions:
- Create an issue in the repository
- Check the documentation in the `docs/` folder
- Review the API documentation at `/swagger`

---

**Quick Links**: [Controllers](docs/CONTROLLERS.md) | [Services](docs/SERVICES.md) | [Models](docs/MODELS.md) | [API Docs](docs/API.md) | [Development](docs/DEVELOPMENT.md) | [EC2 Deploy](DEPLOYMENT.md) | [ECS Deploy](DEPLOYMENT-ECS.md)