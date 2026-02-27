# Sync API Documentation for Device Synchronization

## Overview

This API is designed for high-volume sync devices that need to efficiently synchronize files with the tenant's cloud storage. The sync device uses a long-lived sync token instead of JWT tokens for automated operations.

## Authentication

Use the sync token provided by the admin:
```bash
Authorization: Bearer sync_<tenant_id>_<hash>
```

## Available APIs for Sync Devices

### 1. Get Sync Metadata

**GET** `/api/v1/files/sync/metadata`

Returns metadata for all files in the tenant's storage, optimized for sync operations.

**Usage:** Call this first to get the complete file list from the server.

**Response:**
```json
{
  "success": true,
  "message": "Sync metadata retrieved",
  "data": {
    "files": [
      {
        "path": "schedules/report.pdf",
        "size": 102400,
        "last_modified": "2026-01-02T08:26:13.026469Z",
        "hash": "",
        "mime_type": "application/pdf",
        "file_id": "0c69c775-f6c5-4ac0-8bac-b3fe442584a2"
      },
      {
        "path": "documents/contract.docx",
        "size": 51200,
        "last_modified": "2026-01-02T08:34:59.681154Z",
        "hash": "",
        "mime_type": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
        "file_id": "b17e014c-8b18-48b9-a07e-a513c3d91efb"
      }
    ],
    "total_files": 2,
    "total_size": 153600
  }
}
```

### 2. Generate Batch Download URLs

**POST** `/api/v1/files/sync/download-urls`

Creates presigned download URLs for multiple files in a single request.

**Usage:** Call this with files that need to be downloaded (new or modified).

**Request:**
```json
{
  "files": [
    {"path": "schedules/report.pdf"},
    {"path": "documents/contract.docx"}
  ]
}
```

**Response:**
```json
{
  "success": true,
  "message": "Sync download URLs generated",
  "data": {
    "downloads": [
      {
        "path": "schedules/report.pdf",
        "download_url": "http://localhost:9000/file-storage-bucket/tenants/...",
        "expires_in": 900
      },
      {
        "path": "documents/contract.docx", 
        "download_url": "http://localhost:9000/file-storage-bucket/tenants/...",
        "expires_in": 900
      }
    ]
  }
}
```

## Sync Device Implementation Guide

### Step-by-Step Workflow

#### Step 1: Get Remote File List
```bash
curl -H "Authorization: Bearer sync_..." \
  "http://localhost:8081/api/v1/files/sync/metadata"
```

#### Step 2: Compare with Local Files
The sync device should:
- Compare file paths between remote and local
- Compare file sizes and modification dates
- Identify files that need to be downloaded (new or modified)
- Identify files that no longer exist remotely (optional cleanup)

#### Step 3: Request Download URLs for Changed Files
```bash
curl -X POST \
  -H "Authorization: Bearer sync_..." \
  -H "Content-Type: application/json" \
  -d '{"files": [{"path": "schedules/report.png"}]}' \
  "http://localhost:8081/api/v1/files/sync/download-urls"
```

#### Step 4: Download Files Directly
Use the presigned URLs to download files directly from storage (MinIO/S3):
```bash
curl -o "local/path/schedules/report.png" "http://localhost:9000/file-storage-bucket/..."
```

### Complete Sync Device Example

