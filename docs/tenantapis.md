# Tenant APIs Documentation

This document provides comprehensive documentation for all tenant-related APIs. These APIs allow tenants to manage their accounts, view their profiles, manage passwords, and access their sync tokens.

**Base URL:** `http://localhost:8081/api/v1` (or your deployed server)

**Authentication:** Most tenant endpoints require a valid JWT access token with tenant role. Include the token in the `Authorization` header as `Bearer <token>`.

---

## Table of Contents

1. [Authentication](#authentication)
2. [Tenant Profile Management](#tenant-profile-management)
3. [Password Management](#password-management)
4. [Sync Token Access](#sync-token-access)
5. [Response Format](#response-format)
6. [Error Handling](#error-handling)
7. [Rate Limiting](#rate-limiting)

---

## Authentication

### Tenant Login

Creates a new session for a tenant user and returns access/refresh tokens. Tenants must use their email and password to authenticate.

**Endpoint:** `POST /auth/tenant/login`

**Authentication:** None (public endpoint)

**Request Body:**
```json
{
  "email": "tenant@example.com",
  "password": "your_password"
}
```

**Parameters:**
- `email` (string, required): Tenant user's email address
- `password` (string, required): Tenant user's password

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 3600,
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "email": "tenant@example.com",
      "full_name": "John Doe",
      "company_name": "Acme Corp",
      "user_type": "tenant",
      "must_change_password": false
    }
  }
}
```

**Response (401 Unauthorized):**
```json
{
  "success": false,
  "error": "Invalid credentials"
}
```

**Errors:**
- `400 Bad Request`: Invalid email or password format
- `401 Unauthorized`: Invalid credentials or account not found
- `403 Forbidden`: Tenant account is disabled

---

### Refresh Token

Generates a new access token using a refresh token. This can be used by both admin and tenant users.

**Endpoint:** `POST /auth/refresh`

**Authentication:** None (public endpoint)

**Request Body:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Parameters:**
- `refresh_token` (string, required): The refresh token from login

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Token refreshed",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 3600
  }
}
```

**Errors:**
- `400 Bad Request`: Invalid refresh token format
- `401 Unauthorized`: Refresh token is invalid or expired

---

### Logout

Invalidates the current session and revokes the refresh token.

**Endpoint:** `POST /auth/logout`

**Authentication:** Required (Bearer token)

**Request Body:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Parameters:**
- `refresh_token` (string, required): The refresh token to revoke

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Logout successful"
}
```

**Errors:**
- `401 Unauthorized`: Invalid or missing access token
- `400 Bad Request`: Invalid request body

---

## Tenant Profile Management

### Get Tenant Profile

Retrieves the current tenant's profile information. This endpoint returns detailed information about the authenticated tenant user.

**Endpoint:** `GET /tenant/profile`

**Authentication:** Required (Bearer token with tenant role)

**Path Parameters:** None

**Query Parameters:** None

**Request Body:** None

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Profile retrieved",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440001",
    "email": "tenant@example.com",
    "full_name": "John Doe",
    "company_name": "Acme Corp",
    "password_changed_at": "2026-01-07T10:30:00Z",
    "password_expires_at": "2026-04-07T10:30:00Z",
    "must_change_password": false,
    "is_active": true,
    "disabled_at": null,
    "disabled_reason": null,
    "s3_prefix": "tenants/550e8400-e29b-41d4-a716-446655440001/",
    "created_by": "550e8400-e29b-41d4-a716-446655440000",
    "created_at": "2026-01-05T14:20:00Z",
    "updated_at": "2026-01-07T10:30:00Z",
    "last_login_at": "2026-01-07T10:30:00Z",
    "last_login_ip": "192.168.1.100"
  }
}
```

**Field Descriptions:**
- `id` (UUID): Unique identifier for the tenant
- `email` (string): Tenant's email address
- `full_name` (string): Tenant's full name
- `company_name` (string, optional): Company name associated with the tenant
- `password_changed_at` (timestamp): When the password was last changed
- `password_expires_at` (timestamp, optional): When the password will expire
- `must_change_password` (boolean): Whether the tenant must change password on next login
- `is_active` (boolean): Whether the tenant account is active
- `disabled_at` (timestamp, optional): When the account was disabled
- `disabled_reason` (string, optional): Reason for account disabling
- `s3_prefix` (string): S3 storage prefix for this tenant's files
- `created_by` (UUID): ID of the admin who created this tenant
- `created_at` (timestamp): Account creation timestamp
- `updated_at` (timestamp): Last update timestamp
- `last_login_at` (timestamp, optional): Last successful login timestamp
- `last_login_ip` (string, optional): IP address of last login

**Errors:**
- `401 Unauthorized`: Invalid or missing access token
- `403 Forbidden`: User is not a tenant or account is disabled
- `404 Not Found`: Tenant not found

---

## Password Management

### Change Password

Allows a tenant to change their password. The current password must be provided for verification.

**Endpoint:** `POST /tenant/change-password`

**Authentication:** Required (Bearer token with tenant role)

**Request Body:**
```json
{
  "current_password": "old_password_123",
  "new_password": "new_secure_password_456"
}
```

**Parameters:**
- `current_password` (string, required): The tenant's current password
- `new_password` (string, required, minimum 8 characters): The new password to set

**Password Requirements:**
- Minimum 8 characters
- Must contain uppercase and lowercase letters
- Must contain at least one number
- Must contain at least one special character (!@#$%^&*)

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Password changed successfully"
}
```

**Response (400 Bad Request) - Weak Password:**
```json
{
  "success": false,
  "error": "Password must contain at least one uppercase letter, one lowercase letter, one number, and one special character"
}
```

**Response (401 Unauthorized) - Invalid Current Password:**
```json
{
  "success": false,
  "error": "Invalid current password"
}
```

**Errors:**
- `400 Bad Request`: Invalid request body or password doesn't meet requirements
- `401 Unauthorized`: Invalid or missing access token, or invalid current password
- `403 Forbidden`: User is not a tenant
- `404 Not Found`: Tenant not found

**Side Effects:**
- All existing refresh tokens are revoked for security
- Audit log entry is created for this action
- Password change timestamp is updated

---

## Sync Token Access

Tenants can view their own sync tokens and their statistics. These endpoints provide read-only access to sync tokens created for the tenant.

### List Own Sync Tokens

Retrieves all sync tokens associated with the current tenant. Sync tokens are used for programmatic API access.

**Endpoint:** `GET /tenant/sync-tokens`

**Authentication:** Required (Bearer token with tenant role)

**Query Parameters:**
- `limit` (integer, optional): Maximum number of records to return (default: 10, max: 100)
- `offset` (integer, optional): Number of records to skip (default: 0)
- `sort` (string, optional): Sort field (default: `-created_at`). Use `-` prefix for descending order.

**Request Body:** None

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Sync tokens retrieved",
  "data": [
    {
      "id": "660e8400-e29b-41d4-a716-446655440001",
      "tenant_id": "550e8400-e29b-41d4-a716-446655440001",
      "name": "Mobile App Token",
      "token": "st_live_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
      "token_hash": "9f86d081884c7d6d9ffd60bb51d3b3a1d4e6e7c...",
      "permissions": ["read", "write"],
      "last_used_at": "2026-01-07T09:45:00Z",
      "expires_at": "2026-12-31T23:59:59Z",
      "is_revoked": false,
      "revoked_at": null,
      "created_at": "2026-01-05T14:20:00Z",
      "updated_at": "2026-01-07T09:45:00Z"
    },
    {
      "id": "660e8400-e29b-41d4-a716-446655440002",
      "tenant_id": "550e8400-e29b-41d4-a716-446655440001",
      "name": "Desktop App Token",
      "token": "st_live_yyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy",
      "token_hash": "8f96d081884c7d6d9ffd60bb51d3b3a1d4e6e7b...",
      "permissions": ["read"],
      "last_used_at": "2026-01-07T08:20:00Z",
      "expires_at": "2026-06-30T23:59:59Z",
      "is_revoked": false,
      "revoked_at": null,
      "created_at": "2026-01-03T10:15:00Z",
      "updated_at": "2026-01-07T08:20:00Z"
    }
  ],
  "metadata": {
    "total": 2,
    "limit": 10,
    "offset": 0
  }
}
```

**Field Descriptions:**
- `id` (UUID): Unique identifier for the sync token
- `tenant_id` (UUID): ID of the tenant who owns this token
- `name` (string): Human-readable name for the token
- `token` (string): The actual token value (only shown in list view)
- `token_hash` (string): Hash of the token for security
- `permissions` (array): List of permissions granted to this token (read, write, delete)
- `last_used_at` (timestamp, optional): When the token was last used for API access
- `expires_at` (timestamp, optional): When the token expires
- `is_revoked` (boolean): Whether the token has been revoked
- `revoked_at` (timestamp, optional): When the token was revoked
- `created_at` (timestamp): Token creation timestamp
- `updated_at` (timestamp): Last update timestamp

**Errors:**
- `401 Unauthorized`: Invalid or missing access token
- `403 Forbidden`: User is not a tenant

---

### Get Own Sync Token Details

Retrieves detailed information about a specific sync token owned by the current tenant.

**Endpoint:** `GET /tenant/sync-tokens/:id`

**Authentication:** Required (Bearer token with tenant role)

**Path Parameters:**
- `id` (UUID, required): The ID of the sync token to retrieve

**Query Parameters:** None

**Request Body:** None

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Sync token retrieved",
  "data": {
    "id": "660e8400-e29b-41d4-a716-446655440001",
    "tenant_id": "550e8400-e29b-41d4-a716-446655440001",
    "name": "Mobile App Token",
    "token_hash": "9f86d081884c7d6d9ffd60bb51d3b3a1d4e6e7c...",
    "permissions": ["read", "write"],
    "last_used_at": "2026-01-07T09:45:00Z",
    "expires_at": "2026-12-31T23:59:59Z",
    "is_revoked": false,
    "revoked_at": null,
    "created_at": "2026-01-05T14:20:00Z",
    "updated_at": "2026-01-07T09:45:00Z"
  }
}
```

**Errors:**
- `401 Unauthorized`: Invalid or missing access token
- `403 Forbidden`: User is not a tenant or does not own this token
- `404 Not Found`: Sync token not found

---

### Get Own Sync Token Statistics

Retrieves usage statistics for a specific sync token owned by the current tenant. This includes API call counts, bandwidth usage, and error rates.

**Endpoint:** `GET /tenant/sync-tokens/:id/stats`

**Authentication:** Required (Bearer token with tenant role)

**Path Parameters:**
- `id` (UUID, required): The ID of the sync token

**Query Parameters:**
- `period` (string, optional): Time period for statistics (default: `7d`). Options: `1d`, `7d`, `30d`, `90d`, `1y`, `all`
- `granularity` (string, optional): Granularity level (default: `hourly`). Options: `hourly`, `daily`, `weekly`, `monthly`

**Request Body:** None

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Sync token stats retrieved",
  "data": {
    "id": "660e8400-e29b-41d4-a716-446655440001",
    "name": "Mobile App Token",
    "period": "7d",
    "total_requests": 15847,
    "successful_requests": 15821,
    "failed_requests": 26,
    "total_bandwidth_bytes": 52428800,
    "total_errors": 26,
    "top_errors": [
      {
        "error_code": "AUTH_EXPIRED",
        "count": 15,
        "percentage": 57.69
      },
      {
        "error_code": "RATE_LIMIT_EXCEEDED",
        "count": 8,
        "percentage": 30.77
      },
      {
        "error_code": "INVALID_REQUEST",
        "count": 3,
        "percentage": 11.54
      }
    ],
    "requests_by_endpoint": [
      {
        "endpoint": "POST /files/upload-url",
        "count": 8920,
        "success_rate": 99.98
      },
      {
        "endpoint": "GET /files",
        "count": 6234,
        "success_rate": 99.95
      },
      {
        "endpoint": "POST /files/complete-upload",
        "count": 693,
        "success_rate": 100.0
      }
    ],
    "requests_by_ip": [
      {
        "ip_address": "192.168.1.100",
        "count": 10234,
        "success_rate": 99.96
      },
      {
        "ip_address": "192.168.1.101",
        "count": 5613,
        "success_rate": 99.94
      }
    ],
    "first_used_at": "2026-01-05T14:20:00Z",
    "last_used_at": "2026-01-07T09:45:00Z",
    "daily_stats": [
      {
        "date": "2026-01-07",
        "requests": 2847,
        "successful_requests": 2842,
        "failed_requests": 5,
        "bandwidth_bytes": 8946560
      },
      {
        "date": "2026-01-06",
        "requests": 2456,
        "successful_requests": 2451,
        "failed_requests": 5,
        "bandwidth_bytes": 7874560
      }
    ]
  }
}
```

**Query Response Fields:**
- `total_requests` (integer): Total number of API requests made with this token
- `successful_requests` (integer): Number of successful requests
- `failed_requests` (integer): Number of failed requests
- `total_bandwidth_bytes` (integer): Total bytes transferred
- `success_rate` (float): Percentage of successful requests
- `top_errors` (array): Most common errors encountered
- `requests_by_endpoint` (array): Breakdown of requests by API endpoint
- `requests_by_ip` (array): Breakdown of requests by source IP address
- `daily_stats` (array): Daily statistics for the selected period

**Errors:**
- `400 Bad Request`: Invalid period or granularity parameter
- `401 Unauthorized`: Invalid or missing access token
- `403 Forbidden`: User is not a tenant or does not own this token
- `404 Not Found`: Sync token not found

---

## Response Format

### Success Response

All successful API responses follow this format:

```json
{
  "success": true,
  "message": "Descriptive message about the response",
  "data": {}
}
```

- `success` (boolean): Always `true` for successful responses
- `message` (string): Human-readable message describing the result
- `data` (object/array): The actual response data (structure varies by endpoint)

### Error Response

All error responses follow this format:

```json
{
  "success": false,
  "error": "Error description"
}
```

- `success` (boolean): Always `false` for error responses
- `error` (string): Human-readable error message

---

## Error Handling

The API uses standard HTTP status codes to indicate the result of API requests:

| Status Code | Description |
|------------|-------------|
| 200 OK | Request succeeded |
| 400 Bad Request | Invalid request parameters or body |
| 401 Unauthorized | Missing or invalid authentication token |
| 403 Forbidden | Authenticated but not authorized for this resource |
| 404 Not Found | Resource not found |
| 409 Conflict | Resource conflict (e.g., duplicate email) |
| 429 Too Many Requests | Rate limit exceeded |
| 500 Internal Server Error | Server error |
| 503 Service Unavailable | Service temporarily unavailable |

### Common Error Codes

| Error Code | HTTP Status | Description |
|-----------|------------|-------------|
| `ErrBadRequest` | 400 | Invalid request format or parameters |
| `ErrUnauthorized` | 401 | Authentication failed or missing |
| `ErrForbidden` | 403 | Not authorized to access this resource |
| `ErrNotFound` | 404 | Resource not found |
| `ErrConflict` | 409 | Resource already exists or conflict |
| `ErrInternalServer` | 500 | Unexpected server error |
| `ErrInvalidCredentials` | 401 | Email/password combination is incorrect |
| `ErrDisabled` | 403 | Account is disabled |
| `ErrMustChangePassword` | 403 | Account must change password before proceeding |

---

## Rate Limiting

### Rate Limit Headers

All API responses include rate limit information in the headers:

```
X-RateLimit-Limit: 10
X-RateLimit-Remaining: 9
X-RateLimit-Reset: 1641564000
```

- `X-RateLimit-Limit`: Maximum requests allowed per time window
- `X-RateLimit-Remaining`: Number of requests remaining in current window
- `X-RateLimit-Reset`: Unix timestamp when the rate limit resets

### Rate Limit Policies

**Tenant Endpoints:**
- Default: 10 requests per second (burst of 20 allowed)
- File operations: 10 requests per second (burst of 20 allowed)
- Authentication: 5 requests per second (burst of 10 allowed)

**Exceeding Rate Limits:**

When rate limit is exceeded, the API returns:

```json
{
  "success": false,
  "error": "Too many requests. Please try again later."
}
```

HTTP Status Code: `429 Too Many Requests`

Headers:
```
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1641564030
Retry-After: 30
```

---

## Example Usage

### Complete Authentication Flow

```bash
# 1. Login as tenant
curl -X POST http://localhost:8080/api/v1/auth/tenant/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "tenant@example.com",
    "password": "password123"
  }'

# Response includes access_token and refresh_token
# {
#   "success": true,
#   "data": {
#     "access_token": "eyJhbGciOiJIUzI1NiIs...",
#     "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
#   }
# }

# 2. Get profile using access token
curl -X GET http://localhost:8080/api/v1/tenant/profile \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."

# 3. Change password
curl -X POST http://localhost:8080/api/v1/tenant/change-password \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  -H "Content-Type: application/json" \
  -d '{
    "current_password": "password123",
    "new_password": "NewPassword@123"
  }'

# 4. List sync tokens
curl -X GET http://localhost:8080/api/v1/tenant/sync-tokens \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."

# 5. Get sync token stats
curl -X GET "http://localhost:8080/api/v1/tenant/sync-tokens/660e8400-e29b-41d4-a716-446655440001/stats?period=7d" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."
```

### Using Refresh Token to Get New Access Token

```bash
# Refresh access token
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
  }'
```

### Logout

```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
  }'
```

---

## Security Considerations

1. **Token Storage:** Store tokens securely. Never expose tokens in logs or client-side code.
2. **HTTPS Only:** Always use HTTPS in production to prevent token interception.
3. **Token Expiration:** Access tokens expire after a set period. Use refresh tokens to get new access tokens.
4. **Password Security:** Passwords must meet minimum strength requirements. Change passwords regularly.
5. **IP Whitelisting:** Consider restricting API access to known IP addresses.
6. **Audit Logs:** Monitor audit logs for suspicious activity.
7. **Rate Limiting:** Respect rate limits to avoid service disruption.

---

## SDK/Library Support

For easier integration with tenant APIs, consider using the following SDKs:

- **JavaScript/Node.js:** Available in npm registry
- **Python:** Available in PyPI
- **Go:** Available as Go module
- **Java:** Available in Maven Central

---

## Support and Resources

For additional help:

- **Documentation:** See [README.md](README.md)
- **Issues:** Report bugs on GitHub Issues
- **API Status:** Check API status at [status page](https://status.example.com)
- **Contact:** support@example.com

---

**Last Updated:** January 7, 2026
**API Version:** v1
**Documentation Version:** 1.0
