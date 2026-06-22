<p align="center">
  <img src="frontend/public/caddypilot-logo.png" width="128" alt="CaddyPilot logo">
  <h1 align="center">CaddyPilot</h1>
  <p align="center">
    A lightweight, self-contained web dashboard for managing the <a href="https://caddyserver.com">Caddy</a> bundled with CaddyPilot.
    <br />
    Manage proxy sites, preview configs, publish with one click, and roll back safely — all from a single Docker image.
  </p>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go" alt="Go version">
  <img src="https://img.shields.io/badge/React-19-61DAFB?style=flat&logo=react" alt="React version">
  <img src="https://img.shields.io/badge/Caddy-2.10-00B247?style=flat&logo=caddy" alt="Caddy">
  <img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License">
  <img src="https://img.shields.io/badge/Docker-Compose-2496ED?style=flat&logo=docker" alt="Docker">
</p>

---

## What is CaddyPilot?

CaddyPilot wraps your Caddy server with a clean web UI so you never edit JSON configs by hand. It bundles a React frontend, Go API, and Caddy into **one Docker image** — deploy, log in, add sites, and publish.

All configuration changes are versioned. If a publish fails, your previous working config stays live. If you need to go back, rollback is one click away.

### Feature highlights

- **Proxy site management** — create, edit, clone, soft-delete, enable/disable reverse proxy sites
- **Typed upstreams** — HTTP, HTTPS, h2c, and Unix Socket with type-specific settings
- **Access & certificates** — reusable Basic Auth password vault plus single-domain or wildcard certificates with Aliyun DNS-01
- **Config preview & publish** — see the exact Caddy JSON before it goes live; push to Caddy Admin API in one step
- **Version history & rollback** — every publish is recorded; diff, inspect, or rollback to any previous version
- **Dashboard** — at-a-glance status: site counts, Caddy health, last publish time
- **Caddy version management** — the backend downloads, starts, monitors, updates, and stops Caddy automatically
- **Self-protecting** — the management port (`:8080`) is never removed from generated configs, so the UI always stays reachable
- **Single system** — Caddy + Go API + React ship together; users never install or start Caddy separately

### What it does NOT do

MVP scope is intentionally narrow. CaddyPilot does **not** handle:

Multi-node clusters, role-based access, Caddyfile import/export, DNS providers other than Aliyun, Layer 4 routing, automatic container discovery, advanced log analytics, or full Caddy config visualization.

---

## Architecture

```
Browser -> :8080 Caddy -> /api/* -> 127.0.0.1:25610 Go API
                    \-> React static files /app/frontend

Go API -> SQLite /data/caddypilot.db
Go API -> Caddy Admin API 127.0.0.1:2019
Go API -> managed Caddy process lifecycle and binary updates
```

Every generated Caddy config automatically includes the management server block on `:8080`. The API validates this invariant before every publish and rollback, so you can't lock yourself out.

Detailed design: [docs/design.md](docs/design.md)

---

## Tech stack

| Layer | Technology |
|-------|-----------|
| Frontend | React 19, TypeScript, Vite, Tailwind CSS, shadcn/ui, Zustand, React Hook Form, Zod |
| Backend | Go, Fiber, Huma (OpenAPI), GORM, SQLite |
| Proxy | Caddy 2 |
| Runtime | Backend-managed Caddy process; Docker or native Windows |

---

## Quick start

```bash
# Clone the repo
git clone https://github.com/ydfk/caddy-pilot.git
cd caddy-pilot

# Set a JWT secret (required)
export JWT_SECRET="replace-with-a-long-random-string"

# Build and start
docker compose up -d --build

# Open the dashboard
# http://localhost:8080
```

On first launch, the login page shows an admin initialization form. Create your admin account and you're in.

Stop everything:

```bash
docker compose down
```

---

## Docker Compose configuration

```yaml
services:
  caddypilot:
    image: caddypilot:latest
    build:
      context: .
      dockerfile: Dockerfile
    container_name: caddypilot
    restart: unless-stopped
    ports:
      - "8080:8080"   # management UI
      - "80:80"       # proxy traffic
      - "443:443"     # proxy HTTPS
    environment:
      JWT_SECRET: ${JWT_SECRET}
      DATABASE_DSN: /data/caddypilot.db
      CADDY_ADMIN_API: http://127.0.0.1:2019
      CADDYPILOT_MANAGE_ADDR: :8080
    volumes:
      - ./data:/data
```

## Security

**Do not expose port 2019 to the host or the internet.** The Caddy Admin API listens on `127.0.0.1:2019` inside the container only. The default compose file does not map it. The UI never proxies it to the browser.

Change `JWT_SECRET` before deploying. The default value is a placeholder and produces a warning at startup.

---

## Data

All persistent data lives under `./data/`:

| Path | Content |
|------|---------|
| `/data/caddypilot.db` | Users, proxy sites, config versions |
| `/data/caddy/` | Caddy certificates and runtime state |

Back up both when migrating or upgrading. The `data/` directory is git-ignored.

---

## Development

**One-click start (Windows):**

Double-click `dev.cmd` or run in PowerShell:

```powershell
.\dev.cmd
```

This starts Vite and the Go backend natively (no Docker). The backend automatically downloads a private Caddy runtime when needed, starts it, and exposes the complete system at `http://localhost:8080`. No system-wide Caddy installation is required.

**Manual start:**

```bash
# Backend (terminal 1)
cd backend
go test ./...
go run ./cmd

# Frontend (terminal 2)
cd frontend
pnpm install
pnpm dev
```

Do not start Caddy manually. The backend owns its lifecycle in both Docker and native environments; if no bundled runtime exists, it downloads the configured version into `data/runtime/`.

Environment variables for development:

| Variable | Default | Purpose |
|----------|---------|---------|
| `CADDY_VERSION` | `2.10.0` | Managed Caddy target version |
| `CADDY_VERSION_CHECK_URL` | GitHub releases API | Version check endpoint |
| `CADDY_DOWNLOAD_URL` | Caddy custom build API | Managed build with the Aliyun DNS module; supports `{version}`, `{os}`, `{arch}` |
| `CADDY_CHECKSUM_URL` | empty | Optional SHA-512 manifest URL for a custom download source |
| `CADDYPILOT_SECRET_KEY` | falls back to `JWT_SECRET` | AES-GCM key used to encrypt DNS credentials; do not rotate without re-entering credentials |
| `VITE_PROXY_HOST` | `http://127.0.0.1:25610` | Vite dev proxy target |
| `JWT_SECRET` | — | JWT signing key (required in production) |

---

## Documentation

- [Design & security boundaries](docs/design.md)
- [API reference](docs/api.md)
- [Caddy JSON generation](docs/caddy-json.md)
- [Deployment guide](docs/deployment.md)

---

## Roadmap

- Merge `advanced_json` into generated configs via a controlled allowlist (currently saved but not applied)
- Add DNS providers beyond Aliyun
- Expand `enable_log` into per-site access log configuration
- Auto-restore the last successful business config on restart (currently boots with a backend-generated protected management config)

---

## License

MIT — see [backend/LICENSE](backend/LICENSE)