```python
import requests
import os
import hashlib
import json
from datetime import datetime

class SyncDevice:
    def __init__(self, sync_token, base_url, local_path):
        self.sync_token = sync_token
        self.base_url = base_url
        self.local_path = local_path
        self.headers = {'Authorization': f'Bearer {sync_token}'}
    
    def get_remote_metadata(self):
        """Get complete file list from server"""
        try:
            response = requests.get(
                f'{self.base_url}/api/v1/files/sync/metadata',
                headers=self.headers,
                timeout=30
            )
            response.raise_for_status()
            return response.json()['data']['files']
        except requests.RequestException as e:
            print(f"Error getting remote metadata: {e}")
            return []
    
    def get_local_files(self):
        """Scan local directory for existing files"""
        local_files = {}
        for root, dirs, files in os.walk(self.local_path):
            for file in files:
                full_path = os.path.join(root, file)
                rel_path = os.path.relpath(full_path, self.local_path)
                stat = os.stat(full_path)
                local_files[rel_path] = {
                    'size': stat.st_size,
                    'modified': stat.st_mtime,
                    'path': full_path
                }
        return local_files
    
    def compare_files(self, remote_files, local_files):
        """Compare remote and local files to identify what needs syncing"""
        files_to_download = []
        
        for remote_file in remote_files:
            path = remote_file['path']
            
            # Check if file exists locally
            if path not in local_files:
                # New file - needs download
                files_to_download.append({'path': path})
                print(f"New file detected: {path}")
            else:
                local_file = local_files[path]
                
                # Compare sizes (simple change detection)
                if local_file['size'] != remote_file['size']:
                    files_to_download.append({'path': path})
                    print(f"Modified file detected: {path}")
                else:
                    print(f"File up to date: {path}")
        
        # Optionally identify files that no longer exist remotely
        remote_paths = {f['path'] for f in remote_files}
        for local_path in local_files:
            if local_path not in remote_paths:
                print(f"Local file no longer exists remotely: {local_path}")
                # Optionally delete local file
        
        return files_to_download
    
    def download_files(self, files_to_download):
        """Download files using batch download URLs"""
        if not files_to_download:
            print("No files to download")
            return
        
        print(f"Requesting download URLs for {len(files_to_download)} files...")
        
        # Get download URLs
        try:
            response = requests.post(
                f'{self.base_url}/api/v1/files/sync/download-urls',
                headers=self.headers,
                json={'files': files_to_download},
                timeout=30
            )
            response.raise_for_status()
            downloads = response.json()['data']['downloads']
            
            # Download files
            successful_downloads = 0
            for download in downloads:
                if self.download_single_file(download):
                    successful_downloads += 1
            
            print(f"Successfully downloaded {successful_downloads}/{len(downloads)} files")
            
        except requests.RequestException as e:
            print(f"Error getting download URLs: {e}")
    
    def download_single_file(self, download_info):
        """Download a single file using presigned URL"""
        path = download_info['path']
        url = download_info['download_url']
        
        # Create directory if needed
        full_path = os.path.join(self.local_path, path)
        os.makedirs(os.path.dirname(full_path), exist_ok=True)
        
        try:
            # Download file
            response = requests.get(url, timeout=60)
            response.raise_for_status()
            
            # Save file
            with open(full_path, 'wb') as f:
                f.write(response.content)
            
            print(f"Downloaded: {path}")
            return True
            
        except requests.RequestException as e:
            print(f"Error downloading {path}: {e}")
            return False
    
    def sync(self):
        """Main sync operation"""
        print("Starting sync operation...")
        
        # Step 1: Get remote metadata
        remote_files = self.get_remote_metadata()
        if not remote_files:
            print("No remote files found or error occurred")
            return
        
        print(f"Found {len(remote_files)} remote files")
        
        # Step 2: Get local files
        local_files = self.get_local_files()
        print(f"Found {len(local_files)} local files")
        
        # Step 3: Compare and identify files to download
        files_to_download = self.compare_files(remote_files, local_files)
        
        # Step 4: Download files
        if files_to_download:
            self.download_files(files_to_download)
        else:
            print("All files are up to date")
        
        print("Sync operation completed")

# Usage Example
if __name__ == "__main__":
    # Configuration
    SYNC_TOKEN = "sync_80a27b4d-55f8-48f8-acb4-6c9e2da759fe_81f9cf90d3e60b055482937b476cc52bf756a68bbf09d9656cd89d546d6e9363"
    BASE_URL = "http://localhost:8081"
    LOCAL_PATH = "./sync_storage"
    
    # Create and run sync device
    sync_device = SyncDevice(SYNC_TOKEN, BASE_URL, LOCAL_PATH)
    
    # Run sync (can be called periodically)
    sync_device.sync()
```

### Production Deployment Considerations

#### 1. Error Handling
- Network timeouts and retries
- Invalid token handling
- Storage service unavailability
- Disk space management

#### 2. Performance Optimization
- **Batch Processing**: Always use batch download URLs
- **Parallel Downloads**: Download multiple files concurrently
- **Incremental Sync**: Track last sync time to avoid full scans
- **Compression**: Use gzip for API requests if supported

#### 3. Security
- **Token Protection**: Store sync tokens securely
- **URL Validation**: Validate download URLs before use
- **File Verification**: Verify file integrity after download

#### 4. Monitoring
- **Sync Statistics**: Track files synced, bandwidth used
- **Error Logging**: Log failed downloads for retry
- **Health Checks**: Monitor sync token validity

### API Limits and Best Practices

| Feature | Limit | Recommendation |
|---------|-------|----------------|
| **Batch Size** | 100 files per request | Use smaller batches for reliability |
| **URL Expiration** | 900 seconds (15 minutes) | Download files promptly |
| **Rate Limiting** | Per-token limits | Implement exponential backoff |
| **File Size** | Configurable per tenant | Check available disk space |

### Troubleshooting

#### Common Issues

1. **"internal server error"**
   - Check if sync token is valid and not revoked
   - Verify file paths exist in metadata

2. **Empty downloads array**
   - File paths may not match exactly
   - Files might not have completed upload status

3. **URL expiration**
   - Download files within 15 minutes
   - Request fresh URLs if needed

4. **Path not found**
   - Use exact paths from metadata endpoint
   - Check for leading/trailing spaces

#### Debug Mode
Enable verbose logging:
```python
import logging
logging.basicConfig(level=logging.DEBUG)
```

## Security Features

- **Tenant Isolation**: Sync tokens only access their tenant's files
- **Permission Control**: Sync tokens have configurable read/write/delete permissions
- **Token Expiration**: Long-lived but revocable sync tokens
- **Audit Logging**: All sync operations are logged
- **Rate Limiting**: Prevents abuse of sync endpoints

## Monitoring

Admins can monitor sync token usage:
- View sync token statistics
- Track request counts and bandwidth usage
- Revoke compromised tokens
- Monitor active sync devices
