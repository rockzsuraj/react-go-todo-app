# Docker Setup

## Quick Start

1. Copy environment file:
```bash
cp .env.example .env
```

2. Start all services with hot reload (development):
```bash
docker-compose up --build
```

## Services

- **Frontend**: http://localhost:3000 (with hot reload)
- **Backend**: http://localhost:8080 (with hot reload)
- **Database**: localhost:5432
- **Redis**: localhost:6379

## Environment Variables

Key variables in `.env`:
- `ENV`: development or production (determines Docker target)
- `NODE_ENV`: development or production
- `REACT_APP_API_URL`: Frontend API URL
- `DB_NAME`, `DB_USER`, `DB_PASSWORD`: Database credentials
- `REDIS_ADDR`: Redis connection string

## Development vs Production

### Development (Hot Reload)
```bash
# Set environment variables
ENV=development
NODE_ENV=development

# Start with hot reload
docker-compose up --build
```

Features:
- **Backend**: Air hot reload for Go code changes
- **Frontend**: React development server with hot reload
- **Volumes**: Source code mounted for live updates

### Production
```bash
# Set environment variables
ENV=production
NODE_ENV=production

# Start production build
docker-compose up --build
```

Features:
- **Backend**: Compiled binary (no hot reload)
- **Frontend**: Nginx serving static build
- **Optimized**: Production-ready containers

## Hot Reload Details

### Backend (Air)
- Monitors `*.go` files in `apps/api/`
- Auto-restarts on file changes
- Configuration in `.air.toml`

### Frontend (React)
- React development server
- Hot Module Replacement (HMR)
- Live reload on component changes

## Cleanup

```bash
docker-compose down -v
```
