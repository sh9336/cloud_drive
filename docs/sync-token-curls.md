# Sync token curl examples
Base: http://localhost:8081
Replace <ADMIN_TOKEN>, <TENANT_TOKEN>, <TOKEN_ID>, <TENANT_ID> as needed.

Admin routes

Create a sync token (POST /api/v1/admin/sync-tokens)
curl --location 'http://localhost:8081/api/v1/admin/sync-tokens' \
--header 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiZjMxODUwOTUtYTg4Zi00ZjBhLWJiYzQtMWY3OWM2M2I0MmU1IiwiZW1haWwiOiJzYXVyYWJoMjVkZWNlbTIwMjBAZ21haWwuY29tIiwidXNlcl90eXBlIjoiYWRtaW4iLCJpc3MiOiJmaWxlLXN0b3JhZ2UtYXBpIiwic3ViIjoiZjMxODUwOTUtYTg4Zi00ZjBhLWJiYzQtMWY3OWM2M2I0MmU1IiwiZXhwIjoxNzY2ODE4NTMxLCJpYXQiOjE3NjY4MTc2MzF9.iIrHFnS9UunafkVTRoVuU4GHaRaaA99dJAolFC1CS0s' \
--header 'Content-Type: application/json' \
--data '{"tenant_id":"f00e2e71-b157-4dfc-b953-e03e3975dc76","name":"Scheduler-Sync","can_read":true,"can_write":false,"can_delete":false,"expires_in_days":365}'

List all sync tokens (GET /api/v1/admin/sync-tokens)
curl -H "Authorization: Bearer <ADMIN_TOKEN>" "http://localhost:8081/api/v1/admin/sync-tokens"

Get a sync token by id (GET /api/v1/admin/sync-tokens/:id)
curl -H "Authorization: Bearer <ADMIN_TOKEN>" "http://localhost:8081/api/v1/admin/sync-tokens/<TOKEN_ID>"

List a tenant's sync tokens (GET /api/v1/admin/tenants/:id/sync-tokens)
curl -H "Authorization: Bearer <ADMIN_TOKEN>" "http://localhost:8081/api/v1/admin/tenants/<TENANT_ID>/sync-tokens"

Rotate a sync token (POST /api/v1/admin/sync-tokens/:id/rotate)
curl -X POST "http://localhost:8081/api/v1/admin/sync-tokens/<TOKEN_ID>/rotate" \
  -H "Authorization: Bearer <ADMIN_TOKEN>"

Delete a sync token (DELETE /api/v1/admin/sync-tokens/:id)
curl -X DELETE "http://localhost:8081/api/v1/admin/sync-tokens/<TOKEN_ID>" \
  -H "Authorization: Bearer <ADMIN_TOKEN>"

Get sync token stats (GET /api/v1/admin/sync-tokens/:id/stats)
curl -H "Authorization: Bearer <ADMIN_TOKEN>" "http://localhost:8081/api/v1/admin/sync-tokens/<TOKEN_ID>/stats"

Tenant routes

List tenant's tokens (GET /api/v1/tenant/sync-tokens)
curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMDlhYmE0OWEtNTcxOC00YmFlLWExODctZjk0YWYwZWRmYTkyIiwiZW1haWwiOiJ0ZW5hbnQxQGV4YW1wbGUuY29tIiwidXNlcl90eXBlIjoidGVuYW50IiwiaXNzIjoiZmlsZS1zdG9yYWdlLWFwaSIsInN1YiI6IjA5YWJhNDlhLTU3MTgtNGJhZS1hMTg3LWY5NGFmMGVkZmE5MiIsImV4cCI6MTc2NjgxODc3NywiaWF0IjoxNzY2ODE3ODc3fQ.4lkpwXllMn2T0WyWmWPQmHaPTtP8f9tE8LKWMKzpKFM" "http://localhost:8081/api/v1/tenant/sync-tokens"

Get tenant token by id (GET /api/v1/tenant/sync-tokens/:id)
curl -H "Authorization: Bearer <TENANT_TOKEN>" "http://localhost:8081/api/v1/tenant/sync-tokens/<TOKEN_ID>"

Get tenant token stats (GET /api/v1/tenant/sync-tokens/:id/stats)
curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMDlhYmE0OWEtNTcxOC00YmFlLWExODctZjk0YWYwZWRmYTkyIiwiZW1haWwiOiJ0ZW5hbnQxQGV4YW1wbGUuY29tIiwidXNlcl90eXBlIjoidGVuYW50IiwiaXNzIjoiZmlsZS1zdG9yYWdlLWFwaSIsInN1YiI6IjA5YWJhNDlhLTU3MTgtNGJhZS1hMTg3LWY5NGFmMGVkZmE5MiIsImV4cCI6MTc2NjgxODc3NywiaWF0IjoxNzY2ODE3ODc3fQ.4lkpwXllMn2T0WyWmWPQmHaPTtP8f9tE8LKWMKzpKFM" "http://localhost:8081/api/v1/tenant/sync-tokens/6fba9f1a-53fd-42db-9f32-6ffa2cd99303/stats"