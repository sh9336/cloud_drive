# Admin APIs Documentation

This document provides comprehensive documentation for all admin-related APIs. These APIs are used to manage tenants and sync tokens in the system.

**Base URL:** `http://localhost:8080/api/v1` (or your deployed server)

**Authentication:** All admin endpoints require a valid JWT access token with admin role. Include the token in the `Authorization` header as `Bearer <token>`.

---

## Table of Contents

1. [Authentication](#authentication)
2. [Admin Account Management](#admin-account-management)
3. [Tenant Management](#tenant-management)
4. [Sync Token Management](#sync-token-management)
5. [Response Format](#response-format)
6. [Error Handling](#error-handling)
7. [Rate Limiting](#rate-limiting)

---

## Authentication

### Admin Login

Creates a new session for an admin user and returns access/refresh tokens.

**Endpoint:** `POST /auth/admin/login`

**Authentication:** None (public endpoint)

**Request Body:**
```json
{
  "email": "admin@example.com",
  "password": "your_password"
}
```

**Parameters:**
- `email` (string, required): Admin user's email address
- `password` (string, required): Admin user's password

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "token_type": "Bearer",
    "expires_in": 3600,
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "email": "admin@example.com",
      "full_name": "Admin User",
      "user_type": "admin",
      "must_change_password": false
    }
  }
}
```

**Errors:**
- `400 Bad Request`: Invalid email or password format
- `401 Unauthorized`: Invalid credentials
- `403 Forbidden`: Admin account is disabled

---

### Refresh Token

Generates a new access token using a refresh token.

**Endpoint:** `POST /auth/refresh`

**Authentication:** None (public endpoint)

**Request Body:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
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
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "token_type": "Bearer",
    "expires_in": 3600
  }
}
```

**Errors:**
- `400 Bad Request`: Invalid refresh token format
- `401 Unauthorized`: Invalid or expired refresh token

---

### Logout

Invalidates the current refresh token and ends the session.

**Endpoint:** `POST /auth/logout`

**Authentication:** Required (Bearer token)

**Request Body:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**Parameters:**
- `refresh_token` (string, required): The refresh token to invalidate

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Logout successful",
  "data": null
}
```

**Errors:**
- `400 Bad Request`: Missing refresh token
- `401 Unauthorized`: Invalid token

---

## Admin Account Management

### Change Admin Password

Allows an authenticated admin to change their own password.

**Endpoint:** `POST /auth/admin/change_password`

**Authentication:** Required (Admin Bearer token)

**Request Body:**
```json
{
  "current_password": "old_password_123",
  "new_password": "new_secure_password_456"
}
```

**Parameters:**
- `current_password` (string, required): Admin's current password
- `new_password` (string, required, min 8 characters): New password to set

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Password changed successfully",
  "data": null
}
```

**Errors:**
- `400 Bad Request`: Invalid password format or new password too short
- `401 Unauthorized`: Current password is incorrect or invalid token
- `403 Forbidden`: Admin account is disabled

---

## Tenant Management

### Create Tenant

Creates a new tenant account with a temporary password.

**Endpoint:** `POST /admin/tenants`

**Authentication:** Required (Admin Bearer token)

**Request Body:**
```json
{
  "email": "tenant@example.com",
  "full_name": "John Doe",
  "company_name": "Acme Corp"
}
```

**Parameters:**
- `email` (string, required, email format): Tenant's email address
- `full_name` (string, required): Tenant's full name
- `company_name` (string, optional): Company/organization name

**Response (201 Created):**
```json
{
  "success": true,
  "message": "Tenant created successfully",
  "data": {
    "tenant": {
      "id": "660e8400-e29b-41d4-a716-446655440000",
      "email": "tenant@example.com",
      "full_name": "John Doe",
      "company_name": "Acme Corp",
      "is_active": true,
      "must_change_password": true,
      "s3_prefix": "tenants/550e8400-e29b-41d4-a716-446655440001/",
      "created_by": "550e8400-e29b-41d4-a716-446655440000",
      "created_at": "2024-01-07T10:30:00Z",
      "updated_at": "2024-01-07T10:30:00Z",
      "password_changed_at": "2024-01-07T10:30:00Z",
      "last_login_at": null,
      "last_login_ip": null
    },
    "temporary_password": "TempPass123!@#"
  }
}
```

**Errors:**
- `400 Bad Request`: Invalid email format or missing required fields
- `409 Conflict`: Email already exists
- `401 Unauthorized`: Invalid admin token
- `403 Forbidden`: User is not an admin

---

### List Tenants

Retrieves a list of all tenants in the system.

**Endpoint:** `GET /admin/tenants`

**Authentication:** Required (Admin Bearer token)

**Query Parameters:** None

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Tenants retrieved",
  "data": [
    {
      "id": "660e8400-e29b-41d4-a716-446655440000",
      "email": "tenant1@example.com",
      "full_name": "John Doe",
      "company_name": "Acme Corp",
      "is_active": true,
      "must_change_password": false,
      "s3_prefix": "tenants/550e8400-e29b-41d4-a716-446655440001/",
      "created_by": "550e8400-e29b-41d4-a716-446655440000",
      "created_at": "2024-01-07T10:30:00Z",
      "updated_at": "2024-01-07T10:30:00Z",
      "password_changed_at": "2024-01-07T10:30:00Z",
      "disabled_at": null
    },
    {
      "id": "770e8400-e29b-41d4-a716-446655440000",
      "email": "tenant2@example.com",
      "full_name": "Jane Smith",
      "company_name": "TechCorp",
      "is_active": true,
      "must_change_password": true,
      "s3_prefix": "tenants/550e8400-e29b-41d4-a716-446655440002/",
      "created_by": "550e8400-e29b-41d4-a716-446655440000",
      "created_at": "2024-01-06T14:20:00Z",
      "updated_at": "2024-01-06T14:20:00Z",
      "password_changed_at": "2024-01-06T14:20:00Z",
      "disabled_at": null
    }
  ]
}
```

**Errors:**
- `401 Unauthorized`: Invalid or expired token
- `403 Forbidden`: User is not an admin

---

### Get Tenant Details

Retrieves details of a specific tenant.

**Endpoint:** `GET /admin/tenants/:id`

**Authentication:** Required (Admin Bearer token)

**Path Parameters:**
- `id` (string, required, UUID format): The tenant's ID

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Tenant retrieved",
  "data": {
    "id": "660e8400-e29b-41d4-a716-446655440000",
    "email": "tenant@example.com",
    "full_name": "John Doe",
    "company_name": "Acme Corp",
    "is_active": true,
    "must_change_password": false,
    "s3_prefix": "tenants/550e8400-e29b-41d4-a716-446655440001/",
    "created_by": "550e8400-e29b-41d4-a716-446655440000",
    "created_at": "2024-01-07T10:30:00Z",
    "updated_at": "2024-01-07T10:30:00Z",
    "password_changed_at": "2024-01-07T10:30:00Z",
    "last_login_at": "2024-01-07T11:15:00Z",
    "last_login_ip": "192.168.1.100"
  }
}
```

**Errors:**
- `400 Bad Request`: Invalid UUID format for ID
- `404 Not Found`: Tenant with given ID does not exist
- `401 Unauthorized`: Invalid or expired token
- `403 Forbidden`: User is not an admin

---

### Reset Tenant Password

Generates a new temporary password for a tenant. Revokes all existing tokens for that tenant.

**Endpoint:** `POST /admin/tenants/:id/reset-password`

**Authentication:** Required (Admin Bearer token)

**Path Parameters:**
- `id` (string, required, UUID format): The tenant's ID

**Request Body:** Empty

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Password reset successful",
  "data": {
    "temporary_password": "NewTempPass456!@#"
  }
}
```

**Errors:**
- `400 Bad Request`: Invalid UUID format for ID
- `404 Not Found`: Tenant with given ID does not exist
- `401 Unauthorized`: Invalid or expired token
- `403 Forbidden`: User is not an admin

---

### Update Tenant Status

Activates or deactivates a tenant account. When disabling, all tokens for that tenant are revoked.

**Endpoint:** `PATCH /admin/tenants/:id/status`

**Authentication:** Required (Admin Bearer token)

**Path Parameters:**
- `id` (string, required, UUID format): The tenant's ID

**Request Body:**
```json
{
  "is_active": true,
  "disabled_reason": "Account suspended due to policy violation"
}
```

**Parameters:**
- `is_active` (boolean, required): Whether the tenant account should be active
- `disabled_reason` (string, optional): Reason for disabling the account

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Tenant status updated",
  "data": null
}
```

**Errors:**
- `400 Bad Request`: Invalid UUID format or missing required fields
- `404 Not Found`: Tenant with given ID does not exist
- `401 Unauthorized`: Invalid or expired token
- `403 Forbidden`: User is not an admin

---

## Sync Token Management

### Create Sync Token

Creates a new sync token for a tenant with specified permissions.

**Endpoint:** `POST /admin/sync-tokens`

**Authentication:** Required (Admin Bearer token)

**Request Body:**
```json
{
  "tenant_id": "660e8400-e29b-41d4-a716-446655440000",
  "name": "Mobile App Token",
  "can_read": true,
  "can_write": true,
  "can_delete": false,
  "expires_in_days": 90
}
```

**Parameters:**
- `tenant_id` (string, required, UUID format): The tenant's ID
- `name` (string, required): Human-readable name for the token
- `can_read` (boolean, optional, default: false): Permission to read files
- `can_write` (boolean, optional, default: false): Permission to write/upload files
- `can_delete` (boolean, optional, default: false): Permission to delete files
- `expires_in_days` (integer, required): Token expiration period in days (1-365)

**Response (201 Created):**
```json
{
  "success": true,
  "message": "Sync token created successfully",
  "data": {
    "id": "880e8400-e29b-41d4-a716-446655440000",
    "tenant_id": "660e8400-e29b-41d4-a716-446655440000",
    "name": "Mobile App Token",
    "can_read": true,
    "can_write": true,
    "can_delete": false,
    "is_active": true,
    "created_by": "550e8400-e29b-41d4-a716-446655440000",
    "created_at": "2024-01-07T10:30:00Z",
    "updated_at": "2024-01-07T10:30:00Z",
    "expires_at": "2024-04-07T10:30:00Z",
    "total_requests": 0,
    "total_bytes_uploaded": 0,
    "total_bytes_downloaded": 0,
    "last_used_at": null,
    "sync_token": "sync_token_abc123def456ghi789..."
  }
}
```

**Errors:**
- `400 Bad Request`: Invalid parameters or format
- `404 Not Found`: Tenant with given ID does not exist
- `401 Unauthorized`: Invalid or expired token
- `403 Forbidden`: User is not an admin

---

### List All Sync Tokens

Retrieves all sync tokens across all tenants.

**Endpoint:** `GET /admin/sync-tokens`

**Authentication:** Required (Admin Bearer token)

**Query Parameters:** None

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Sync tokens retrieved",
  "data": [
    {
      "id": "880e8400-e29b-41d4-a716-446655440000",
      "tenant_id": "660e8400-e29b-41d4-a716-446655440000",
      "name": "Mobile App Token",
      "can_read": true,
      "can_write": true,
      "can_delete": false,
      "is_active": true,
      "created_by": "550e8400-e29b-41d4-a716-446655440000",
      "created_at": "2024-01-07T10:30:00Z",
      "updated_at": "2024-01-07T10:30:00Z",
      "expires_at": "2024-04-07T10:30:00Z",
      "total_requests": 150,
      "total_bytes_uploaded": 5242880,
      "total_bytes_downloaded": 10485760,
      "last_used_at": "2024-01-07T11:20:00Z",
      "tenant_email": "tenant@example.com",
      "tenant_name": "John Doe",
      "company_name": "Acme Corp"
    }
  ]
}
```

**Errors:**
- `401 Unauthorized`: Invalid or expired token
- `403 Forbidden`: User is not an admin

---

### Get Sync Token Details

Retrieves details of a specific sync token.

**Endpoint:** `GET /admin/sync-tokens/:id`

**Authentication:** Required (Admin Bearer token)

**Path Parameters:**
- `id` (string, required, UUID format): The sync token's ID

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Sync token retrieved",
  "data": {
    "id": "880e8400-e29b-41d4-a716-446655440000",
    "tenant_id": "660e8400-e29b-41d4-a716-446655440000",
    "name": "Mobile App Token",
    "can_read": true,
    "can_write": true,
    "can_delete": false,
    "is_active": true,
    "created_by": "550e8400-e29b-41d4-a716-446655440000",
    "created_at": "2024-01-07T10:30:00Z",
    "updated_at": "2024-01-07T10:30:00Z",
    "expires_at": "2024-04-07T10:30:00Z",
    "total_requests": 150,
    "total_bytes_uploaded": 5242880,
    "total_bytes_downloaded": 10485760,
    "last_used_at": "2024-01-07T11:20:00Z",
    "revoked_at": null,
    "revoked_by": null,
    "revoked_reason": null
  }
}
```

