#!/usr/bin/env sh
set -eu

cd "$(dirname "$0")"

if [ ! -f ".env" ]; then
    echo "Missing .env" >&2
    exit 1
fi

APP_ENV_VALUE="$(grep -E '^APP_ENV=' .env | tail -n 1 | cut -d '=' -f 2-)"
APP_ENV_VALUE="$(printf '%s' "$APP_ENV_VALUE" | sed "s/[[:space:]]//g; s/^\"//; s/\"$//; s/^'//; s/'$//")"

if [ "$APP_ENV_VALUE" = "production" ]; then
    COMPOSE_FILES="-f docker-compose.prod.yml"
else
    COMPOSE_FILES="-f docker-compose.yml"
fi

docker compose $COMPOSE_FILES up -d --force-recreate "$@"
docker compose $COMPOSE_FILES ps
