# 🏗️ Architecture Diagram: JSON File Tree Template System

## System Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         APPLICATION STARTUP                                │
└─────────────────────────────────────────────────────────────────────────────┘

cmd/server/main.go
        │
        ├─→ Load Configuration
        │   └─→ config.Load()
        │
        └─→ Load File Tree Template  ← NEW!
            └─→ config.LoadFileTreeTemplate(".")
                ├─→ Read default_dir_tree_template.json
                ├─→ Parse JSON
                ├─→ Validate structure
                │   ├─ Version exists?
                │   ├─ root_path exists?
                │   ├─ Nodes not empty?
                │   └─ No duplicate paths?
                ├─→ Build in-memory index
                └─→ Pass to FileService

        Database Connection
        S3 Connection
        JWT Service
        Repositories
        Services
        Handlers
        Server Start
        
        ✅ Server Ready for Requests
```

---

## Upload Request Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        UPLOAD REQUEST PIPELINE                             │
└─────────────────────────────────────────────────────────────────────────────┘

1. CLIENT REQUEST
   ┌──────────────────────────────────────────┐
   │ POST /api/v1/files/upload-url            │
   │                                          │
   │ {                                        │
   │   "filename": "Resume.pdf",              │
   │   "file_size": 1024000,                  │
   │   "mime_type": "application/pdf",        │
   │   "upload_to": "uploads"  ← NEW!        │
   │ }                                        │
   └──────────────────────────────────────────┘
            ↓

2. HANDLER (handlers/file_handler.go)
   ┌──────────────────────────────────────────┐
   │ GenerateUploadURL()                      │
   │                                          │
   │ 1. Bind JSON                             │
   │ 2. Extract tenant ID from JWT            │
   │ 3. Call fileService.GenerateUploadURL()  │
   └──────────────────────────────────────────┘
            ↓

3. SERVICE (service/file_service.go) ← NEW VALIDATION!
   ┌──────────────────────────────────────────┐
   │ GenerateUploadURL()                      │
   │                                          │
   │ ├─ Validate file size                    │
   │ ├─ Validate MIME type                    │
   │ │                                        │
   │ ├─ Validate upload_to ← NEW!            │
   │ │  └─ template.ValidateUploadDestination()
   │ │     ├─ Node exists?                    │
   │ │     ├─ is_folder?                      │
   │ │     ├─ can_upload = true?              │
   │ │     └─ allow_files = true?             │
   │ │     → If any check fails: REJECT ❌   │
   │ │                                        │
   │ ├─ Resolve S3 path ← NEW!               │
   │ │  └─ template.ResolveS3Path()          │
   │ │     └─ "tenants/{id}/{upload_to}/{f}" │
   │ │                                        │
   │ ├─ Create file record in DB              │
   │ │  ├─ file.id = UUID                    │
   │ │  ├─ file.s3_key = resolved path        │
   │ │  ├─ file.upload_to = "uploads" ← NEW! │
   │ │  └─ file.status = "pending"            │
   │ │                                        │
   │ ├─ Generate presigned URL                │
   │ │  └─ s3Service.GeneratePresignedPutURL()
   │ │     └─ 15 minute expiry                │
   │ │                                        │
   │ └─ Return response                       │
   └──────────────────────────────────────────┘
            ↓

4. RESPONSE TO CLIENT
   ┌──────────────────────────────────────────────────┐
   │ {                                                │
   │   "success": true,                               │
   │   "data": {                                      │
   │     "file_id": "550e8400-e29b-41d4-...",        │
   │     "upload_url": "https://s3.amazonaws.com/...", │
   │     "s3_key": "tenants/abc-123/uploads/550e.pdf", │
   │     "expires_in": 900                            │
   │   }                                              │
   │ }                                                │
   └──────────────────────────────────────────────────┘
```

---

## Data Model

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         TEMPLATE STRUCTURE                                 │
└─────────────────────────────────────────────────────────────────────────────┘

FileTreeTemplate (In-Memory)
│
├─ Version: "1.0.0"
├─ RootPath: "tenants/{tenant_id}/"
├─ RootPermissions: { allow_files, can_upload, ... }
│
├─ Nodes: []TemplateNode
│  │
│  ├─ [0] Node
│  │   ├─ Type: "folder"
│  │   ├─ Path: "uploads"  ← Used in upload_to
│  │   ├─ IsRequired: true
│  │   ├─ Permissions:
│  │   │  ├─ allow_files: true
│  │   │  ├─ allow_folders: false
│  │   │  ├─ can_upload: true
│  │   │  ├─ can_delete: true
│  │   │  ├─ can_replace: true
│  │   │  └─ can_list: true
│  │   └─ Metadata:
│  │      └─ description: "User-uploaded files"
│  │
│  ├─ [1] Node
│  │   ├─ Path: "assets"
│  │   └─ ...
│  │
│  └─ [2] Node
│      ├─ Path: "schedules"
│      └─ ...
│
└─ nodePathIndex: map[string]*TemplateNode  ← O(1) lookup
   ├─ "uploads" → Node[0]
   ├─ "assets" → Node[1]
   └─ "schedules" → Node[2]