**Errors:**
- `400 Bad Request`: Invalid UUID format for ID
- `404 Not Found`: Sync token with given ID does not exist
- `401 Unauthorized`: Invalid or expired token
- `403 Forbidden`: User is not an admin

---

### List Tenant Sync Tokens

Retrieves all sync tokens for a specific tenant.

**Endpoint:** `GET /admin/tenants/:id/sync-tokens`

**Authentication:** Required (Admin Bearer token)

**Path Parameters:**
- `id` (string, required, UUID format): The tenant's ID

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Sync tokens retrieved",
  "data": [
    {
      "id": "880e8400-e29b-41d4-a716-446655440000",
      "tenant_id": "660e8400-e29b-41d4-a716-446655440000",
      "name": "Mobile App Token",
      "can_read": true,
      "can_write": true,
      "can_delete": false,
      "is_active": true,
      "created_by": "550e8400-e29b-41d4-a716-446655440000",
      "created_at": "2024-01-07T10:30:00Z",
      "updated_at": "2024-01-07T10:30:00Z",
      "expires_at": "2024-04-07T10:30:00Z",
      "total_requests": 150,
      "total_bytes_uploaded": 5242880,
      "total_bytes_downloaded": 10485760,
      "last_used_at": "2024-01-07T11:20:00Z"
    },
    {
      "id": "990e8400-e29b-41d4-a716-446655440000",
      "tenant_id": "660e8400-e29b-41d4-a716-446655440000",
      "name": "Desktop App Token",
      "can_read": true,
      "can_write": true,
      "can_delete": true,
      "is_active": true,
      "created_by": "550e8400-e29b-41d4-a716-446655440000",
      "created_at": "2024-01-06T15:45:00Z",
      "updated_at": "2024-01-06T15:45:00Z",
      "expires_at": "2025-01-06T15:45:00Z",
      "total_requests": 500,
      "total_bytes_uploaded": 20971520,
      "total_bytes_downloaded": 52428800,
      "last_used_at": "2024-01-07T10:15:00Z"
    }
  ]
}
```

**Errors:**
- `400 Bad Request`: Invalid UUID format for ID
- `404 Not Found`: Tenant with given ID does not exist
- `401 Unauthorized`: Invalid or expired token
- `403 Forbidden`: User is not an admin

---

### Rotate Sync Token

Creates a new sync token and optionally provides a grace period for the old token.

**Endpoint:** `POST /admin/sync-tokens/:id/rotate`

**Authentication:** Required (Admin Bearer token)

**Path Parameters:**
- `id` (string, required, UUID format): The sync token's ID to rotate

**Request Body:**
```json
{
  "token_id": "880e8400-e29b-41d4-a716-446655440000",
  "expires_in_days": 90,
  "grace_period_days": 7
}
```

**Parameters:**
- `token_id` (string, required, UUID format): The sync token's ID
- `expires_in_days` (integer, required): New token expiration period in days (1-365)
- `grace_period_days` (integer, optional, default: 7): Days to keep old token active (0-30)

**Response (201 Created):**
```json
{
  "success": true,
  "message": "Sync token rotated successfully",
  "data": {
    "old_token": {
      "id": "880e8400-e29b-41d4-a716-446655440000",
      "is_active": true,
      "expires_at": "2024-01-14T10:30:00Z"
    },
    "new_token": {
      "id": "aa0e8400-e29b-41d4-a716-446655440000",
      "tenant_id": "660e8400-e29b-41d4-a716-446655440000",
      "name": "Mobile App Token",
      "can_read": true,
      "can_write": true,
      "can_delete": false,
      "is_active": true,
      "created_by": "550e8400-e29b-41d4-a716-446655440000",
      "created_at": "2024-01-07T11:00:00Z",
      "expires_at": "2024-04-07T11:00:00Z",
      "sync_token": "sync_token_new123..."
    }
  }
}
```

**Errors:**
- `400 Bad Request`: Invalid parameters or format
- `404 Not Found`: Sync token with given ID does not exist
- `401 Unauthorized`: Invalid or expired token
- `403 Forbidden`: User is not an admin

---

### Revoke Sync Token

Immediately disables a sync token.

**Endpoint:** `DELETE /admin/sync-tokens/:id`

**Authentication:** Required (Admin Bearer token)

**Path Parameters:**
- `id` (string, required, UUID format): The sync token's ID

**Request Body:**
```json
{
  "reason": "Token compromised"
}
```

**Parameters:**
- `reason` (string, required): Reason for revoking the token

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Sync token revoked",
  "data": null
}
```

