#!/bin/bash
set -euo pipefail
cd "$(dirname "$0")/.."
export LJC_ACCEPTANCE_ENV=${LJC_ACCEPTANCE_ENV:-/home/ubuntu/weknora-fork-migration/20261020-upstream/acceptance.env}
python3 scripts/ljc_prepare_acceptance.py --private-dir "$(dirname "$LJC_ACCEPTANCE_ENV")"
docker compose --env-file "$LJC_ACCEPTANCE_ENV" -f deploy/upstream-integration/compose.acceptance.yml config --quiet
docker compose --env-file "$LJC_ACCEPTANCE_ENV" -f deploy/upstream-integration/compose.acceptance.yml up -d postgres redis neo4j docreader
# Explicit schema-only entry point; never dispatches application tasks.
docker run --rm --network ljc-upstream-validate --env-file "$LJC_ACCEPTANCE_ENV" --entrypoint /app/ljc-migrate weknora-app:ljc-upstream-eecfe4b3 -mode up
docker compose --env-file "$LJC_ACCEPTANCE_ENV" -f deploy/upstream-integration/compose.acceptance.yml up -d app official-ui

# nginx resolves the backend hostname at config load; refresh candidate routes
# after a recreated backend receives a different IP. No production UI is touched.
docker exec ljc-upstream-acceptance-official-ui nginx -s reload
if [ "$(docker inspect -f '{{.State.Running}}' ljc-upstream-acceptance-custom-ui 2>/dev/null || true)" = true ]; then
 docker exec ljc-upstream-acceptance-custom-ui nginx -s reload
fi
