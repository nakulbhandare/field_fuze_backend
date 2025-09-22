# API Documentation

Comprehensive API reference for the FieldFuze Backend API with detailed endpoints, authentication, and examples.

## 📋 Table of Contents

- [Overview](#overview)
- [Authentication](#authentication)
- [Base URL and Versioning](#base-url-and-versioning)
- [Response Format](#response-format)
- [Error Handling](#error-handling)
- [User Management API](#user-management-api)
- [Role Management API](#role-management-api)
- [Infrastructure API](#infrastructure-api)
- [Rate Limiting](#rate-limiting)
- [SDKs and Tools](#sdks-and-tools)

## Overview

The FieldFuze Backend API is a RESTful API that provides:
- **User Authentication & Management**: Complete user lifecycle management
- **Role-Based Access Control**: Fine-grained permission system
- **Infrastructure Management**: Automated DynamoDB management
- **Real-time Monitoring**: Worker status and health monitoring

### API Features
- 🔐 **JWT Authentication**: Secure token-based authentication
- 🛡️ **Permission System**: 8-level permission model with role-based access
- 📊 **Structured Responses**: Consistent JSON response format
- 🔄 **Auto-Discovery**: Smart permission detection based on HTTP methods
- 📝 **Interactive Documentation**: Swagger UI with built-in authentication
- ⚡ **High Performance**: Optimized with caching and concurrent processing

## Base URL and Versioning

```
Base URL: https://your-domain.com/api/v1/auth
```

### Environment URLs
- **Development**: `http://localhost:8081/api/v1/auth`
- **Staging**: `https://staging-api.fieldfuze.com/api/v1/auth`
- **Production**: `https://api.fieldfuze.com/api/v1/auth`

### API Versioning
The API uses URL versioning with the format `/api/v{version}/`. Current version is `v1`.

## Authentication

### Authentication Flow

1. **User Registration**: Create user account (no token required)
2. **User Login**: Authenticate and receive JWT token
3. **Token Usage**: Include token in Authorization header for protected endpoints
4. **Token Refresh**: Login again when token expires (30 minutes default)

### JWT Token Structure

```json
{
  "user_id": "user-123",
  "email": "user@example.com",
  "username": "john_doe",
  "role": "user",
  "status": "active",
  "roles": [
    {
      "role_name": "admin",
      "level": 8,
      "permissions": ["read", "write", "delete", "admin"],
      "context": {
        "department": "engineering"
      }
    }
  ],
  "context": {
    "organization_id": "org-123",
    "customer_id": "cust-456"
  },
  "exp": 1640995200,
  "iat": 1640993400
}
```

### Authorization Header Format

```http
Authorization: Bearer YOUR_JWT_TOKEN
```

### Permission System

The API uses an 8-level permission system:

| Permission | Level | Description | HTTP Methods |
|------------|-------|-------------|--------------|
| `view` | 1 | Read-only access, reports | GET (read-only) |
| `read` | 2 | Data retrieval | GET |
| `write` | 3 | Data modification | PATCH |
| `create` | 4 | Data creation | POST |
| `update` | 5 | Data updates | PUT |
| `delete` | 6 | Data deletion | DELETE |
| `manage` | 7 | Create, update, delete operations | POST, PUT, DELETE |
| `admin` | 8 | Full access (highest level) | ALL |

## Response Format

### Standard Response Structure

All API responses follow a consistent structure:

```json
{
  "status": "success|error",
  "code": 200,
  "message": "Human-readable message",
  "data": {},
  "error": {
    "type": "ErrorType",
    "details": "Detailed error description",
    "field": "field_name"
  }
}
```

### Success Response Example

```json
{
  "status": "success",
  "code": 200,
  "message": "User retrieved successfully",
  "data": {
    "id": "user-123",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "status": "active",
    "created_at": "2024-01-15T10:30:45Z"
  }
}
```

### Paginated Response Example

```json
{
  "status": "success",
  "code": 200,
  "message": "Users retrieved successfully",
  "data": {
    "users": [...],
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 100,
      "total_pages": 10,
      "has_next": true,
      "has_previous": false
    }
  }
}
```

## Error Handling

### Error Response Structure

```json
{
  "status": "error",
  "code": 400,
  "message": "Validation failed",
  "error": {
    "type": "ValidationError",
    "details": "email is required; password must be at least 8 characters",
    "field": "email"
  }
}
```

### HTTP Status Codes

| Status Code | Description | Usage |
|-------------|-------------|--------|
| 200 | OK | Successful GET, PUT, PATCH |
| 201 | Created | Successful POST |
| 400 | Bad Request | Invalid request data |
| 401 | Unauthorized | Authentication required |
| 403 | Forbidden | Insufficient permissions |
| 404 | Not Found | Resource not found |
| 409 | Conflict | Resource already exists |
| 422 | Unprocessable Entity | Validation errors |
| 429 | Too Many Requests | Rate limit exceeded |
| 500 | Internal Server Error | Server error |

### Error Types

| Error Type | Description | Status Code |
|------------|-------------|-------------|
| `ValidationError` | Input validation failed | 400, 422 |
| `AuthenticationError` | Authentication required | 401 |
| `AuthorizationError` | Insufficient permissions | 403 |
| `NotFoundError` | Resource not found | 404 |
| `ConflictError` | Resource conflict | 409 |
| `DatabaseError` | Database operation failed | 500 |
| `ServerError` | Internal server error | 500 |
| `WorkerError` | Worker operation failed | 500 |

## User Management API

### Register User

Create a new user account.

**Endpoint**: `POST /user/register`  
**Authentication**: None required

#### Request Body

```json
{
  "email": "user@example.com",
  "username": "john_doe",
  "password": "securePassword123",
  "first_name": "John",
  "last_name": "Doe",
  "phone": "+1234567890",
  "company_name": "Acme Corp"
}
```

#### Response

```json
{
  "status": "success",
  "code": 201,
  "message": "User registered successfully",
  "data": {
    "id": "user-123",
    "email": "user@example.com",
    "username": "john_doe",
    "first_name": "John",
    "last_name": "Doe",
    "status": "active",
    "created_at": "2024-01-15T10:30:45Z"
  }
}
```

### User Login

Authenticate user and receive JWT token.

**Endpoint**: `POST /user/login`  
**Authentication**: None required

#### Request Body

```json
{
  "email": "user@example.com",
  "password": "securePassword123"
}
```

#### Response

```json
{
  "status": "success",
  "code": 200,
  "message": "Login successful",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 1800,
    "user": {
      "id": "user-123",
      "email": "user@example.com",
      "username": "john_doe",
      "first_name": "John",
      "last_name": "Doe",
      "roles": [
        {
          "role_name": "user",
          "level": 2,
          "permissions": ["read", "view"]
        }
      ]
    }
  }
}
```

### Get User Details

Retrieve user information by ID.

**Endpoint**: `GET /user/{id}`  
**Authentication**: Required  
**Permissions**: `user_details`

#### Response

```json
{
  "status": "success",
  "code": 200,
  "message": "User details retrieved successfully",
  "data": {
    "id": "user-123",
    "email": "user@example.com",
    "username": "john_doe",
    "first_name": "John",
    "last_name": "Doe",
    "status": "active",
    "roles": [
      {
        "role_name": "admin",
        "level": 8,
        "permissions": ["read", "write", "delete", "admin"],
        "assigned_at": "2024-01-15T10:30:45Z"
      }
    ],
    "last_login_at": "2024-01-15T10:30:45Z",
    "created_at": "2024-01-15T10:30:45Z",
    "updated_at": "2024-01-15T10:30:45Z"
  }
}
```

### List Users

Retrieve a paginated list of users.

**Endpoint**: `GET /user/list`  
**Authentication**: Required  
**Permissions**: `user_list`

#### Query Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `page` | integer | 1 | Page number |
| `limit` | integer | 10 | Items per page (max 100) |
| `sort` | string | asc | Sort order (asc/desc) |

#### Response

```json
{
  "status": "success",
  "code": 200,
  "message": "User list retrieved successfully",
  "data": {
    "users": [
      {
        "id": "user-123",
        "email": "user@example.com",
        "first_name": "John",
        "last_name": "Doe",
        "status": "active",
        "created_at": "2024-01-15T10:30:45Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 100,
      "total_pages": 10,
      "has_next": true,
      "has_previous": false
    }
  }
}
```

### Update User

Update user information.

**Endpoint**: `PATCH /user/update/{id}`  
**Authentication**: Required  
**Permissions**: `user_update`

#### Request Body

```json
{
  "first_name": "John Updated",
  "last_name": "Doe Updated",
  "phone": "+1234567890"
}
```

#### Response

```json
{
  "status": "success",
  "code": 200,
  "message": "User updated successfully",
  "data": {
    "id": "user-123",
    "email": "user@example.com",
    "first_name": "John Updated",
    "last_name": "Doe Updated",
    "phone": "+1234567890",
    "updated_at": "2024-01-15T10:30:45Z"
  }
}
```

### User Logout

Logout user and revoke JWT token.

**Endpoint**: `POST /user/logout`  
**Authentication**: Required

#### Response

```json
{
  "status": "success",
  "code": 200,
  "message": "Logout successful",
  "data": {
    "logged_out_at": "2024-01-15T10:30:45Z",
    "user_id": "user-123"
  }
}
```

### Validate Token

Validate a JWT token and return user information.

**Endpoint**: `POST /user/validate`  
**Authentication**: None required

#### Request Body

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

#### Response

```json
{
  "status": "success",
  "code": 200,
  "message": "Token is valid",
  "data": {
    "valid": true,
    "user": {
      "id": "user-123",
      "email": "user@example.com",
      "username": "john_doe",
      "roles": [
        {
          "role_name": "admin",
          "level": 8,
          "permissions": ["read", "write", "delete", "admin"]
        }
      ]
    },
    "expires_at": "2024-01-15T11:00:45Z"
  }
}
```

## Role Management API

### List Roles

Retrieve a list of roles with pagination.

**Endpoint**: `GET /user/role`  
**Authentication**: Required  
**Permissions**: `role_list`

#### Query Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `page` | integer | 1 | Page number |
| `limit` | integer | 10 | Items per page (max 100) |
| `status` | string | all | Filter by status (active/inactive/archived) |

#### Response

```json
{
  "status": "success",
  "code": 200,
  "message": "Roles retrieved successfully",
  "data": {
    "roles": [
      {
        "role_id": "role-123",
        "role_name": "Administrator",
        "level": 8,
        "permissions": ["read", "write", "delete", "admin"],
        "context": {
          "department": "engineering"
        },
        "assigned_at": "2024-01-15T10:30:45Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 25,
      "total_pages": 3,
      "has_next": true,
      "has_previous": false
    }
  }
}
```

### Create Role

Create a new role assignment.

**Endpoint**: `POST /user/role`  
**Authentication**: Required  
**Permissions**: `role_create`

#### Request Body

```json
{
  "role_name": "Content Manager",
  "level": 5,
  "permissions": ["read", "write", "create", "update"],
  "context": {
    "department": "marketing",
    "team": "content"
  }
}
```

#### Response

```json
{
  "status": "success",
  "code": 201,
  "message": "Role created successfully",
  "data": {
    "role_id": "role-456",
    "role_name": "Content Manager",
    "level": 5,
    "permissions": ["read", "write", "create", "update"],
    "context": {
      "department": "marketing",
      "team": "content"
    },
    "assigned_at": "2024-01-15T10:30:45Z"
  }
}
```

### Get Role Details

Retrieve role details by ID.

**Endpoint**: `GET /user/role/{id}`  
**Authentication**: Required  
**Permissions**: `role_list`

#### Response

```json
{
  "status": "success",
  "code": 200,
  "message": "Role details retrieved successfully",
  "data": {
    "role_id": "role-123",
    "role_name": "Administrator",
    "level": 8,
    "permissions": ["read", "write", "delete", "admin"],
    "context": {
      "department": "engineering"
    },
    "assigned_at": "2024-01-15T10:30:45Z"
  }
}
```

### Update Role

Update an existing role.

**Endpoint**: `PUT /user/role/{id}`  
**Authentication**: Required  
**Permissions**: `role_update`

#### Request Body

```json
{
  "role_name": "Senior Administrator",
  "level": 9,
  "permissions": ["read", "write", "delete", "admin", "manage"],
  "context": {
    "department": "engineering",
    "seniority": "senior"
  }
}
```

#### Response

```json
{
  "status": "success",
  "code": 200,
  "message": "Role updated successfully",
  "data": {
    "role_id": "role-123",
    "role_name": "Senior Administrator",
    "level": 9,
    "permissions": ["read", "write", "delete", "admin", "manage"],
    "context": {
      "department": "engineering",
      "seniority": "senior"
    },
    "updated_at": "2024-01-15T10:30:45Z"
  }
}
```

### Delete Role

Delete a role by ID.

**Endpoint**: `DELETE /user/role/{id}`  
**Authentication**: Required  
**Permissions**: `role_delete`

#### Response

```json
{
  "status": "success",
  "code": 200,
  "message": "Role deleted successfully",
  "data": {
    "deleted_role_id": "role-123"
  }
}
```

### Assign Role to User

Assign an existing role to a user.

**Endpoint**: `POST /user/{user_id}/role/{role_id}`  
**Authentication**: Required  
**Permissions**: `role_assign`

#### Response

```json
{
  "status": "success",
  "code": 200,
  "message": "Role assigned successfully",
  "data": {
    "id": "user-123",
    "email": "user@example.com",
    "roles": [
      {
        "role_id": "role-123",
        "role_name": "Administrator",
        "level": 8,
        "permissions": ["read", "write", "delete", "admin"],
        "assigned_at": "2024-01-15T10:30:45Z"
      }
    ]
  }
}
```

### Remove Role from User

Remove a role from a user.

**Endpoint**: `DELETE /user/{user_id}/role/{role_id}`  
**Authentication**: Required  
**Permissions**: `role_assign`

#### Response

```json
{
  "status": "success",
  "code": 200,
  "message": "Role removed successfully",
  "data": {
    "id": "user-123",
    "email": "user@example.com",
    "roles": [
      // Remaining roles after removal
    ]
  }
}
```

## Infrastructure API

### Get Worker Status

Get detailed status of the infrastructure worker.

**Endpoint**: `GET /infrastructure/worker/status`  
**Authentication**: Required  
**Permissions**: `admin`

#### Response

```json
{
  "status": "success",
  "code": 200,
  "message": "Infrastructure is ready and healthy",
  "data": {
    "success": true,
    "status": "completed",
    "phase": "validation",
    "start_time": "2024-01-15T10:30:45Z",
    "end_time": "2024-01-15T10:32:15Z",
    "duration": "1m30s",
    "progress": {
      "current_step": 5,
      "total_steps": 5,
      "step_name": "Infrastructure validation",
      "percentage": 100
    },
    "tables_created": [
      {
        "name": "users",
        "status": "ACTIVE",
        "created_at": "2024-01-15T10:30:45Z",
        "became_active_at": "2024-01-15T10:31:15Z",
        "index_count": 2,
        "expected_indexes": 2,
        "billing_mode": "PAY_PER_REQUEST"
      }
    ],
    "indexes_created": [
      {
        "name": "email-index",
        "status": "ACTIVE",
        "created_at": "2024-01-15T10:31:00Z"
      }
    ],
    "environment": "development",
    "health_status": "healthy",
    "next_action": "Continue monitoring",
    "metadata": {
      "worker_version": "1.0.0",
      "last_validation": "2024-01-15T10:32:15Z"
    }
  }
}
```

### Check Worker Health

Check if the infrastructure worker is healthy.

**Endpoint**: `GET /infrastructure/worker/health`  
**Authentication**: Required  
**Permissions**: `admin`

#### Response

```json
{
  "status": "success",
  "code": 200,
  "message": "Worker health check completed",
  "data": {
    "healthy": true,
    "status": "healthy",
    "reason": "Worker completed successfully"
  }
}
```

### Restart Worker

Restart the infrastructure worker.

**Endpoint**: `POST /infrastructure/worker/restart`  
**Authentication**: Required  
**Permissions**: `admin`

#### Request Body

```json
{
  "force": false
}
```

#### Response

```json
{
  "status": "success",
  "code": 200,
  "message": "Worker restart initiated successfully",
  "data": {
    "service_name": "infrastructure-worker",
    "status": "completed",
    "start_time": "2024-01-15T10:30:45Z",
    "end_time": "2024-01-15T10:31:00Z",
    "output": "Worker restarted successfully"
  }
}
```

### Auto-Restart Worker

Check worker health and restart if needed.

**Endpoint**: `POST /infrastructure/worker/auto-restart`  
**Authentication**: Required  
**Permissions**: `admin`

#### Response

```json
{
  "status": "success",
  "code": 200,
  "message": "Worker is healthy, no restart needed",
  "data": {
    "checked_at": "2024-01-15T10:30:45Z",
    "was_healthy": true,
    "health_reason": "Worker completed successfully",
    "status": "not_needed"
  }
}
```

## Rate Limiting

The API implements rate limiting to prevent abuse:

- **Default Limit**: 100 requests per minute per IP
- **Rate Limit Headers**: Included in responses
- **Rate Limit Exceeded**: Returns HTTP 429 status

### Rate Limit Headers

```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1640995200
```

### Rate Limit Exceeded Response

```json
{
  "status": "error",
  "code": 429,
  "message": "Rate limit exceeded",
  "error": {
    "type": "RateLimitError",
    "details": "Too many requests. Please try again later.",
    "retry_after": 60
  }
}
```

## SDKs and Tools

### Interactive Documentation

Access the interactive Swagger UI documentation:

```
https://your-domain.com/swagger
```

Features:
- **Try It Out**: Test endpoints directly from the browser
- **Built-in Authentication**: Login form integrated into the UI
- **Auto Token Management**: Automatic Bearer token insertion
- **Request/Response Examples**: Complete examples for all endpoints

### cURL Examples

#### Register User
```bash
curl -X POST "https://your-domain.com/api/v1/auth/user/register" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "username": "john_doe",
    "password": "securePassword123",
    "first_name": "John",
    "last_name": "Doe"
  }'
```

#### Login
```bash
curl -X POST "https://your-domain.com/api/v1/auth/user/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securePassword123"
  }'
```

#### Get User (with authentication)
```bash
curl -X GET "https://your-domain.com/api/v1/auth/user/user-123" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### JavaScript/Node.js SDK

```javascript
class FieldFuzeAPI {
  constructor(baseURL, token = null) {
    this.baseURL = baseURL;
    this.token = token;
  }

  async request(method, endpoint, data = null) {
    const headers = {
      'Content-Type': 'application/json',
    };

    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`;
    }

    const config = {
      method,
      headers,
    };

    if (data) {
      config.body = JSON.stringify(data);
    }

    const response = await fetch(`${this.baseURL}${endpoint}`, config);
    return await response.json();
  }

  async login(email, password) {
    const response = await this.request('POST', '/user/login', {
      email,
      password,
    });
    
    if (response.status === 'success') {
      this.token = response.data.access_token;
    }
    
    return response;
  }

  async getUser(userId) {
    return await this.request('GET', `/user/${userId}`);
  }

  async createRole(roleData) {
    return await this.request('POST', '/user/role', roleData);
  }

  async getWorkerStatus() {
    return await this.request('GET', '/infrastructure/worker/status');
  }
}

// Usage
const api = new FieldFuzeAPI('https://your-domain.com/api/v1/auth');

// Login
const loginResult = await api.login('user@example.com', 'password123');
console.log('Login result:', loginResult);

// Get user
const user = await api.getUser('user-123');
console.log('User:', user);
```

### Python SDK

```python
import requests
import json

class FieldFuzeAPI:
    def __init__(self, base_url, token=None):
        self.base_url = base_url
        self.token = token
        self.session = requests.Session()

    def _request(self, method, endpoint, data=None):
        headers = {'Content-Type': 'application/json'}
        
        if self.token:
            headers['Authorization'] = f'Bearer {self.token}'
        
        url = f"{self.base_url}{endpoint}"
        
        if data:
            response = self.session.request(method, url, headers=headers, json=data)
        else:
            response = self.session.request(method, url, headers=headers)
        
        return response.json()

    def login(self, email, password):
        response = self._request('POST', '/user/login', {
            'email': email,
            'password': password
        })
        
        if response.get('status') == 'success':
            self.token = response['data']['access_token']
        
        return response

    def get_user(self, user_id):
        return self._request('GET', f'/user/{user_id}')

    def create_role(self, role_data):
        return self._request('POST', '/user/role', role_data)

    def get_worker_status(self):
        return self._request('GET', '/infrastructure/worker/status')

# Usage
api = FieldFuzeAPI('https://your-domain.com/api/v1/auth')

# Login
login_result = api.login('user@example.com', 'password123')
print('Login result:', login_result)

# Get user
user = api.get_user('user-123')
print('User:', user)
```

---

**Related Documentation**: [Authentication](AUTHENTICATION.md) | [Controllers](CONTROLLERS.md) | [Models](MODELS.md) | [Development Guide](DEVELOPMENT.md)