# Go Fiber Starter

A Go API starter built with Fiber v3, Huma, and GORM. It uses OpenAPI 3.1 code-first: write typed Go handlers first, then run the application to get live API documentation and JSON/YAML specifications. There is no spec-first code generation step.

[简体中文](README_zh.md)

## Features

- Fiber v3 HTTP server
- Huma v2 code-first OpenAPI 3.1
- Runtime API documentation at `/docs`
- JWT Bearer authentication
- Official GORM SQLite, PostgreSQL, and MySQL drivers
- Layered YAML configuration
- Zap logging and automatic database migration
- Win11 native scripts and Docker-based hot reload
- Multi-stage Alpine Linux production image

## OpenAPI

After startup:

- Documentation UI: `http://localhost:25610/docs`
- OpenAPI 3.1 JSON: `http://localhost:25610/openapi.json`
- OpenAPI 3.1 YAML: `http://localhost:25610/openapi.yaml`
- OpenAPI 3.0.3 compatibility: `/openapi-3.0.json` and `/openapi-3.0.yaml`

The project does not contain generated Swagger files and does not use `swag init`. Define typed inputs and outputs, register a `huma.Operation`, and Huma updates the specification at runtime.

## Development

### Native Win11

Install Go, Air, and MinGW-w64. Ensure `gcc.exe` is on `PATH`, then enable CGO:

```powershell
go env -w CGO_ENABLED=1
go install github.com/air-verse/air@v1.65.3
..\scripts\dev-server.cmd
```

Project-level scripts are centralized at the repository root:

```bat
..\scripts\build.cmd
..\scripts\dev.cmd
..\scripts\test.cmd
```

### Production

```bash
docker compose up -d --build
docker compose logs -f
```

The production image builds SQLite with CGO on Alpine Linux and runs as a non-root user.

## API

| Method | Path | Description | Authentication |
| --- | --- | --- | --- |
| `POST` | `/api/auth/register` | Register a user | No |
| `POST` | `/api/auth/login` | Get a JWT | No |
| `GET` | `/api/auth/profile` | Get the current user | Bearer JWT |

Huma validates request bodies and returns real HTTP statuses such as `201`, `401`, `409`, and `422`.

## Configuration

Configuration files are loaded in this order, with later files overriding earlier values:

1. `config/config.yaml`
2. `config/config.<env>.yaml`
3. `config/config.local.yaml`
4. `config/config.<env>.local.yaml`

See `config/config.local.yaml.example` for local private settings. Database drivers include `sqlite`, `postgres`, `postgresql`, and `mysql`.

## Project Layout

```text
go-fiber-starter/
├── cmd/                         # Application and Huma setup
├── config/                      # YAML configuration
├── internal/api/auth/           # Typed handlers, operations, and JWT middleware
├── internal/model/              # Database models
├── internal/service/            # Business services
├── pkg/                         # Config, database, logging, and utilities
├── Dockerfile                   # Production image
├── Dockerfile.dev               # CGO and Air development image
├── docker-compose.yml
└── docker-compose.dev.yml
```

## Testing

Native testing requires a working C compiler:

```bash
go test ./...
go vet ./...
```

The test suite covers authentication, JWT validation, database configuration, and the generated OpenAPI 3.1 specification.

## Dependency Decisions

- `github.com/danielgtaylor/huma/v2` directly supports Fiber v3 and generates OpenAPI 3.1 at runtime.
- `gorm.io/driver/sqlite` is the officially maintained GORM driver and requires CGO.
- Swaggo, generated Swagger 2 files, Fiber contrib JWT, and `github.com/glebarez/sqlite` have been removed.

`GOPROXY` and `APK_MIRROR` are overridable Docker build arguments.

## License

[MIT](LICENSE)
