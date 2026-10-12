#!/usr/bin/env bash
set -euo pipefail
cd /home/ubuntu/RAG_Weknora
mkdir -p runtime
exec 9>runtime/pdf-unlimited-deploy.lock
flock -n 9 || exit 0
trap 'echo $? > runtime/deploy-unlimited.status' EXIT
# Immutable rollback image tags are created only once.
if ! docker image inspect weknora-app:before-pdf-unlimited >/dev/null 2>&1; then
 docker tag "$(docker inspect -f '{{.Image}}' WeKnora-app)" weknora-app:before-pdf-unlimited
fi
if ! docker image inspect archive-intelligence-ui:before-pdf-unlimited >/dev/null 2>&1; then
 docker tag "$(docker inspect -f '{{.Image}}' WeKnora-frontend)" archive-intelligence-ui:before-pdf-unlimited
fi
docker compose -f docker-compose.yml -f deploy/pdf-unlimited/compose.yaml --env-file .env up -d --no-deps app
ready=0
for n in $(seq 1 60); do
 if curl -fsS http://127.0.0.1:18081/health >/dev/null; then ready=1;break;fi
 sleep 2
done
[ "$ready" = 1 ]
docker tag archive-intelligence-ui:pdf-unlimited-20261010 archive-intelligence-ui:0.1.0
bash deploy/frontend/frontendctl switch custom
curl -fsS http://127.0.0.1:9960/full-page-status
curl -fsS http://127.0.0.1:18081/health
printf '\nDeployment healthy\n'
