# 🧪 Complete API Testing Guide - localhost:8081

## 📋 Table of Contents
1. [Health Checks](#health-checks)
2. [Admin Authentication](#admin-authentication)
3. [Tenant Management (Admin)](#tenant-management-admin)
4. [Tenant Authentication](#tenant-authentication)
5. [Tenant Operations](#tenant-operations)
6. [File Operations](#file-operations)
7. [Token Management](#token-management)

---

## 🔧 Health Checks

### 1. Health Check
```bash
curl -X GET http://localhost:8081/health
```

### 2. Readiness Check
```bash
curl -X GET http://localhost:8081/ready
```

---

## 🔐 Admin Authentication

### 3. Admin Login
```bash
curl -X POST http://localhost:8081/api/v1/auth/admin/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "your email",
    "password": "YOUR_ADMIN_PASSWORD"
  }'
```

**Response:** Save `access_token` and `refresh_token`

**Expected:**
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "access_token": "eyJhbGci...",
    "refresh_token": "eyJhbGci...",
    "token_type": "Bearer",
    "expires_in": 900,
    "user": {
      "id": "xxx-xxx-xxx",
      "email": "saurabh25decem2020@gmail.com",
      "full_name": "Your Name",
      "user_type": "admin"
    }
  }
}
```

---

## 👥 Tenant Management (Admin)

### 4. Create Tenant (Admin Only)
```bash
curl -X POST http://localhost:8081/api/v1/admin/tenants \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ADMIN_ACCESS_TOKEN" \
  -d '{
    "email": "tenant1@example.com",
    "full_name": "Test Tenant",
    "company_name": "Test Company"
  }'
```

**Response:** Save `temporary_password`

---

### 5. List All Tenants (Admin Only)
```bash
curl -X GET http://localhost:8081/api/v1/admin/tenants \
  -H "Authorization: Bearer YOUR_ADMIN_ACCESS_TOKEN"
```

---

### 6. Get Tenant Details (Admin Only)
```bash
curl -X GET http://localhost:8081/api/v1/admin/tenants/09aba49a-5718-4bae-a187-f94af0edfa92 \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiZjMxODUwOTUtYTg4Zi00ZjBhLWJiYzQtMWY3OWM2M2I0MmU1IiwiZW1haWwiOiJzYXVyYWJoMjVkZWNlbTIwMjBAZ21haWwuY29tIiwidXNlcl90eXBlIjoiYWRtaW4iLCJpc3MiOiJmaWxlLXN0b3JhZ2UtYXBpIiwic3ViIjoiZjMxODUwOTUtYTg4Zi00ZjBhLWJiYzQtMWY3OWM2M2I0MmU1IiwiZXhwIjoxNzY2NzI5NzcwLCJpYXQiOjE3NjY3Mjg4NzB9.96VYI3R3BlkEn6gtmFXncjQugNAkyKtOub7kvz5RRbU"
```

---

### 7. Reset Tenant Password (Admin Only)
```bash
curl -X POST http://localhost:8081/api/v1/admin/tenants/TENANT_ID/reset-password \
  -H "Authorization: Bearer YOUR_ADMIN_ACCESS_TOKEN"
```

**Response:** Returns new `temporary_password`

---

### 8. Enable/Disable Tenant (Admin Only)
```bash
# Disable tenant
curl -X PATCH http://localhost:8081/api/v1/admin/tenants/TENANT_ID/status \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ADMIN_ACCESS_TOKEN" \
  -d '{
    "is_active": false,
    "disabled_reason": "Subscription expired"
  }'

# Enable tenant
curl -X PATCH http://localhost:8081/api/v1/admin/tenants/TENANT_ID/status \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ADMIN_ACCESS_TOKEN" \
  -d '{
    "is_active": true
  }'
```

---

## 🔑 Tenant Authentication

### 9. Tenant Login
```bash
curl -X POST http://localhost:8081/api/v1/auth/tenant/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "tenant1@example.com",
    "password": "TEMPORARY_PASSWORD_FROM_CREATION"
  }'
```

**Response:** Save `access_token` and `refresh_token`

---

## 👤 Tenant Operations

### 10. Get Tenant Profile
```bash
curl -X GET http://localhost:8081/api/v1/tenant/profile \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMDlhYmE0OWEtNTcxOC00YmFlLWExODctZjk0YWYwZWRmYTkyIiwiZW1haWwiOiJ0ZW5hbnQxQGV4YW1wbGUuY29tIiwidXNlcl90eXBlIjoidGVuYW50IiwiaXNzIjoiZmlsZS1zdG9yYWdlLWFwaSIsInN1YiI6IjA5YWJhNDlhLTU3MTgtNGJhZS1hMTg3LWY5NGFmMGVkZmE5MiIsImV4cCI6MTc2NzAwMTcwMCwiaWF0IjoxNzY3MDAwODAwfQ._1GoQPtwcBTlfaeH1Qa2AjDBWrOhdqyQCPSisszc_No"
```

---

### 11. Change Tenant Password
```bash
curl -X POST http://localhost:8081/api/v1/tenant/change-password \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiYTUxZmZlOTQtZDcxYS00ZGU1LTljOWEtOGZkOGJmZGQzMzIzIiwiZW1haWwiOiJ2aW5vZGt1bGthcm5pQGdyb3Zlc3lzdGVtcy5jbyIsInVzZXJfdHlwZSI6InRlbmFudCIsImlzcyI6ImZpbGUtc3RvcmFnZS1hcGkiLCJzdWIiOiJhNTFmZmU5NC1kNzFhLTRkZTUtOWM5YS04ZmQ4YmZkZDMzMjMiLCJleHAiOjE3NjcwODEyOTUsImlhdCI6MTc2NzA4MDM5NX0.j-BSs4TGl2C67vciN0PT97VexIc4z4Otq6AGKZcQqfI" \
  -d '{
    "current_password": "J4NMfRiSGq9__qv7",
    "new_password": "Vinod1234"
  }'