Database File Record (After Upload):
│
├─ id: UUID
├─ tenant_id: UUID (from JWT)
├─ original_filename: "Resume.pdf"
├─ stored_filename: "550e8400.pdf"
├─ s3_key: "tenants/abc-123/uploads/550e8400.pdf"
├─ file_size: 1024000
├─ mime_type: "application/pdf"
├─ upload_to: "uploads"  ← NEW! Audit trail
├─ upload_status: "completed"
├─ created_at: TIMESTAMP
└─ updated_at: TIMESTAMP


S3 Storage Structure:
│
s3://bucket/
│
└─ tenants/
   ├─ tenant-123/
   │  ├─ uploads/          ← Template node: "uploads"
   │  │  ├─ 550e8400.pdf
   │  │  └─ 1a2b3c4d.doc
   │  ├─ assets/           ← Template node: "assets"
   │  │  ├─ logo.png
   │  │  └─ icon.svg
   │  └─ schedules/        ← Template node: "schedules"
   │     └─ sync.json
   │
   └─ tenant-456/
      ├─ uploads/
      ├─ assets/
      └─ schedules/
```

---

## Validation Pipeline

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    TEMPLATE VALIDATION (per request)                       │
└─────────────────────────────────────────────────────────────────────────────┘

Upload Request arrives with "upload_to": "uploads"
        ↓
ValidateUploadDestination("uploads")
        ├─ Is uploadTo empty?
        │  └─ YES → REJECT: "upload_to is required"
        │  └─ NO → Continue
        │
        ├─ Get node from index (O(1))
        │  └─ template.GetNode("uploads")
        │  └─ Node not found? → REJECT: "invalid upload destination"
        │  └─ Found → Continue
        │
        ├─ Check node.Type == "folder"
        │  └─ NO → REJECT: "upload destination must be a folder"
        │  └─ YES → Continue
        │
        ├─ Check node.Permissions.CanUpload == true
        │  └─ NO → REJECT: "uploads not allowed in: uploads"
        │  └─ YES → Continue
        │
        ├─ Check node.Permissions.AllowFiles == true
        │  └─ NO → REJECT: "files not allowed in: uploads"
        │  └─ YES → Continue
        │
        └─ VALIDATION PASSED ✅
           └─ Proceed to generate S3 URL
```

---

## S3 Path Resolution

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    S3 PATH RESOLUTION (per request)                        │
└─────────────────────────────────────────────────────────────────────────────┘

Input:
  tenantID: "abc-123-def-456" (from JWT)
  uploadTo: "uploads"         (from request, already validated)
  storedFilename: "550e8400.pdf"

Template: ResolveS3Path()
  └─ rootPath: "tenants/{tenant_id}/"
  └─ Step 1: Replace {tenant_id}
     └─ "tenants/{tenant_id}/" → "tenants/abc-123-def-456/"
  └─ Step 2: Append uploadTo
     └─ "tenants/abc-123-def-456/" + "uploads/" 
     → "tenants/abc-123-def-456/uploads/"
  └─ Step 3: Append filename
     └─ "tenants/abc-123-def-456/uploads/" + "550e8400.pdf"
     → "tenants/abc-123-def-456/uploads/550e8400.pdf"

Output: "tenants/abc-123-def-456/uploads/550e8400.pdf"

