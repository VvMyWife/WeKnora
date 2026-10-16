#!/usr/bin/env bash
# Run from the authoritative repository. Never builds/replaces original frontend.
set -euo pipefail
cd /home/ubuntu/RAG_Weknora
image=weknora-app:external-chunks-20261016
case "${1:-}" in
  build)
    docker run --rm -v "$PWD:/src" -v "$PWD/.go-modcache:/go/pkg/mod" \
      -v weknora-go-build-cache:/root/.cache/go-build -w /src weknora-builder:pdf-unlimited \
      sh -c 'go build -buildvcs=false -o runtime/WeKnora.external-chunks-20261016 ./cmd/server'
    mkdir -p runtime/external-migrations/versioned runtime/external-migrations/sqlite
    cp migrations/versioned/000093* runtime/external-migrations/versioned/
    cp migrations/sqlite/000014* runtime/external-migrations/sqlite/
    docker build -f deploy/external-chunks/Dockerfile.runtime -t "$image" runtime
    ;;
  deploy)
    docker image inspect "$image" >/dev/null
    ENRICHMENT_APP_IMAGE="$image" docker compose -f docker-compose.yml \
      -f deploy/pdf-unlimited/compose.yaml -f deploy/enrichment-retry/compose.yaml up -d --no-deps app
    ;;
  rollback)
    ENRICHMENT_APP_IMAGE=weknora-app:wiki-output-12288-20261016 docker compose -f docker-compose.yml \
      -f deploy/pdf-unlimited/compose.yaml -f deploy/enrichment-retry/compose.yaml up -d --no-deps app
    ;;
  *) echo 'Usage: bash deploy/external-chunks/release.sh build|deploy|rollback' >&2; exit 2 ;;
esac
