#!/bin/sh
set -eu

mkdir -p /data /data/caddy

: "${APP_ENV:=production}"
: "${DATABASE_DSN:=/data/caddypilot.db}"
: "${CADDY_ADMIN_API:=http://127.0.0.1:2019}"
: "${CADDY_BINARY:=/usr/bin/caddy}"
: "${CADDY_DATA_DIR:=/data/caddy}"
: "${CADDYPILOT_RUNTIME_DIR:=/data/runtime}"
: "${CADDYPILOT_BACKEND_ADDR:=127.0.0.1:25610}"
: "${CADDYPILOT_FRONTEND_DIR:=/app/frontend}"
: "${CADDYPILOT_MANAGE_ADDR:=:8080}"
: "${TZ:=Asia/Shanghai}"

export APP_ENV DATABASE_DSN CADDY_ADMIN_API CADDY_BINARY CADDY_DATA_DIR
export CADDYPILOT_RUNTIME_DIR CADDYPILOT_BACKEND_ADDR CADDYPILOT_FRONTEND_DIR
export CADDYPILOT_MANAGE_ADDR TZ

exec /app/backend/caddypilot
