# 🎯 Template System: Quick Start Guide

## What You Now Have

A **production-ready, JSON-controlled file upload system** where:
- ✅ All upload destinations defined in one file
- ✅ No database changes needed to add folders
- ✅ Tenant isolation guaranteed
- ✅ Security built-in
- ✅ Fully tested

---

## The Template File

📄 **Location:** `/home/saurabh/backend/default_dir_tree_template.json`

```json
{
  "version": "1.0.0",
  "root_path": "tenants/{tenant_id}/",
  "nodes": [
    {
      "path": "uploads",
      "type": "folder",
      "permissions": { "can_upload": true, "allow_files": true }
    },
    {
      "path": "assets",
      "type": "folder",
      "permissions": { "can_upload": true, "allow_files": true }
    },
    {
      "path": "schedules",
      "type": "folder",
      "permissions": { "can_upload": true, "allow_files": true }
    }
  ]
}
```

---

## How It Works

### Client Requests Upload
```json
POST /api/v1/files/upload-url

{
  "filename": "Resume.pdf",
  "file_size": 1024000,
  "mime_type": "application/pdf",
  "upload_to": "uploads"  ← Must match template node
}
```

### Backend Validates
```
1. Check if upload_to exists in template
2. Check if that node allows uploads
3. Resolve S3 path: tenants/{tenant_id}/uploads/uuid.pdf
4. Return presigned URL
```

### No Cross-Tenant Access
```
Tenant A cannot access:
- Tenant B's upload folder
- Other tenant's files
- Arbitrary S3 paths

Everything isolated by tenant_id
```

---

## Adding a New Upload Folder

### 1️⃣ Edit Template File
```bash
vim default_dir_tree_template.json
```

### 2️⃣ Add New Node
```json
{
  "path": "reports",
  "type": "folder",
  "permissions": {
    "can_upload": true,
    "allow_files": true,
    "can_delete": true,
    "can_replace": true,
    "can_list": true
  }
}
```

### 3️⃣ Restart Backend
```bash
make docker-restart
# or docker-compose restart backend
```

### Done! ✨
Now tenants can upload to `reports` folder without any code changes.

---

## Files Modified/Created

| File | Purpose |
|------|---------|
| [default_dir_tree_template.json](default_dir_tree_template.json) | Upload destinations (your control) |
| [internal/models/template.go](internal/models/template.go) | Template structs + validation |
| [internal/config/template_loader.go](internal/config/template_loader.go) | Load template at startup |
| [internal/models/file.go](internal/models/file.go) | Added `upload_to` field |
| [internal/service/file_service.go](internal/service/file_service.go) | Validate against template |
| [cmd/server/main.go](cmd/server/main.go) | Load template on startup |

---

## Security Guarantees ✅

- ✅ **No path traversal** (`../` blocked)
- ✅ **No arbitrary paths** (only template nodes allowed)
- ✅ **Tenant isolation** (paths always start with `tenants/{tenant_id}/`)
- ✅ **Permission enforcement** (can_upload must be true)
- ✅ **Zero DB queries** (validation in-memory)

---

## Key Classes

### `models.FileTreeTemplate`
- Represents the template in memory
- Has node lookup index
- Methods:
  - `ValidateUploadDestination(uploadTo)` → validates path
  - `ResolveS3Path(tenantID, uploadTo, filename)` → builds S3 key
  - `GetNode(path)` → O(1) lookup

### `config.LoadFileTreeTemplate(projectRoot)`
- Loads JSON file at startup
- Validates structure
- Builds internal index
- Fails application if invalid

### Updated `FileService`
- Takes template in constructor
- Validates every upload request
- Uses template to resolve S3 paths
- No more hardcoded S3 prefixes

---

## Tests ✅

All tests passing:
```bash
go test ./internal/models -run TestFileTreeTemplate
go test ./internal/config -run TestLoad
```

Coverage:
- Template validation logic
- S3 path resolution
- Node lookup
- Template loading from file
- Error scenarios

---

## Example Upload Request

```bash
curl -X POST http://localhost:8080/api/v1/files/upload-url \
  -H "Authorization: Bearer TOKEN" \
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
  "data": {
    "file_id": "550e8400-e29b-41d4-a716-446655440000",
    "upload_url": "https://s3.amazonaws.com/...",
    "s3_key": "tenants/tenant-123/uploads/550e8400.pdf",
    "expires_in": 900
  }
}
```

---

## Invalid Request

```bash
# ERROR: upload_to doesn't exist in template
curl -X POST http://localhost:8080/api/v1/files/upload-url \
  -H "Authorization: Bearer TOKEN" \
  -d '{
    ...
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

---

## Design Principles

1. **Template is authoritative** — All upload paths defined here
2. **No DB migrations** — Template changes don't require schema updates
3. **Tenant-agnostic** — Same template for all tenants
4. **Fast validation** — O(1) lookups, no database queries
5. **Secure by default** — All access controlled by template

---

## What's Next?

- [x] ✅ Implement template system
- [x] ✅ Validate upload requests
- [x] ✅ Resolve S3 paths securely
- [x] ✅ Add unit tests
- [x] ✅ Build passes

**Ready for deployment!** 🚀

---

## Support

For full implementation details, see: [IMPLEMENTATION.md](IMPLEMENTATION.md)

For API examples, see: [API_EXAMPLES.md](API_EXAMPLES.md)
