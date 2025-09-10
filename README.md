# Lean View - PQ Devnet Visualizer

## Running with Docker Compose

```bash
# Start all services
docker compose up -d

# View logs
docker compose logs -f

# Stop all services
docker compose down

# Rebuild and restart
docker compose up -d --build
```

Access the application:

- Frontend: http://localhost:3000
- Backend API: http://localhost:8080
- API via Frontend proxy: http://localhost:3000/api/

### Configuration

The backend uses `config/docker.config.yml` when running in Docker. This file configures:

- **Lean node endpoints**: Uses `host.docker.internal:5052` to reach services on your host machine
- **Database path**: Points to `/app/data/lean-view.sqlite` for container environment

To modify lean node endpoints for Docker, edit `config/docker.config.yml`:

```yaml
leanapi:
  endpoints:
    - name: "local"
      url: "http://host.docker.internal:5052" # For local development
      # url: "http://192.168.1.100:5052"       # For network services
      # url: "http://node.example.com:5052"    # For remote services
```

The default `backend/config/default.config.yml` remains configured for local development with `localhost:5052`.

## Running with Docker (Individual Containers)

### Backend

```bash
# Build the backend image
cd backend
docker build -f Dockerfile.backend -t leanview-backend .

# Run the backend container
docker run -d \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  -e DATABASE_FILE=/app/data/lean-view.sqlite \
  --name leanview-backend \
  leanview-backend

# Check logs
docker logs leanview-backend

# Check health
curl http://localhost:8080/health
```

### Frontend

```bash
# Build the frontend image
cd frontend
docker build -f Dockerfile.frontend -t leanview-frontend .
```
