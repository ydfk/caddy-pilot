# Contributing to CaddyPilot

Thank you for your interest in contributing! This document outlines the development workflow and conventions.

## Table of Contents

- [Development environment](#development-environment)
- [Project structure](#project-structure)
- [Making changes](#making-changes)
- [Code style](#code-style)
- [Testing](#testing)
- [Pull request process](#pull-request-process)
- [Commit conventions](#commit-conventions)

---

## Development environment

### Prerequisites

- **Go** 1.25+ (see `backend/go.mod`)
- **Node.js** 20+ and **pnpm** (see `frontend/package.json`)
- **C compiler** (required by SQLite via GORM — use `gcc` on Linux/macOS, or the MinGW toolchain on Windows)
- **Docker** with Compose v2 (optional, for containerized workflow)

### One-click start

On Windows, double-click `scripts\dev.cmd` or run:

```powershell
.\scripts\dev.cmd
```

This starts the Vite dev server, the Go backend, and a managed Caddy process automatically.

### Manual start

```bash
# Terminal 1: backend
cd backend
go run ./cmd

# Terminal 2: frontend
cd frontend
pnpm install
pnpm dev
```

The complete system is available at `http://localhost:8080`. Do **not** start Caddy manually — the Go backend manages its lifecycle.

---

## Project structure

```
caddy-pilot/
├── backend/           # Go API server (Fiber + Huma + GORM)
│   ├── cmd/           #   Entrypoint (main.go)
│   ├── internal/      #   Core logic
│   └── ...
├── frontend/          # React 19 SPA (Vite + Tailwind + shadcn/ui)
│   ├── src/           #   Components, pages, stores
│   └── ...
├── docs/              # Design and reference documentation
├── scripts/           # PowerShell development and CI helpers
├── data/              # Runtime data (git-ignored)
├── Dockerfile         # Production multi-stage build
└── docker-compose.yml # Default Compose file
```

---

## Making changes

1. **Fork** the repository and create your feature branch from `main`:
   ```bash
   git checkout -b feat/my-feature
   ```

2. **Make your changes**, keeping them focused and well-scoped.

3. **Run tests** before committing (see [Testing](#testing)).

4. **Keep commits atomic** — each commit should represent one logical change.

---

## Code style

### Backend (Go)

- Follow [Effective Go](https://go.dev/doc/effective_go) and standard `gofmt` formatting.
- Use `gofumpt` for stricter formatting if possible.
- Error handling: wrap errors with context using `fmt.Errorf("…: %w", err)`.
- Prefer explicit naming over short variable names for exported symbols.
- Run `go vet ./...` before committing.

### Frontend (TypeScript / React)

- Use the project's ESLint and Prettier configs.
- Follow the existing component patterns (shadcn/ui style).
- Use Zustand for global state, React Hook Form + Zod for forms.
- Run `pnpm lint` and `pnpm typecheck` before committing.

### Commit messages

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]
```

**Types:** `feat`, `fix`, `chore`, `docs`, `style`, `refactor`, `test`, `ci`

**Scope examples:** `backend`, `frontend`, `docker`, `docs`

---

## Testing

### Backend

```bash
cd backend
go test ./... -v
```

### Frontend

```bash
cd frontend
pnpm test:run
```

### Both (Windows)

```powershell
.\scripts\test.cmd
```

Before submitting a pull request, ensure:
- All tests pass.
- `go vet ./...` reports no issues.
- The frontend builds without errors.

---

## Pull request process

1. Ensure your branch is up to date with `main`.
2. Write a clear PR title and description explaining **what** and **why**.
3. Reference any related issue numbers.
4. A maintainer will review your changes — expect constructive feedback.
5. Once approved, a maintainer will merge your PR.

---

## Questions?

Open a [discussion](https://github.com/ydfk/caddy-pilot/discussions) or issue — we're happy to help!
