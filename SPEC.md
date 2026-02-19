# D01: compose-stack — Production Docker Compose Stack

**Catalog ID:** D01 | **Size:** S | **Language:** Docker / YAML
**Repo name:** `compose-stack`
**One-liner:** A production-ready Docker Compose reference architecture with a Go API, PostgreSQL, Redis, Nginx reverse proxy, Prometheus metrics, and Grafana dashboards — fully wired with health checks, profiles, and one-command startup.

---

## Why This Stands Out

- **Complete reference architecture** — not just "app + database" but a full production topology with monitoring
- **Multi-stage Go build** — final image under 20MB, no compiler or source in production
- **Nginx reverse proxy** — SSL termination with self-signed certs, rate limiting, gzip, security headers
- **Health checks everywhere** — every service has a health check; compose waits for dependencies
- **Prometheus + Grafana** — app exposes `/metrics`, Prometheus scrapes, Grafana dashboards auto-provisioned
- **Dev vs Prod profiles** — `docker compose --profile dev` vs `--profile prod` for different configurations
- **Makefile for operations** — `make up`, `make down`, `make logs`, `make ps`, `make build`, `make clean`
- **Docker networking best practices** — named networks, service isolation, minimal port exposure
- **Everything documented** — `.env.example` with every variable, architecture diagram in README

---

## Architecture

```
compose-stack/
├── app/
│   ├── main.go                    # Go API: health, CRUD, metrics endpoint
│   ├── handler.go                 # HTTP handlers
│   ├── handler_test.go
│   ├── middleware.go              # Logging, request ID, Prometheus metrics
│   ├── store.go                   # PostgreSQL + Redis data layer
│   ├── store_test.go
│   ├── go.mod
│   └── go.sum
├── nginx/
│   ├── nginx.conf                 # Main nginx config
│   ├── conf.d/
│   │   ├── default.conf           # Reverse proxy to app
│   │   └── ssl.conf               # SSL configuration
│   ├── certs/
│   │   └── generate-certs.sh      # Self-signed cert generation script
│   └── security-headers.conf      # HSTS, X-Frame, CSP, etc.
├── prometheus/
│   └── prometheus.yml             # Scrape config: app + node-exporter targets
├── grafana/
│   ├── provisioning/
│   │   ├── datasources/
│   │   │   └── prometheus.yml     # Auto-provision Prometheus data source
│   │   └── dashboards/
│   │       ├── dashboard.yml      # Dashboard provisioning config
│   │       └── app-dashboard.json # Pre-built app metrics dashboard
│   └── grafana.ini                # Grafana config (anonymous access for demo)
├── db/
│   └── init.sql                   # Database initialization (schema + seed data)
├── docker-compose.yml             # All services, profiles, networks, volumes
├── Dockerfile                     # Multi-stage Go build
├── .dockerignore
├── Makefile                       # Operations: up, down, build, logs, ps, clean, certs
├── .env.example                   # All environment variables documented
├── .gitignore
├── LICENSE
└── README.md
```

---

## Services

| Service | Image | Port (host) | Port (internal) | Profile | Health Check |
|---------|-------|-------------|----------------|---------|-------------|
| `app` | Custom (Go) | — | 8080 | all | `GET /health` |
| `postgres` | postgres:16-alpine | 5432 (dev only) | 5432 | all | `pg_isready` |
| `redis` | redis:7-alpine | — | 6379 | all | `redis-cli ping` |
| `nginx` | nginx:alpine | 80, 443 | 80, 443 | all | `curl -f http://localhost/health` |
| `prometheus` | prom/prometheus | 9090 (dev only) | 9090 | monitoring | `wget --spider http://localhost:9090/-/healthy` |
| `grafana` | grafana/grafana | 3000 (dev only) | 3000 | monitoring | `wget --spider http://localhost:3000/api/health` |

---

## Go API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check: returns DB + Redis status |
| GET | `/api/items` | List items (PostgreSQL) |
| POST | `/api/items` | Create item |
| GET | `/api/items/:id` | Get item by ID |
| DELETE | `/api/items/:id` | Delete item |
| GET | `/api/cache/:key` | Get cached value (Redis) |
| PUT | `/api/cache/:key` | Set cached value (Redis, with TTL) |
| GET | `/metrics` | Prometheus metrics (request count, duration histogram, active connections) |