```

---

## 📁 File Operations

### 12. Generate Upload URL
```bash
curl -X POST http://localhost:8081/api/v1/files/upload-url \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TENANT_ACCESS_TOKEN" \
  -d '{
    "filename": "test-document.pdf",
    "file_size": 1024000,
    "mime_type": "application/pdf"
  }'
```

**Response:** Save `file_id` and `upload_url`

---

### 13. Upload File to S3 (Use Presigned URL)
```bash
curl -X PUT "PRESIGNED_UPLOAD_URL_FROM_STEP_12" \
  -H "Content-Type: application/pdf" \
  --data-binary "@/path/to/your/file.pdf"
```

**Note:** Use the exact `upload_url` from step 12

---

### 14. Complete Upload
```bash
curl -X POST http://localhost:8081/api/v1/files/complete-upload \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TENANT_ACCESS_TOKEN" \
  -d '{
    "file_id": "FILE_ID_FROM_STEP_12"
  }'
```

---

### 15. List All Files
```bash
curl -X GET http://localhost:8081/api/v1/files \
  -H "Authorization: Bearer YOUR_TENANT_ACCESS_TOKEN"
```

---

### 16. Generate Download URL
```bash
curl -X GET http://localhost:8081/api/v1/files/FILE_ID/download-url \
  -H "Authorization: Bearer YOUR_TENANT_ACCESS_TOKEN"
```

**Response:** Save `download_url`

---

### 17. Download File (Use Presigned URL)
```bash
curl -o downloaded-file.pdf "PRESIGNED_DOWNLOAD_URL_FROM_STEP_16"
```

---

### 18. Delete File
```bash
curl -X DELETE http://localhost:8081/api/v1/files/FILE_ID \
  -H "Authorization: Bearer YOUR_TENANT_ACCESS_TOKEN"
```

---

## 🔄 Token Management

### 19. Refresh Access Token
```bash
curl -X POST http://localhost:8081/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "YOUR_REFRESH_TOKEN"
  }'
```

**Response:** Returns new `access_token`

---

### 20. Logout
```bash
curl -X POST http://localhost:8081/api/v1/auth/logout \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "YOUR_REFRESH_TOKEN"
  }'
