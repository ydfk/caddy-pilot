<p align="center">
  <img src="frontend/public/caddypilot-logo.png" width="128" alt="CaddyPilot logo">
  <h1 align="center">CaddyPilot</h1>
  <p align="center">
    A lightweight, self-contained web dashboard for managing the <a href="https://caddyserver.com">Caddy</a> web server.
    <br />
    Manage proxy sites, preview configs, publish with one click, and roll back safely — all from a single Docker image.
  </p>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go" alt="Go version">
  <img src="https://img.shields.io/badge/React-19-61DAFB?style=flat&logo=react" alt="React version">
  <img src="https://img.shields.io/badge/Caddy-2.x-00B247?style=flat&logo=caddy" alt="Caddy">
  <img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License">
  <img src="https://img.shields.io/badge/Docker-Compose-2496ED?style=flat&logo=docker" alt="Docker">
</p>

---

## Overview

CaddyPilot wraps your Caddy server with a clean web UI so you never edit JSON configs by hand. It bundles a React frontend, Go API, and Caddy into **one Docker image** — deploy, log in, add sites, and publish.

All configuration changes are versioned. If a publish fails, your previous working config stays live. If you need to go back, rollback is one click away.

### Features

- **Proxy site management** — create, edit, clone, soft-delete, enable/disable reverse proxy sites
- **Typed upstreams** — HTTP, HTTPS, h2c, and Unix Socket with type-specific settings
- **Access & certificates** — reusable Basic Auth password vault plus single-domain or wildcard certificates with Aliyun DNS-01
- **Password or Passkey login** — keep password recovery while registering multiple encrypted Passkey credentials
- **Config preview & publish** — publish protected JSON while previewing and exporting a validated, read-only Caddyfile view
- **Nginx migration** — import common server, upstream, proxy_pass, TLS listener, and HTTPS redirect patterns as reviewable disabled sites
- **Version history & rollback** — every publish is recorded; diff, inspect, or roll back to any previous version
- **Dashboard** — at-a-glance status: site counts, Caddy health, last publish time
- **Caddy workbench** — runtime health, validate/publish flow, config history, and Caddy updates in one place
- **System identity** — global sidebar shows the CaddyPilot version embedded from the release tag
- **Runtime observability** — inspect certificate issuance state and follow system, Caddy, or credential-safe DNS provider audit logs online
- **Self-protecting** — the management port (`:8080`) is never removed from generated configs, so the UI always stays reachable
- **Single system** — Caddy + Go API + React ship together; users never install or start Caddy separately

### What it does NOT do

MVP scope is intentionally narrow. CaddyPilot does **not** handle multi-node clusters, role-based access, editable Caddyfile import, DNS providers other than Aliyun, Layer 4 routing, automatic container discovery, advanced log analytics, or full Caddy config visualization.

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

### Prerequisites

- Docker Engine with Docker Compose v2

### Start with Docker

```bash
# Clone the repo
git clone https://github.com/ydfk/caddy-pilot.git
cd caddy-pilot

# Build and start
docker compose up -d --build

# Open the dashboard
# http://localhost:8080
```

On first launch, the login page shows an admin initialization form. Create your admin account and you're in.

```bash
# Stop everything
docker compose down
```

### Production deployment

For a production host pulling the published Docker Hub image:

```bash
docker compose -f docker-compose.prod.yml up -d
```

> **Tip:** Pin a specific version in production. Set `CADDYPILOT_VERSION` in your environment to a stable tag (e.g., `1.2.3`) and update only after testing. The default `latest` tag tracks the most recent stable release.

---

## Configuration

### Port mapping

| Environment variable | Default host port | Container port | Purpose |
|---------------------|-------------------|----------------|---------|
| `CADDYPILOT_HTTP_PORT` | 80 | 80 | HTTP proxy sites |
| `CADDYPILOT_HTTPS_PORT` | 443 | 443 | HTTPS proxy sites and global redirect target |
| — | 8080 | 8080 | Management UI |

To use non-standard host ports, create a `.env` file in the project root:

```dotenv
CADDYPILOT_HTTP_PORT=18080
CADDYPILOT_HTTPS_PORT=18443
```

> **Note:** Avoid port `10080` — Chromium-based browsers consider it unsafe and return `ERR_UNSAFE_PORT`.

The proxy-site list uses these two values when opening a domain. Standard ports are omitted from the URL; non-standard ports are included.

### Passkey and container DNS

Passkey keeps password login enabled and stores credential material encrypted in the database. The default `localhost` configuration supports local development. For a deployed management domain, configure the exact HTTPS origin in `.env`:

```dotenv
CADDYPILOT_PASSKEY_RP_ID=pilot.example.com
CADDYPILOT_PASSKEY_RP_NAME=CaddyPilot
CADDYPILOT_PASSKEY_ORIGINS=https://pilot.example.com
```