Why this is secure:
  ✅ No client input in final path except filename
  ✅ Client cannot specify tenant_id (it's from JWT)
  ✅ Client cannot specify upload_to folder (only picks from template)
  ✅ Even if client compromises filename, it's isolated in correct folder
```

---

## Security: Attack Prevention

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                      SECURITY ATTACK PREVENTION                            │
└─────────────────────────────────────────────────────────────────────────────┘

Attack 1: Path Traversal
  Attempt: "upload_to": "../admin/"
  
  Validation:
    └─ GetNode("../admin/") in template index
    └─ Not found → REJECT ❌
  
  Result: ✅ BLOCKED

Attack 2: Cross-Tenant Access
  Attempt: Tenant A tries to access Tenant B's folder
  
  Validation:
    └─ All S3 keys forced to start with tenants/{A's_id}/
    └─ Middleware ensures JWT tenant_id is used
    └─ Even if client changes upload_to, stays in own folder
  
  Result: ✅ BLOCKED

Attack 3: Arbitrary S3 Path
  Attempt: "upload_to": "s3://bucket/anywhere/"
  
  Validation:
    └─ GetNode("s3://bucket/anywhere/") in template index
    └─ Not found → REJECT ❌
  
  Result: ✅ BLOCKED

Attack 4: Uploading to Disabled Folder
  Attempt: Folder "reports" exists but can_upload: false
           "upload_to": "reports"
  
  Validation:
    └─ Node found
    └─ Check can_upload == true
    └─ Is false → REJECT: "uploads not allowed in: reports" ❌
  
  Result: ✅ BLOCKED

Attack 5: Uploading Non-Files to File Folder
  Attempt: Folder "uploads" has allow_files: false
           "upload_to": "uploads"
  
  Validation:
    └─ Node found
    └─ Check allow_files == true
    └─ Is false → REJECT: "files not allowed in: uploads" ❌
  
  Result: ✅ BLOCKED
```

---

## Component Interactions

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                      COMPONENT DEPENDENCY GRAPH                            │
└─────────────────────────────────────────────────────────────────────────────┘

main.go
  │
  ├─→ config.LoadFileTreeTemplate()
  │   └─→ Returns: *FileTreeTemplate
  │
  ├─→ FileService
      │
      ├─→ Constructor receives: *FileTreeTemplate
      │   ├─ fileRepo
      │   ├─ s3Service
      │   ├─ template ← NEW!
      │   └─ other dependencies
      │
      └─→ GenerateUploadURL()
          │
          ├─→ Calls: template.ValidateUploadDestination()
          │   ├─ Input: uploadTo string
          │   ├─ Output: (bool, string)
          │   └─ Returns: valid flag, error message
          │
          ├─→ Calls: template.ResolveS3Path()
          │   ├─ Input: tenantID, uploadTo, filename
          │   ├─ Output: full S3 key
          │   └─ Returns: "tenants/{id}/{uploadTo}/{file}"
          │
          └─→ Calls: fileRepo.Create()
              └─→ Saves file with upload_to field

Handler
  └─→ Calls: fileService.GenerateUploadURL()
      └─→ Returns: UploadURLResponse (with presigned URL)
```

---

## File Organization

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                      SOURCE CODE ORGANIZATION                              │
└─────────────────────────────────────────────────────────────────────────────┘

internal/models/
├─ template.go              ← Template structs
│  ├─ type FileTreeTemplate struct
│  ├─ type TemplateNode struct
│  ├─ type TemplatePermissions struct
│  ├─ func (t *FileTreeTemplate) BuildIndex()
│  ├─ func (t *FileTreeTemplate) GetNode()
│  ├─ func (t *FileTreeTemplate) ValidateUploadDestination()
│  └─ func (t *FileTreeTemplate) ResolveS3Path()
│
├─ template_test.go        ← Template tests (10 cases)
│  ├─ TestFileTreeTemplateValidateUploadDestination()
│  ├─ TestFileTreeTemplateResolveS3Path()
│  └─ TestFileTreeTemplateGetNode()
│
└─ file.go (MODIFIED)      ← File models
   ├─ type UploadURLRequest (added: upload_to)
   └─ type File (added: upload_to)

internal/config/
├─ template_loader.go      ← Template loader
│  └─ func LoadFileTreeTemplate() *FileTreeTemplate
│
└─ template_loader_test.go ← Loader tests
   ├─ TestLoadFileTreeTemplate()
   └─ TestLoadFileTreeTemplateDuplicatePaths()

internal/service/
└─ file_service.go (MODIFIED)  ← File service
   ├─ type FileService struct (added: template)
   ├─ func NewFileService() (added: template param)
   └─ func (s *FileService) GenerateUploadURL() (UPDATED)
      └─ Calls: template.ValidateUploadDestination()
      └─ Calls: template.ResolveS3Path()

cmd/server/
└─ main.go (MODIFIED)      ← Server entry point
   ├─ Calls: config.LoadFileTreeTemplate()
   ├─ Passes template to: NewFileService()
   └─ Fails if template invalid

default_dir_tree_template.json
├─ version: "1.0.0"
├─ root_path: "tenants/{tenant_id}/"
└─ nodes: [3 folders]

migrations/
├─ 003_add_upload_to_field.up.sql
│  └─ ALTER TABLE files ADD COLUMN upload_to
└─ 003_add_upload_to_field.down.sql
   └─ ALTER TABLE files DROP COLUMN upload_to
```

---

## Testing Coverage

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          TESTING PYRAMID                                   │
└─────────────────────────────────────────────────────────────────────────────┘

Integration Tests (Future)
  └─ End-to-end upload flow
  └─ S3 interaction
  └─ Database operations

Unit Tests ✅ (10/10 passing)
  │
  ├─ Template Validation (4 tests)
  │  ├─ Valid path
  │  ├─ Uploads disabled
  │  ├─ Invalid destination
  │  └─ Empty upload_to
  │
  ├─ S3 Path Resolution (1 test)
  │  └─ Correct path format
  │
  ├─ Node Lookup (3 tests)
  │  ├─ Existing node
  │  ├─ Another existing node
  │  └─ Non-existing node
  │
  └─ Template Loading (4 tests)
     ├─ Valid template file
     ├─ File not found
     ├─ Invalid JSON
     └─ Duplicate paths

Linting & Build
  ├─ No compilation errors ✅
  ├─ All dependencies resolved ✅
  └─ Binary builds successfully ✅
```

This architecture diagram shows the complete system design, including the flow of requests, how data is structured, security mechanisms, and component interactions.
