#!/usr/bin/env bash
# Builds and restarts the production stack. Runs on the VM, from the repo
# checkout, after the deploy workflow has copied the env files into deploy/.
set -euo pipefail

cd "$(dirname "$0")"

for f in .env backend.env runtime.env; do
  if [ ! -f "$f" ]; then
    echo "deploy/$f is missing. Run the Deploy workflow, which writes it from GitHub secrets." >&2
    exit 1
  fi
done
chmod 600 .env backend.env runtime.env

echo "==> Building images"
docker compose build --pull

echo "==> Starting Postgres and Redis"
docker compose up -d --wait postgres redis

echo "==> Applying database migrations"
docker compose run --rm --no-deps app sh -c "cd /app/packages/db && bunx prisma migrate deploy"

echo "==> Starting services"
docker compose up -d --remove-orphans

echo "==> Waiting for the backend and runtime"
for i in $(seq 1 30); do
  if docker compose exec -T app bun -e "fetch('http://localhost:4000/').then(r => process.exit(r.ok ? 0 : 1)).catch(() => process.exit(1))" >/dev/null 2>&1 \
    && docker compose exec -T agent-runtime python -c "import urllib.request; urllib.request.urlopen('http://localhost:8000/')" >/dev/null 2>&1; then
    echo "==> Healthy"
    docker image prune -f >/dev/null
    exit 0
  fi
  sleep 2
done

echo "Services did not become healthy. Recent logs:" >&2
docker compose logs --tail 50 app agent-runtime >&2
exit 1
