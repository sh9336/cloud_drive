# Setting up CORS for Backblaze B2

The frontend is making a direct upload request to the Backblaze B2 storage bucket using a pre-signed URL. By default, Backblaze B2 buckets do not allow Cross-Origin Resource Sharing (CORS), which causes the browser to block the preflight `OPTIONS` request.

You need to correctly configure the CORS rules on your Backblaze bucket so it accepts `PUT` requests (and any other methods needed) from your frontend's domain.

## How to fix this in Backblaze B2

### Option 1: Using the Backblaze Website ( easiest )
1. Log in to your Backblaze account.
2. Go to **Buckets** in the left sidebar.
3. Find your bucket (`trident-file-storage-2026`).
4. Click on **CORS Rules** for that bucket.
5. You can choose to use the "Share everything" template, or create a custom one.
6. A **custom rule** is more secure. Add a rule with these settings:
   - **Allowed Origins:** `*` (or directly specify your frontend URL, e.g., `http://localhost:8084` or your production domain `https://your-frontend.com`)
   - **Allowed Operations:** `PUT`, `GET`, `POST`, `HEAD`
   - **Allowed Headers:** `*` (or specifically `Content-Type`, `Authorization`, `x-amz-*`)
   - **Expose Headers:** `ETag` (important if your frontend needs the ETag to verify upload)
   - **Max Age:** `3600` (this caches the preflight OPTIONS request for 1 hour)
7. Save the rule.

### Option 2: Using the B2 CLI tool
If you have the `b2` CLI installed and authenticated:
```bash
b2 update-bucket --corsRules '[
  {
      "corsRuleName": "allowFrontendUploads",
      "allowedOrigins": [
          "*"
      ],
      "allowedOperations": [
          "b2_upload_file",
          "b2_download_file_by_name",
          "b2_download_file_by_id"
      ],
      "allowedHeaders": [
          "*"
      ],
      "exposeHeaders": [
          "ETag"
      ],
      "maxAgeSeconds": 3600
  }
]' trident-file-storage-2026
```