```

---

## 📦 Postman Collection JSON

Copy this JSON to import into Postman:

```json
{
  "info": {
    "name": "File Storage API",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "item": [
    {
      "name": "Health Checks",
      "item": [
        {
          "name": "Health Check",
          "request": {
            "method": "GET",
            "header": [],
            "url": {
              "raw": "http://localhost:8081/health",
              "protocol": "http",
              "host": ["localhost"],
              "port": "8081",
              "path": ["health"]
            }
          }
        },
        {
          "name": "Ready Check",
          "request": {
            "method": "GET",
            "header": [],
            "url": {
              "raw": "http://localhost:8081/ready",
              "protocol": "http",
              "host": ["localhost"],
              "port": "8081",
              "path": ["ready"]
            }
          }
        }
      ]
    },
    {
      "name": "Admin Auth",
      "item": [
        {
          "name": "Admin Login",
          "event": [
            {
              "listen": "test",
              "script": {
                "exec": [
                  "if (pm.response.code === 200) {",
                  "    var jsonData = pm.response.json();",
                  "    pm.environment.set(\"admin_access_token\", jsonData.data.access_token);",
                  "    pm.environment.set(\"admin_refresh_token\", jsonData.data.refresh_token);",
                  "}"
                ]
              }
            }
          ],
          "request": {
            "method": "POST",
            "header": [{"key": "Content-Type", "value": "application/json"}],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"email\": \"saurabh25decem2020@gmail.com\",\n  \"password\": \"YOUR_PASSWORD\"\n}"
            },
            "url": {
              "raw": "http://localhost:8081/api/v1/auth/admin/login",
              "protocol": "http",
              "host": ["localhost"],
              "port": "8081",
              "path": ["api", "v1", "auth", "admin", "login"]
            }
          }
        }
      ]
    },
    {
      "name": "Tenant Management",
      "item": [
        {
          "name": "Create Tenant",
          "event": [
            {
              "listen": "test",
              "script": {
                "exec": [
                  "if (pm.response.code === 201) {",
                  "    var jsonData = pm.response.json();",
                  "    pm.environment.set(\"tenant_id\", jsonData.data.tenant.id);",
                  "    pm.environment.set(\"tenant_temp_password\", jsonData.data.temporary_password);",
                  "}"
                ]
              }
            }
          ],
          "request": {
            "method": "POST",
            "header": [
              {"key": "Content-Type", "value": "application/json"},
              {"key": "Authorization", "value": "Bearer {{admin_access_token}}"}
            ],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"email\": \"tenant1@example.com\",\n  \"full_name\": \"Test Tenant\",\n  \"company_name\": \"Test Company\"\n}"
            },
            "url": {
              "raw": "http://localhost:8081/api/v1/admin/tenants",
              "protocol": "http",
              "host": ["localhost"],
              "port": "8081",
              "path": ["api", "v1", "admin", "tenants"]
            }
          }
        },
        {
          "name": "List Tenants",
          "request": {
            "method": "GET",
            "header": [
              {"key": "Authorization", "value": "Bearer {{admin_access_token}}"}
            ],
            "url": {
              "raw": "http://localhost:8081/api/v1/admin/tenants",
              "protocol": "http",
              "host": ["localhost"],
              "port": "8081",
              "path": ["api", "v1", "admin", "tenants"]
            }
          }
        }
      ]
    },
    {
      "name": "Tenant Auth",
      "item": [
        {
          "name": "Tenant Login",
          "event": [
            {
              "listen": "test",
              "script": {
                "exec": [
                  "if (pm.response.code === 200) {",
                  "    var jsonData = pm.response.json();",
                  "    pm.environment.set(\"tenant_access_token\", jsonData.data.access_token);",
                  "    pm.environment.set(\"tenant_refresh_token\", jsonData.data.refresh_token);",
                  "}"
                ]
              }
            }
          ],
          "request": {
            "method": "POST",
            "header": [{"key": "Content-Type", "value": "application/json"}],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"email\": \"tenant1@example.com\",\n  \"password\": \"{{tenant_temp_password}}\"\n}"
            },
            "url": {
              "raw": "http://localhost:8081/api/v1/auth/tenant/login",
              "protocol": "http",
              "host": ["localhost"],
              "port": "8081",
              "path": ["api", "v1", "auth", "tenant", "login"]
            }
          }
        }
      ]
    },
    {
      "name": "File Operations",
      "item": [
        {
          "name": "Generate Upload URL",
          "event": [
            {
              "listen": "test",
              "script": {
                "exec": [
                  "if (pm.response.code === 200) {",
                  "    var jsonData = pm.response.json();",
                  "    pm.environment.set(\"file_id\", jsonData.data.file_id);",
                  "    pm.environment.set(\"upload_url\", jsonData.data.upload_url);",
                  "}"
                ]
              }
            }
          ],
          "request": {
            "method": "POST",
            "header": [
              {"key": "Content-Type", "value": "application/json"},
              {"key": "Authorization", "value": "Bearer {{tenant_access_token}}"}
            ],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"filename\": \"test.pdf\",\n  \"file_size\": 1024000,\n  \"mime_type\": \"application/pdf\"\n}"
            },
            "url": {
              "raw": "http://localhost:8081/api/v1/files/upload-url",
              "protocol": "http",
              "host": ["localhost"],
              "port": "8081",
              "path": ["api", "v1", "files", "upload-url"]
            }
          }
        },
        {
          "name": "Complete Upload",
          "request": {
            "method": "POST",
            "header": [
              {"key": "Content-Type", "value": "application/json"},
              {"key": "Authorization", "value": "Bearer {{tenant_access_token}}"}
            ],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"file_id\": \"{{file_id}}\"\n}"
            },
            "url": {
              "raw": "http://localhost:8081/api/v1/files/complete-upload",
              "protocol": "http",
              "host": ["localhost"],
              "port": "8081",
              "path": ["api", "v1", "files", "complete-upload"]
            }
          }
        },
        {
          "name": "List Files",
          "request": {
            "method": "GET",
            "header": [
              {"key": "Authorization", "value": "Bearer {{tenant_access_token}}"}
            ],
            "url": {
              "raw": "http://localhost:8081/api/v1/files",
              "protocol": "http",
              "host": ["localhost"],
              "port": "8081",
              "path": ["api", "v1", "files"]
            }
          }
        }
      ]
    }
  ]
}
```

---

## 🔧 Postman Environment Variables

Create these environment variables in Postman:

```json
{
  "name": "File Storage Dev",
  "values": [
    {"key": "base_url", "value": "http://localhost:8081", "enabled": true},
    {"key": "admin_access_token", "value": "", "enabled": true},
    {"key": "admin_refresh_token", "value": "", "enabled": true},
    {"key": "tenant_access_token", "value": "", "enabled": true},
    {"key": "tenant_refresh_token", "value": "", "enabled": true},
    {"key": "tenant_id", "value": "", "enabled": true},
    {"key": "tenant_temp_password", "value": "", "enabled": true},
    {"key": "file_id", "value": "", "enabled": true},
    {"key": "upload_url", "value": "", "enabled": true}
  ]
}
```

---

## 🧪 Testing Flow

1. **Health Check** → Verify API is running
2. **Admin Login** → Get admin token
3. **Create Tenant** → Get tenant credentials
4. **Tenant Login** → Get tenant token
5. **Generate Upload URL** → Get presigned URL
6. **Upload File** → Upload to S3 using presigned URL
7. **Complete Upload** → Mark upload as completed
8. **List Files** → View all files
9. **Download File** → Get download URL and download
10. **Delete File** → Remove file

---

## 📝 Notes

- Replace `YOUR_ADMIN_PASSWORD` with your actual admin password
- Replace `TENANT_ID`, `FILE_ID` with actual UUIDs from responses
- Tokens expire in 15 minutes (access) and 7 days (refresh)
- Use refresh token endpoint to get new access tokens
- All file operations require tenant authentication
- All admin operations require admin authentication

---

## 🎯 Quick Test Sequence

```bash
# 1. Health
curl http://localhost:8081/health

# 2. Admin Login (save token)
curl -X POST http://localhost:8081/api/v1/auth/admin/login \
  -H "Content-Type: application/json" \
  -d '{"email":"saurabh25decem2020@gmail.com","password":"YOUR_PASSWORD"}'

# 3. Create Tenant (use admin token from step 2)
curl -X POST http://localhost:8081/api/v1/admin/tenants \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -d '{"email":"tenant@test.com","full_name":"Test"}'

# 4. Tenant Login (use temp password from step 3)
curl -X POST http://localhost:8081/api/v1/auth/tenant/login \
  -H "Content-Type: application/json" \
  -d '{"email":"tenant@test.com","password":"TEMP_PASSWORD"}'

# 5. List files (use tenant token from step 4)
curl http://localhost:8081/api/v1/files \
  -H "Authorization: Bearer TENANT_TOKEN"
```

🚀 **Your API is ready for testing!