# Build compose-stack — Production Docker Compose Stack

You are building a **portfolio project** for a Senior AI Engineer's public GitHub. It must be impressive, clean, and production-grade. Read these docs before writing any code:

1. **`D01-docker-compose-stack.md`** — Complete project spec: architecture, service topology, Nginx config, Prometheus/Grafana setup, commit plan. This is your primary blueprint. Follow it phase by phase.
2. **`github-portfolio.md`** — Portfolio goals and Definition of Done (Level 1 + Level 2). Understand the quality bar.
3. **`github-portfolio-checklist.md`** — Pre-publish checklist. Every item must pass before you're done.

---

## Instructions

### Read first, build second
Read all three docs completely before writing any code. Understand the full service topology (app → nginx → client, app → postgres/redis, prometheus → app, grafana → prometheus), the Docker Compose v2 features (profiles, health checks, depends_on conditions), and the networking architecture.

### Follow the phases in order
The project spec has 5 phases. Do them in order:
1. **Go Application** — HTTP server with health check, PostgreSQL CRUD, Redis cache, Prometheus metrics middleware
2. **Docker Setup** — multi-stage Dockerfile (< 20MB), docker-compose.yml with services/networks/volumes, health checks, profiles
3. **Nginx Reverse Proxy** — reverse proxy config, SSL termination with self-signed certs, security headers, rate limiting
4. **Monitoring Stack** — Prometheus scrape config, Grafana auto-provisioned data source + dashboard
5. **Documentation & Polish** — .env.example, Makefile, README with architecture diagram, final verification

### Commit frequently
Follow the commit plan in the spec. Use **conventional commits** (`feat:`, `test:`, `refactor:`, `docs:`, `ci:`, `chore:`). Each commit should be a logical unit.

### Quality non-negotiables
- **Multi-stage Dockerfile.** Build stage with Go compiler, final stage with just the binary on Alpine. Final image must be under 20MB. Non-root user in production image.
- **Health checks on ALL services.** Every service in docker-compose.yml has a `healthcheck` block. `depends_on` uses `condition: service_healthy`. No `sleep` hacks.
- **Named networks with isolation.** Frontend network (nginx + app), backend network (app + postgres + redis), monitoring network (prometheus + grafana + app). Services only connect to networks they need.
- **Profiles for flexibility.** Core stack runs by default. Monitoring (prometheus + grafana) is a profile. Dev (exposed DB/Redis ports) is a profile.
- **Nginx with security headers.** HSTS, X-Content-Type-Options, X-Frame-Options, Content-Security-Policy. Not just a basic proxy_pass.
- **Grafana auto-provisions.** On first startup, Grafana has the Prometheus data source and the app dashboard ready. No manual setup required.
- **Makefile covers all operations.** `make up`, `make down`, `make build`, `make logs`, `make ps`, `make clean`, `make certs`. Developer doesn't need to memorize docker compose flags.
- **No secrets committed.** `.env.example` with placeholder values. `.gitignore` covers `.env`, cert files, data volumes.
- **Go app is stdlib HTTP.** No Gin, no Chi. Use `net/http` with Go 1.22+ routing. Prometheus client library is the only non-stdlib dependency.

### What NOT to do
- Don't use docker-compose v1 syntax (`version: "3.x"`). Use Docker Compose v2 (no version key needed).
- Don't use `sleep` or `wait-for-it.sh` scripts. Use health checks + `depends_on` conditions.
- Don't expose database or Redis ports to the host by default. Only in the `dev` profile.
- Don't bake certificates into Docker images. Mount them as volumes.
- Don't use a Go HTTP framework. stdlib `net/http` only (Prometheus client library is allowed).
- Don't leave `// TODO` or `// FIXME` comments anywhere.

---

## GitHub Username

The GitHub username is **devaloi**. For Go module paths, use `github.com/devaloi/compose-stack`. For any GitHub URLs in the README, use `github.com/devaloi/compose-stack`.

## Start

Read the three docs. Then begin Phase 1 from `D01-docker-compose-stack.md`.
