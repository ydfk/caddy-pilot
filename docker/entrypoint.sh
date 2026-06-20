#!/bin/sh
set -eu

mkdir -p /data /data/caddy

: "${APP_ENV:=production}"
: "${DATABASE_DSN:=/data/caddypilot.db}"
: "${CADDY_ADMIN_API:=http://127.0.0.1:2019}"
: "${CADDYPILOT_BACKEND_ADDR:=127.0.0.1:25610}"
: "${CADDYPILOT_FRONTEND_DIR:=/app/frontend}"
: "${CADDYPILOT_MANAGE_ADDR:=:8080}"
: "${TZ:=Asia/Shanghai}"

export APP_ENV DATABASE_DSN CADDY_ADMIN_API CADDYPILOT_BACKEND_ADDR
export CADDYPILOT_FRONTEND_DIR CADDYPILOT_MANAGE_ADDR TZ

exec /usr/bin/supervisord -c /etc/supervisor/conf.d/supervisord.conf
