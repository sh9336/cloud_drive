# Start services
docker compose up -d

# Check status
docker compose ps

# View logs
docker compose logs -f

# Stop services
docker compose down

# Stop and remove volumes (fresh start)
docker compose down -v

saurabh@SAURABH:~/backend$ # Start your docker compose
   docker compose up -d
   
   # Expose it publicly
   cloudflared tunnel --url http://localhost:8082
    Your quick Tunnel has been created! Visit it at (it may take some time to be reachable):  |
2026-01-19T06:25:53Z INF |  https://generally-extract-fashion-symposium.trycloudflare.com 