---

## Prometheus Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `http_requests_total` | Counter | Total HTTP requests by method, path, status |
| `http_request_duration_seconds` | Histogram | Request latency distribution |
| `http_active_connections` | Gauge | Currently active connections |
| `db_connections_active` | Gauge | Active PostgreSQL connections |
| `redis_commands_total` | Counter | Total Redis commands |

---

## Grafana Dashboard Panels

| Panel | Visualization | Query |
|-------|--------------|-------|
| Request Rate | Time series | `rate(http_requests_total[5m])` |
| Error Rate | Stat | `rate(http_requests_total{status=~"5.."}[5m])` |
| Latency P50/P95/P99 | Time series | `histogram_quantile(0.95, http_request_duration_seconds_bucket)` |
| Active Connections | Gauge | `http_active_connections` |
| DB Connections | Stat | `db_connections_active` |
| Requests by Endpoint | Bar chart | `sum by(path)(rate(http_requests_total[5m]))` |

---

## Tech Stack

| Component | Choice |
|-----------|--------|
| App | Go 1.22+ (stdlib net/http) |
| Database | PostgreSQL 16 |
| Cache | Redis 7 |
| Reverse proxy | Nginx (alpine) |
| Monitoring | Prometheus + Grafana |
| Container | Docker Compose v2 (with profiles) |
| Build | Multi-stage Dockerfile |
| Metrics | prometheus/client_golang |

---

## Phased Build Plan

### Phase 1: Go Application

**1.1 — Project setup**
- `app/` directory with `go mod init github.com/devaloi/compose-stack`
- `main.go`: HTTP server with graceful shutdown
- `handler.go`: health check endpoint returning JSON `{"status": "ok"}`
- Makefile: `make build`, `make test`, `make run` (local), `make up`, `make down` (Docker)

**1.2 — PostgreSQL integration**
- `store.go`: connect to PostgreSQL, connection pool, ping check
- `db/init.sql`: create `items` table (id, name, description, created_at)
- CRUD handlers: list, create, get by ID, delete
- Parameterized queries, proper error handling (404, 400, 500)
- Tests with mocked store interface

**1.3 — Redis integration**
- Add Redis client to store
- Cache get/set handlers with configurable TTL
- Redis health check in `/health` response
- Tests for cache operations

**1.4 — Prometheus metrics**
- Import `prometheus/client_golang`
- Middleware: count requests, track duration histogram, track active connections
- Expose `/metrics` endpoint
- Custom collector for DB connection stats
- Tests: metrics increment on requests

### Phase 2: Docker Setup

**2.1 — Multi-stage Dockerfile**
- Stage 1: `golang:1.22-alpine` — build with `CGO_ENABLED=0`
- Stage 2: `alpine:3.19` — copy binary, add CA certs, non-root user
- Final image target: < 20MB
- `.dockerignore`: exclude `.git`, `*.md`, test files, `vendor/`

**2.2 — Docker Compose services**
- `docker-compose.yml` with all 6 services
- Named networks: `frontend` (nginx + app), `backend` (app + postgres + redis), `monitoring` (prometheus + grafana)
- Named volumes: `postgres_data`, `redis_data`, `grafana_data`
- Service dependencies with `depends_on` + health check conditions
- Environment variables from `.env` file

**2.3 — Health checks**
- `app`: HTTP GET `/health` every 10s, 3 retries
- `postgres`: `pg_isready` every 5s
- `redis`: `redis-cli ping` every 5s
- `nginx`: `curl -f http://localhost/health` every 10s
- `prometheus`: `wget --spider` to `/-/healthy`
- `grafana`: `wget --spider` to `/api/health`
- All with `start_period`, `interval`, `timeout`, `retries`

**2.4 — Profiles**
- Default profile: app, postgres, redis, nginx (core stack)
- `monitoring` profile: adds prometheus + grafana
- `dev` profile: expose postgres (5432) and redis (6379) ports to host
- Usage: `docker compose --profile monitoring --profile dev up`

