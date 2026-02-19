# compose-stack

A production-ready Docker Compose reference architecture with a Go API, PostgreSQL, Redis, Nginx reverse proxy, Prometheus metrics, and Grafana dashboards — fully wired with health checks, profiles, and one-command startup.

## Architecture

```
                    ┌─────────────────────────────────────────┐
                    │              Docker Network              │
                    │                                         │
  Client ──────►   │  ┌─────────┐    ┌─────────┐            │
    :80/:443       │  │  Nginx  │───►│   App   │            │
                    │  │ (proxy) │    │  (Go)   │            │
                    │  └─────────┘    └────┬────┘            │
                    │   frontend      ├────┴────┐            │
                    │   network       │         │            │
                    │            ┌────▼───┐ ┌───▼────┐       │
                    │            │Postgres│ │ Redis  │       │
                    │            │  :5432 │ │ :6379  │       │
                    │            └────────┘ └────────┘       │
                    │             backend network             │
                    │                                         │
                    │  ┌────────────┐    ┌─────────┐         │
                    │  │ Prometheus │───►│ Grafana │         │
                    │  │   :9090   │    │  :3000  │         │
                    │  └────────────┘    └─────────┘         │
                    │           monitoring network            │
                    └─────────────────────────────────────────┘
```

## Quick Start

```bash
# Clone the repository
git clone https://github.com/devaloi/compose-stack.git
cd compose-stack

# Configure environment
cp .env.example .env

# Generate self-signed SSL certificates
make certs

# Start all services (core + monitoring + dev ports)
make up-all

# Or start core stack only (app, postgres, redis, nginx)
make up
```

## Services

| Service | Image | Host Port | Internal Port | Profile |
|---------|-------|-----------|---------------|---------|
| app | Custom (Go) | — | 8080 | default |
| postgres | postgres:16-alpine | 5432 (dev) | 5432 | default |
| redis | redis:7-alpine | 6379 (dev) | 6379 | default |
| nginx | nginx:alpine | 80, 443 | 80, 443 | default |
| prometheus | prom/prometheus | 9090 (dev) | 9090 | monitoring |
| grafana | grafana/grafana | 3000 | 3000 | monitoring |

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check (DB + Redis status) |
| GET | `/api/items` | List all items |
| POST | `/api/items` | Create item (`{"name":"...", "description":"..."}`) |
| GET | `/api/items/{id}` | Get item by ID |
| DELETE | `/api/items/{id}` | Delete item |
| GET | `/api/cache/{key}` | Get cached value |
| PUT | `/api/cache/{key}` | Set cached value (`{"value":"...", "ttl":60}`) |
| GET | `/metrics` | Prometheus metrics |

## Makefile Commands

| Command | Description |
|---------|-------------|
| `make up` | Start core stack |
| `make up-all` | Start all services (core + monitoring + dev ports) |
| `make down` | Stop all services |
| `make build` | Build Docker images |
| `make logs` | Follow container logs |
| `make ps` | Show running containers |
| `make clean` | Stop and remove all volumes |
| `make certs` | Generate self-signed SSL certificates |
| `make test` | Run Go tests locally |
| `make shell-app` | Shell into app container |
| `make shell-db` | Connect to PostgreSQL with psql |

## Monitoring

Start with the monitoring profile:

```bash
make up-all
```

- **Grafana**: http://localhost:3000 (anonymous access enabled)
- **Prometheus**: http://localhost:9090 (via dev override)

The Grafana dashboard auto-provisions on first startup with panels for:
- Request rate by method/path/status
- Error rate (5xx responses)
- Latency percentiles (P50, P95, P99)
- Active HTTP connections
- Database connection pool usage
- Requests breakdown by endpoint

## SSL Certificates

Generate self-signed certificates for local development:

```bash
make certs
```

This creates `nginx/certs/server.crt` and `nginx/certs/server.key`. For production, replace these with real certificates from a CA (e.g., Let's Encrypt) and mount them at the same paths.

## Network Topology

| Network | Services | Purpose |
|---------|----------|---------|
| frontend | nginx, app | Public-facing traffic |
| backend | app, postgres, redis | Database and cache access |
| monitoring | app, prometheus, grafana | Metrics collection |

Services only connect to the networks they need. PostgreSQL and Redis are not accessible from the frontend network.

## Environment Variables

See [.env.example](.env.example) for all available configuration options.

| Variable | Default | Description |
|----------|---------|-------------|
| `POSTGRES_DB` | compose_stack | Database name |
| `POSTGRES_USER` | app | Database user |
| `POSTGRES_PASSWORD` | changeme | Database password |
| `DB_URL` | (see .env.example) | Full PostgreSQL connection string |
| `REDIS_URL` | redis://redis:6379/0 | Redis connection string |
| `APP_PORT` | 8080 | Application listen port |
| `APP_LOG_LEVEL` | info | Log level |
| `NGINX_HOST` | localhost | Nginx server name |
| `GF_SECURITY_ADMIN_PASSWORD` | admin | Grafana admin password |

## Production Considerations

- Replace self-signed certificates with real ones from a CA
- Change all default passwords in `.env`
- Set `GF_AUTH_ANONYMOUS_ENABLED=false` in Grafana
- Remove dev port exposure (don't use `docker-compose.dev.yml`)
- Add resource limits (`deploy.resources`) to all services
- Configure PostgreSQL `max_connections` and connection pool sizes
- Enable Redis persistence configuration (`appendfsync`)
- Set up log aggregation (ELK, Loki, etc.)
- Add backup strategy for PostgreSQL data volume

## Tech Stack

- **Go 1.22+** — stdlib `net/http` with Go 1.22 routing
- **PostgreSQL 16** — primary data store
- **Redis 7** — caching layer
- **Nginx** — reverse proxy, SSL termination, rate limiting
- **Prometheus** — metrics collection
- **Grafana** — metrics visualization
- **Docker Compose v2** — container orchestration with profiles

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) with Docker Compose v2
- [Go 1.22+](https://go.dev/dl/) (for local development and testing)
- OpenSSL (for certificate generation)

## License

[MIT](LICENSE)