`CADDYPILOT_PASSKEY_ORIGINS` accepts a comma-separated list. The RP ID must be the management domain without scheme or port. Browsers require HTTPS for Passkey outside `localhost`.

Compose uses `223.5.5.5` by default to avoid Docker embedded-DNS failures while checking GitHub releases. Override it when your network requires another resolver:

```dotenv
CADDYPILOT_DNS_SERVER=1.1.1.1
```

### Data persistence

All persistent data lives under `./data/`:

| Path | Content |
|------|---------|
| `/data/caddypilot.db` | Users, encrypted Passkeys, proxy sites, config versions |
| `/data/caddy/` | Caddy certificates and runtime state |
| `/data/runtime/caddy/` | Managed Caddy binaries, selected version, and active JSON |
| `/data/logs/` | Rotated CaddyPilot and Caddy process logs |
| `/data/.caddypilot-secrets` | Automatically managed JWT and credential-encryption keys |

Back up the entire directory when migrating or upgrading. Losing `.caddypilot-secrets` makes existing encrypted DNS credentials unreadable. The `data/` directory is git-ignored.

### Security

**Do not expose port 2019 to the host or the internet.** The Caddy Admin API listens on `127.0.0.1:2019` inside the container only. The default Compose file does not map it. The UI never proxies it to the browser.

On first start, the container generates independent random JWT and credential-encryption keys. They are stored with mode `0600` under `/data/.caddypilot-secrets` and reused after restart or image upgrade.

---

## Development

### One-click start (Windows)

Double-click `scripts\dev.cmd` or run in PowerShell:

```powershell
.\scripts\dev.cmd
```

This starts Vite and the Go backend natively (no Docker). The backend automatically downloads a private Caddy runtime when needed, starts it, and exposes the complete system at `http://localhost:8080`. No system-wide Caddy installation is required.

### Manual start

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

### Environment variables

| Variable | Default | Purpose |
|----------|---------|---------|
| `CADDY_VERSION` | `2.11.4` | Native development bootstrap version |
| `VITE_PROXY_HOST` | `http://127.0.0.1:25610` | Vite dev proxy target |

### Available scripts

| Command | Purpose |
| --- | --- |
| `scripts\dev.cmd` | Start frontend, backend, and managed Caddy |
| `scripts\dev-web.cmd` | Start only the Vite frontend |
| `scripts\dev-server.cmd` | Start backend and managed Caddy with Air |
| `scripts\build.cmd` | Build frontend and backend |
| `scripts\test.cmd` | Run frontend and backend tests |

---

## Docker image releases

Pushing a semantic version tag such as `v1.2.3` builds and publishes the Linux amd64 image to Docker Hub, then creates a GitHub Release containing only the production Compose file. Configure repository secrets `DOCKERHUB_USERNAME` and `DOCKERHUB_TOKEN` first.

- Stable tags (e.g., `v1.2.3`) update `latest` and create a formal Release.
- Pre-release tags (e.g., `v1.2.3-rc.1`) create a Prerelease and do **not** overwrite `latest`.

The system sidebar displays the CaddyPilot version from the same tag. For example, `v1.2.3` publishes Docker image `ydfk/caddy-pilot:1.2.3` and the dashboard shows version `1.2.3`.

---

## Documentation

- [Design & security boundaries](docs/design.md)
- [API reference](docs/api.md)
- [Caddy JSON generation](docs/caddy-json.md)
- [Deployment guide](docs/deployment.md) (Chinese)

---

## Troubleshooting

| Symptom | Likely cause & action |
|---------|----------------------|
| Management page unreachable | Check container health and Caddy logs with `docker compose logs --tail 100`. |
| API returns 401 | Session expired — re-login. The browser cleans stale tokens automatically. |
| Publish returns 502 | Check the Caddy status page and the failed version's `error_message`. |
| Site unreachable after publish | Verify DNS resolution, firewall rules for ports 80/443, upstream address, and the active Caddy JSON. |
| Container restarts repeatedly | A managed Caddy crash or backend panic triggers a full restart. Inspect logs just before the restart for the root cause. |
| Caddy update fails with `127.0.0.11:53` | Docker DNS cannot resolve the update host. Set `CADDYPILOT_DNS_SERVER` and recreate the container. |
| Caddy update download fails | Downloads support resume and retry up to three times. Check `data/runtime/caddy/update-task.json` for details. |

### After restart

Since Caddy config persistence is intentionally disabled, the system boots with a backend-generated protected management config that keeps the UI reachable. Business proxy sites need to be re-published from the Caddy status page. Automatic restoration of the last successful config is on the roadmap.

---

## Roadmap

- Merge `advanced_json` into generated configs via a controlled allowlist (currently saved but not applied)
- Add DNS providers beyond Aliyun
- Auto-restore the last successful business config on restart

---

## Contributing

Contributions are welcome! Please read our [contributing guidelines](CONTRIBUTING.md) to get started.

---

## License

MIT — see [backend/LICENSE](backend/LICENSE)
