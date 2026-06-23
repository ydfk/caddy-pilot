#!/bin/sh
set -eu

mkdir -p /data /data/caddy

secrets_file=/data/.caddypilot-secrets

generate_secret() {
    od -An -N32 -tx1 /dev/urandom | tr -d ' \n'
}

read_secret() {
    key="$1"
    [ -f "$secrets_file" ] || return 0
    sed -n "s/^${key}=//p" "$secrets_file" | head -n 1
}

persist_secrets() {
    temporary_file="${secrets_file}.tmp.$$"
    umask 077
    {
        printf 'JWT_SECRET=%s\n' "$JWT_SECRET"
        printf 'CADDYPILOT_SECRET_KEY=%s\n' "$CADDYPILOT_SECRET_KEY"
    } > "$temporary_file"
    chmod 600 "$temporary_file"
    mv "$temporary_file" "$secrets_file"
}

if [ -z "${JWT_SECRET:-}" ]; then
    JWT_SECRET="$(read_secret JWT_SECRET)"
fi
if [ -z "${CADDYPILOT_SECRET_KEY:-}" ]; then
    CADDYPILOT_SECRET_KEY="$(read_secret CADDYPILOT_SECRET_KEY)"
fi

: "${JWT_SECRET:=$(generate_secret)}"
: "${CADDYPILOT_SECRET_KEY:=$(generate_secret)}"

if [ ! -f "$secrets_file" ]; then
    persist_secrets
fi
chmod 600 "$secrets_file"

: "${APP_ENV:=production}"
: "${DATABASE_DSN:=/data/caddypilot.db}"
: "${CADDY_ADMIN_API:=http://127.0.0.1:2019}"
: "${CADDY_BINARY:=/usr/bin/caddy}"
: "${CADDY_DATA_DIR:=/data/caddy}"
: "${CADDYPILOT_RUNTIME_DIR:=/data/runtime}"
: "${CADDYPILOT_BACKEND_ADDR:=127.0.0.1:25610}"
: "${CADDYPILOT_FRONTEND_DIR:=/app/frontend}"
: "${CADDYPILOT_MANAGE_ADDR:=:8080}"
: "${CADDYPILOT_HTTPS_PORT:=443}"
: "${TZ:=Asia/Shanghai}"

export APP_ENV DATABASE_DSN CADDY_ADMIN_API CADDY_BINARY CADDY_DATA_DIR
export CADDYPILOT_RUNTIME_DIR CADDYPILOT_BACKEND_ADDR CADDYPILOT_FRONTEND_DIR
export CADDYPILOT_MANAGE_ADDR TZ
export CADDYPILOT_HTTPS_PORT
export JWT_SECRET CADDYPILOT_SECRET_KEY

exec /app/backend/caddypilot
