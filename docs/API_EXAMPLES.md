# 📡 API Usage Examples

## Upload to Different Folders

All examples assume you have a valid JWT token in the `Authorization` header.

### Example 1: Upload to `uploads` folder
```bash
curl -X POST http://localhost:8080/api/v1/files/upload-url \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "filename": "Resume.pdf",
    "file_size": 1024000,
    "mime_type": "application/pdf",
    "upload_to": "uploads"
  }'
```

**Response:**
```json
{
  "success": true,
  "message": "Upload URL generated",
  "data": {
    "file_id": "550e8400-e29b-41d4-a716-446655440000",
    "upload_url": "https://s3.amazonaws.com/bucket/...",
    "s3_key": "tenants/tenant-id-123/uploads/550e8400e29b.pdf",
    "expires_in": 900
  }
}
```

### Example 2: Upload to `assets` folder
```bash
curl -X POST http://localhost:8080/api/v1/files/upload-url \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "filename": "logo.png",
    "file_size": 256000,
    "mime_type": "image/png",
    "upload_to": "assets"
  }'
```

### Example 3: Upload to `schedules` folder
```bash
curl -X POST http://localhost:8080/api/v1/files/upload-url \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "filename": "sync_schedule.json",
    "file_size": 4096,
    "mime_type": "application/json",
    "upload_to": "schedules"
  }'
```

---

## Error Scenarios

### Invalid `upload_to`
```bash
curl -X POST http://localhost:8080/api/v1/files/upload-url \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "filename": "test.pdf",
    "file_size": 1024000,
    "mime_type": "application/pdf",
    "upload_to": "invalid_folder"
  }'
```

**Response (400):**
```json
{
  "success": false,
  "error": "invalid upload destination: invalid upload destination: invalid_folder"
}
```

### Missing `upload_to`
```bash
curl -X POST http://localhost:8080/api/v1/files/upload-url \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "filename": "test.pdf",
    "file_size": 1024000,
    "mime_type": "application/pdf"
  }'
```

**Response (400):**
```json
{
  "success": false,
  "error": "Key: 'UploadURLRequest.UploadTo' Error:Field validation for 'UploadTo' failed on the 'required' tag"
}
```

---

## Complete Upload Flow

### Step 1: Get Upload URL
```bash
FILE_ID=$(curl -X POST http://localhost:8080/api/v1/files/upload-url \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "filename": "data.csv",
    "file_size": 102400,
    "mime_type": "text/csv",
    "upload_to": "uploads"
  }' | jq -r '.data.file_id')

UPLOAD_URL=$(curl -X POST http://localhost:8080/api/v1/files/upload-url \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "filename": "data.csv",
    "file_size": 102400,
    "mime_type": "text/csv",
    "upload_to": "uploads"
  }' | jq -r '.data.upload_url')
```

### Step 2: Upload to Presigned URL
```bash
curl -X PUT "$UPLOAD_URL" \
  -H "Content-Type: text/csv" \
  --data-binary @data.csv
```

### Step 3: Mark Upload Complete
```bash
curl -X POST http://localhost:8080/api/v1/files/complete-upload \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"file_id\": \"$FILE_ID\"
  }"
```

### Step 4: List Files
```bash
curl -X GET http://localhost:8080/api/v1/files \
  -H "Authorization: Bearer TOKEN"
```

### Step 5: Download File
```bash
DOWNLOAD_URL=$(curl -X GET http://localhost:8080/api/v1/files/$FILE_ID/download-url \
  -H "Authorization: Bearer TOKEN" | jq -r '.data.download_url')

curl "$DOWNLOAD_URL" -o data.csv
```

---

## Adding a New Folder (Without Code Changes)

### Step 1: Edit Template
```json
{
  "type": "folder",
  "path": "reports",
  "is_required": true,
  "permissions": {
    "allow_files": true,
    "allow_folders": false,
    "can_upload": true,
    "can_delete": true,
    "can_replace": true,
    "can_list": true
  },
  "metadata": {
    "description": "Generated reports"
  }
}
```

### Step 2: Restart Backend
```bash
make docker-restart
```

### Step 3: Use New Folder
```bash
curl -X POST http://localhost:8080/api/v1/files/upload-url \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "filename": "monthly_report.pdf",
    "file_size": 2048000,
    "mime_type": "application/pdf",
    "upload_to": "reports"
  }'
```

---

## Validation Rules

| Field | Rule | Example |
|-------|------|---------|
| `filename` | Required, non-empty | `"Resume.pdf"` |
| `file_size` | Required, > 0 | `1024000` |
| `mime_type` | Required | `"application/pdf"` |
| `upload_to` | Must match template node | `"uploads"`, `"assets"`, `"schedules"` |

---

## S3 Path Structure

All files stored with structure:
```
s3://bucket-name/
  └── tenants/
      └── {tenant_id}/          # Tenant isolation
          ├── uploads/
          │   ├── file1.pdf
          │   └── file2.doc
          ├── assets/
          │   ├── logo.png
          │   └── icon.svg
          └── schedules/
              ├── sync1.json
              └── sync2.json
```

**Tenant cannot:**
- Access other tenant's folders
- Create arbitrary paths
- Use path traversal (`../`)
- Bypass folder restrictions
