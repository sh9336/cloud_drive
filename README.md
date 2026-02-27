# 🚀 Cloud Drive Backend

A professional, high-performance, multi-tenant file storage and synchronization engine built with **Go**. Designed for high-volume device synchronization with enforced folder structures and secure S3-compatible storage.

---

## ✨ Key Features

- **👥 Multi-Tenancy**: Complete isolation between tenants with hierarchical admin control.
- **📂 Structured storage**: Enforce specific folder structures (uploads, assets, etc.) via JSON-based file tree templates.
- **🔄 Sync Engine**: Optimized for device synchronization using Revocable Sync Tokens (High-volume, low-overhead).
- **🛡️ Security First**:
  - JWT-based authentication for admins and tenants.
  - Granular permissions (Read/Write/Delete) for sync tokens.
  - Secure Direct-to-S3 uploads via Presigned URLs (No proxying through backend).
- **📊 Audit Logging**: Comprehensive tracking of all administrative and file-level actions.
- **📦 S3-Compatible**: Works with AWS S3, Backblaze B2, Cloudflare R2, and MinIO.

---

## 🛠️ Tech Stack

- **Language**: Go 1.24 (Standard library focus)
- **Web Framework**: Gin Gonic
- **Database**: PostgreSQL (with migrations support)
- **Storage**: AWS SDK v2 (S3 Compatible)
- **Auth**: JWT (HS256)
- **Containerization**: Docker & Docker Compose

---

## 🚀 Getting Started (Local Development)

### 1. Prerequisites
- Docker & Docker Compose
- Go 1.24+ (Optional, if running outside Docker)

### 2. Launch Development Environment
The easiest way to start is using Docker Compose, which spins up the Backend, Postgres, and MinIO (Local S3).

```bash
# Clone the repository
git clone https://github.com/sh9336/cloud_drive.git
cd cloud_drive

# Start everything
make fresh
```

### 3. Verify Setup
Check if the health endpoint responds:
```bash
curl http://localhost:8080/health
```

---

## ⚙️ Configuration

The application uses environment variables for configuration. See `.env.dev` for local development or `.env.prod.example` for production templates.

| Variable | Description |
| :--- | :--- |
| `ENVIRONMENT` | `production` or `development` |
| `DATABASE_URL` | Full Postgres DSN (e.g., Neon.tech) |
| `AWS_ENDPOINT` | The S3 API endpoint |
| `S3_BUCKET_NAME` | Target bucket for file storage |
| `MAX_FILE_SIZE` | Maximum upload size in bytes (default: 100MB) |

---

## 📁 File Tree Templates

This backend uses a unique template system to enforce client-side structure. The template defines:
- **Required Folders**: Folders that must exist for every tenant.
- **Permissions**: Which folders allow uploads vs. only listing.

The template is served via `GET /api/v1/config/template` and is validated server-side during upload requests.

---

## 🚢 Production Deployment

For production deployment (Render + Neon + Backblaze), refer to the detailed **[Deployment Guide](deployment.md)**.

### 🛡️ Security Note
**Never** commit your `.env.prod` file. Use the environment secret management system provided by your cloud host (e.g., Render Dashboard).

---

## 📜 License
This project is licensed under the MIT License.