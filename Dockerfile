# syntax=docker/dockerfile:1.7

ARG CADDY_VERSION=2.10.0
ARG APP_VERSION=dev

FROM node:22-alpine AS frontend-build
WORKDIR /src/frontend
RUN corepack enable && corepack prepare pnpm@10.6.2 --activate
COPY frontend/package.json frontend/pnpm-lock.yaml frontend/pnpm-workspace.yaml ./
RUN --mount=type=cache,target=/root/.local/share/pnpm/store \
    pnpm install --frozen-lockfile
COPY frontend/ ./
RUN pnpm build

FROM golang:1.25-alpine AS backend-build
ARG APP_VERSION
RUN apk add --no-cache build-base
WORKDIR /src/backend
COPY backend/go.mod backend/go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY backend/ ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=1 GOOS=linux go build -trimpath -ldflags="-s -w -X go-fiber-starter/pkg/version.Current=${APP_VERSION}" -o /out/caddypilot ./cmd

FROM caddy:${CADDY_VERSION}-builder-alpine AS caddy-build
ARG CADDY_VERSION
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    xcaddy build v${CADDY_VERSION} --with github.com/caddy-dns/alidns

FROM caddy:${CADDY_VERSION}-alpine
RUN apk add --no-cache ca-certificates tzdata \
    && mkdir -p /app/backend/config /app/frontend /data/runtime /data/caddy \
    && chmod 0755 /data

WORKDIR /app/backend
COPY --from=frontend-build /src/frontend/dist/ /app/frontend/
COPY --from=backend-build /out/caddypilot /app/backend/caddypilot
COPY --from=caddy-build /usr/bin/caddy /usr/bin/caddy
COPY backend/config/config.yaml /app/backend/config/config.yaml
COPY docker/entrypoint.sh /usr/local/bin/caddypilot-entrypoint
RUN chmod 0755 /app/backend/caddypilot /usr/local/bin/caddypilot-entrypoint

ENV APP_ENV=production \
    CADDY_ADMIN_API=http://127.0.0.1:2019 \
    CADDY_BINARY=/usr/bin/caddy \
    CADDY_DATA_DIR=/data/caddy \
    CADDYPILOT_RUNTIME_DIR=/data/runtime \
    CADDYPILOT_BACKEND_ADDR=127.0.0.1:25610 \
    CADDYPILOT_FRONTEND_DIR=/app/frontend \
    CADDYPILOT_MANAGE_ADDR=:8080 \
    DATABASE_DSN=/data/caddypilot.db \
    PYTHONWARNINGS=ignore \
    TZ=Asia/Shanghai

EXPOSE 80 443 8080
VOLUME ["/data"]
ENTRYPOINT ["/usr/local/bin/caddypilot-entrypoint"]