### Phase 3: Nginx Reverse Proxy

**3.1 — Nginx configuration**
- Reverse proxy to `app:8080`
- Upstream definition with keepalive connections
- Proxy headers: `X-Real-IP`, `X-Forwarded-For`, `X-Forwarded-Proto`
- Static file serving from `/static/` (if applicable)
- Gzip compression for JSON and text responses
- Rate limiting: 10 req/s per IP with burst

**3.2 — SSL termination**
- `generate-certs.sh`: create self-signed cert + key with `openssl`
- SSL config: TLS 1.2+, modern cipher suite, OCSP stapling (where applicable)
- HTTP → HTTPS redirect
- Security headers: HSTS, X-Content-Type-Options, X-Frame-Options, CSP
- Certs mounted as volume (not baked into image)

### Phase 4: Monitoring Stack

**4.1 — Prometheus config**
- `prometheus.yml`: scrape app on `app:8080/metrics` every 15s
- Job names: `app`, `prometheus` (self-monitoring)
- Relabel configs for clean labels

**4.2 — Grafana auto-provisioning**
- Data source: Prometheus at `http://prometheus:9090`
- Dashboard JSON: 6 panels (request rate, error rate, latency, connections, DB, endpoint breakdown)
- Anonymous access enabled for demo (no login required)
- Dashboard auto-loads on first startup

### Phase 5: Documentation & Polish

**5.1 — .env.example**
- Every variable documented with comments:
  ```env
  # PostgreSQL
  POSTGRES_DB=compose_stack
  POSTGRES_USER=app
  POSTGRES_PASSWORD=changeme
  DB_URL=postgres://app:changeme@postgres:5432/compose_stack?sslmode=disable

  # Redis
  REDIS_URL=redis://redis:6379/0

  # App
  APP_PORT=8080
  APP_LOG_LEVEL=info

  # Nginx
  NGINX_HOST=localhost

  # Grafana
  GF_SECURITY_ADMIN_PASSWORD=admin
  ```

**5.2 — Makefile**
- `make up` — `docker compose up -d`
- `make up-all` — `docker compose --profile monitoring --profile dev up -d`
- `make down` — `docker compose down`
- `make build` — `docker compose build`
- `make logs` — `docker compose logs -f`
- `make ps` — `docker compose ps`
- `make clean` — stop + remove volumes
- `make certs` — generate self-signed certificates
- `make test` — run Go tests locally
- `make shell-app` — exec into app container
- `make shell-db` — exec into postgres with psql

**5.3 — README.md**
- Architecture diagram (ASCII art showing service topology)
- Quick start: `cp .env.example .env && make certs && make up-all`
- Service map with ports and access URLs
- Makefile command reference
- Monitoring: how to access Grafana, what dashboards show
- SSL: how certs work, how to use real certs in production
- Environment variable reference
- Network topology explanation
- Production considerations (what to change for real deployment)

**5.4 — Final checks**
- `make up-all` starts cleanly from scratch
- All health checks pass within 60 seconds
- API responds through nginx on HTTPS
- Prometheus scrapes app metrics
- Grafana dashboard shows live data
- `make clean && make up-all` (clean restart works)
- No secrets committed (only `.env.example`)
- `.gitignore` covers: `.env`, `nginx/certs/*.pem`, `*.key`

---

## Commit Plan

1. `chore: scaffold project with directory structure and Makefile`
2. `feat: add Go API with health check and graceful shutdown`
3. `feat: add PostgreSQL integration with CRUD handlers`
4. `feat: add Redis cache handlers with TTL`
5. `feat: add Prometheus metrics middleware and /metrics endpoint`
6. `feat: add multi-stage Dockerfile (< 20MB final image)`
7. `feat: add docker-compose with services, networks, and volumes`
8. `feat: add health checks for all services`
9. `feat: add dev and monitoring profiles`
10. `feat: add Nginx reverse proxy with gzip and rate limiting`
11. `feat: add SSL termination with self-signed cert generation`
12. `feat: add Prometheus scrape config`
13. `feat: add Grafana auto-provisioned dashboard`
14. `feat: add .env.example with documented variables`
15. `docs: add README with architecture diagram and usage guide`
