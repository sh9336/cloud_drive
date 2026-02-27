# Deployment Guide (Free Tier)

This guide outlines how to deploy the File Storage Backend for free using modern cloud providers.

## 🚀 Recommended Stack

| Component | Provider | Tier |
| :--- | :--- | :--- |
| **Backend (Go)** | [Render](https://render.com/) | Free (Web Service) |
| **Database (Postgres)** | [Neon.tech](https://neon.tech/) | Free (Shared) |
| **Storage (S3)** | [Cloudflare R2](https://www.cloudflare.com/products/r2/) | Free (10GB/mo) |

---

## 1. Database Setup (Neon.tech)

1.  **Register:** Create an account at [Neon.tech](https://neon.tech/).
2.  **Create Project:** Create a new project named `file-storage`.
3.  **Get Connection String:** Copy the connection string. It will look like:
    `postgres://user:password@ep-cool-darkness-123456.us-east-2.aws.neon.tech/neondb?sslmode=require`
4.  **Run Migrations:**
    *   Connect to your Neon database using a tool like TablePlus, DBeaver, or `psql`.
    *   Execute the SQL files in the `migrations/` directory in order:
        1. `001_initial_schema.up.sql`
        2. `002_sync_tokens.up.sql`
        3. `003_add_upload_to_field.up.sql`

---

## 2. Storage Setup (Cloudflare R2)

1.  **Register:** Create an account at [Cloudflare](https://www.cloudflare.com/).
2.  **Enable R2:** Navigate to **R2** in the sidebar.
3.  **Create Bucket:** Create a bucket named `file-storage-bucket`.
4.  **Get API Keys:**
    *   Click on **Manage R2 API Tokens**.
    *   **Create API token** with `Object Read & Write` permissions.
    *   Save the **Access Key ID**, **Secret Access Key**, and the **Jurisdictional Specific Endpoint** (the S3 API URL).

---

## 3. Backend Deployment (Render)

1.  **Push to GitHub:** Ensure your code is pushed to a private GitHub or GitLab repository.
2.  **Create Web Service:**
    *   Log in to [Render](https://dashboard.render.com/).
    *   Click **New +** -> **Web Service**.
    *   Connect your repository.
3.  **Configuration:**
    *   **Name:** `file-storage-backend`
    *   **Environment:** `Docker` (Render will automatically detect your `Dockerfile`)
    *   **Region:** Choose the one closest to your Neon DB region (usually `us-east-1` or `frankfurt`).
4.  **Environment Variables:** Add the following under the **Environment** tab:

| Variable | Value |
| :--- | :--- |
| `ENVIRONMENT` | `production` |
| `DATABASE_URL` | `postgres://user:password@ep-xxx...` (Paste your full Neon string here) |
| `AWS_REGION` | `auto` |
| `AWS_ACCESS_KEY_ID` | (From Cloudflare R2) |
| `AWS_SECRET_ACCESS_KEY` | (From Cloudflare R2) |
| `AWS_ENDPOINT` | (The S3 API URL from Cloudflare R2) |
| `S3_BUCKET_NAME` | `file-storage-bucket` |
| `JWT_ACCESS_SECRET` | (Generate a random 32+ char string) |
| `JWT_REFRESH_SECRET` | (Generate a random 32+ char string) |
| `SERVER_PORT` | `8080` |

---

## ⚠️ Important Production Notes

*   **Cold Starts:** Render's free tier spins down after 15 minutes of inactivity. The first request after a break will take ~30 seconds to respond.
*   **Database Limits:** Neon's free tier has a storage limit (usually 500MB). Monitor your `audit_logs` table size.
*   **Secrets:** Never commit your production `.env` files. Always use the Render dashboard to manage secrets.
*   **Health Checks:** Your Dockerfile already has a `HEALTHCHECK`. Render will use this to ensure the app is running before routing traffic.

## 🛠 Local vs Production
*   **Local:** Uses Docker Compose with PG and MinIO.
*   **Production:** Uses Neon (PG) and Cloudflare R2 (S3).
The Go code automatically switches behavior based on the `ENVIRONMENT=production` variable and the provided `AWS_ENDPOINT`.