**Errors:**
- `400 Bad Request`: Invalid UUID format or missing reason
- `404 Not Found`: Sync token with given ID does not exist
- `401 Unauthorized`: Invalid or expired token
- `403 Forbidden`: User is not an admin

---

### Get Sync Token Statistics

Retrieves usage statistics for a specific sync token.

**Endpoint:** `GET /admin/sync-tokens/:id/stats`

**Authentication:** Required (Admin Bearer token)

**Path Parameters:**
- `id` (string, required, UUID format): The sync token's ID

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Sync token stats retrieved",
  "data": {
    "token_id": "880e8400-e29b-41d4-a716-446655440000",
    "token_name": "Mobile App Token",
    "total_requests": 150,
    "total_bytes_uploaded": 5242880,
    "total_bytes_downloaded": 10485760,
    "last_used_at": "2024-01-07T11:20:00Z",
    "created_at": "2024-01-07T10:30:00Z",
    "expires_at": "2024-04-07T10:30:00Z",
    "days_until_expiry": 90,
    "is_active": true
  }
}
```

**Errors:**
- `400 Bad Request`: Invalid UUID format for ID
- `404 Not Found`: Sync token with given ID does not exist
- `401 Unauthorized`: Invalid or expired token
- `403 Forbidden`: User is not an admin

---

## Response Format

### Success Response

All successful API responses follow this format:

```json
{
  "success": true,
  "message": "Description of the operation",
  "data": {
    // Response data varies by endpoint
  }
}
```

### Error Response

All error responses follow this format:

```json
{
  "success": false,
  "error": "Error message describing what went wrong"
}
```

---

## Error Handling

The API uses standard HTTP status codes:

| Code | Meaning | Example |
|------|---------|---------|
| 200 | OK | Successful GET/PATCH request |
| 201 | Created | Successful POST request |
| 400 | Bad Request | Invalid parameters or format |
| 401 | Unauthorized | Invalid or missing authentication token |
| 403 | Forbidden | Authenticated but insufficient permissions |
| 404 | Not Found | Resource doesn't exist |
| 409 | Conflict | Email already exists |
| 500 | Server Error | Unexpected server error |

### Common Error Messages

- **Invalid email format**: Email doesn't match standard email pattern
- **Email already exists**: The provided email is already registered in the system
- **Invalid UUID format**: The ID parameter is not a valid UUID
- **User not found**: The specified user/tenant/token doesn't exist
- **Invalid token**: JWT token is malformed or expired
- **Insufficient permissions**: User lacks required admin role
- **Account disabled**: The admin or tenant account is inactive

---

## Rate Limiting

Admin endpoints are subject to rate limiting:

- **Per-user limit**: 20 requests per second
- **Burst limit**: Up to 20 concurrent requests

Rate limit information is included in response headers:
- `X-RateLimit-Limit`: Maximum requests allowed
- `X-RateLimit-Remaining`: Requests remaining in current window
- `X-RateLimit-Reset`: Unix timestamp when limit resets

If you exceed the rate limit, you'll receive a 429 status code.

---

## Example Workflows

### Creating a New Tenant and Setting Up Sync Access

1. **Create the tenant:**
   ```bash
   curl -X POST http://localhost:8080/api/v1/admin/tenants \
     -H "Authorization: Bearer <admin_token>" \
     -H "Content-Type: application/json" \
     -d '{
       "email": "newclient@example.com",
       "full_name": "New Client",
       "company_name": "Client Corp"
     }'
   ```
   Save the `temporary_password` from the response.

2. **Create sync tokens for the tenant:**
   ```bash
   curl -X POST http://localhost:8080/api/v1/admin/sync-tokens \
     -H "Authorization: Bearer <admin_token>" \
     -H "Content-Type: application/json" \
     -d '{
       "tenant_id": "<tenant_id>",
       "name": "Production Sync",
       "can_read": true,
       "can_write": true,
       "can_delete": false,
       "expires_in_days": 365
     }'
   ```
   Save the `sync_token` from the response.

### Rotating a Sync Token

```bash
curl -X POST http://localhost:8080/api/v1/admin/sync-tokens/<token_id>/rotate \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "token_id": "<token_id>",
    "expires_in_days": 90,
    "grace_period_days": 7
  }'
```

### Disabling a Tenant Account

```bash
curl -X PATCH http://localhost:8080/api/v1/admin/tenants/<tenant_id>/status \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "is_active": false,
    "disabled_reason": "Account suspended due to policy violation"
  }'
```

---

## Notes for Frontend Development

1. **Token Storage**: Store access tokens in memory or short-lived httpOnly cookies. Store refresh tokens securely (httpOnly cookie or secure storage).

2. **Token Refresh**: Implement automatic token refresh logic using the refresh token before the access token expires.

3. **Error Handling**: Always check the `success` field and `error` message in responses.

4. **UUID Format**: All IDs in the system are UUIDs (36-character strings with hyphens). Validate format before sending to API.

5. **Timestamps**: All timestamps are in ISO 8601 format (UTC). Convert to local time on the frontend as needed.

6. **Permissions**: Sync tokens have granular permissions (read, write, delete). Validate token permissions before allowing file operations.

7. **Rate Limiting**: Implement exponential backoff when receiving rate limit responses (429 status code).

8. **CORS**: The API includes CORS headers. Ensure your frontend domain is configured in the backend.

