FROM node:22-alpine AS frontend-build
WORKDIR /src/frontend
RUN corepack enable && corepack prepare pnpm@10.6.2 --activate
COPY frontend/package.json frontend/pnpm-lock.yaml frontend/pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile
COPY frontend/ ./
RUN pnpm build

FROM golang:1.25-alpine AS backend-build
RUN apk add --no-cache build-base
WORKDIR /src/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=1 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/caddypilot ./cmd

FROM caddy:2.10-alpine
RUN apk add --no-cache ca-certificates supervisor tzdata \
    && mkdir -p /app/backend/config /app/frontend /data /etc/supervisor/conf.d \
    && chmod 0755 /data

WORKDIR /app/backend
COPY --from=frontend-build /src/frontend/dist/ /app/frontend/
COPY --from=backend-build /out/caddypilot /app/backend/caddypilot
COPY backend/config/config.yaml /app/backend/config/config.yaml
COPY docker/Caddyfile /etc/caddy/Caddyfile
COPY docker/supervisord.conf /etc/supervisor/conf.d/supervisord.conf
COPY docker/entrypoint.sh /usr/local/bin/caddypilot-entrypoint
COPY docker/supervisor_shutdown.py /usr/local/bin/supervisor-shutdown
RUN chmod 0755 /app/backend/caddypilot /usr/local/bin/caddypilot-entrypoint /usr/local/bin/supervisor-shutdown

ENV APP_ENV=production \
    CADDY_ADMIN_API=http://127.0.0.1:2019 \
    CADDYPILOT_BACKEND_ADDR=127.0.0.1:25610 \
    CADDYPILOT_FRONTEND_DIR=/app/frontend \
    CADDYPILOT_MANAGE_ADDR=:8080 \
    DATABASE_DSN=/data/caddypilot.db \
    PYTHONWARNINGS=ignore \
    TZ=Asia/Shanghai

EXPOSE 80 443 8080
VOLUME ["/data"]
ENTRYPOINT ["/usr/local/bin/caddypilot-entrypoint"]
