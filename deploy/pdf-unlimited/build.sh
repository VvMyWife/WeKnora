#!/usr/bin/env bash
set -euo pipefail
cd /home/ubuntu/RAG_Weknora
mkdir -p runtime
exec 9>runtime/pdf-unlimited-build.lock
flock -n 9 || exit 0
trap 'echo $? > runtime/full-build.status' EXIT
docker build -t weknora-builder:pdf-unlimited -f deploy/pdf-unlimited/Dockerfile.builder deploy/pdf-unlimited
docker run --rm -v "$PWD:/src" -v "$PWD/.go-modcache:/go/pkg/mod" -v weknora-go-build-cache:/root/.cache/go-build -w /src weknora-builder:pdf-unlimited sh -c 'go test github.com/hibiken/asynq github.com/hibiken/asynq/internal/context -run TestUnlimited -count=1 && go test ./internal/infrastructure/docparser ./internal/application/repository ./internal/application/service ./internal/handler ./internal/router ./internal/config && go build -buildvcs=false -o runtime/WeKnora.full ./cmd/server'
mkdir -p runtime/yanyiwu
cp -a .go-modcache/github.com/yanyiwu/. runtime/yanyiwu/
docker build -f deploy/pdf-unlimited/Dockerfile.runtime -t weknora-app:pdf-unlimited-20261010 runtime
docker run --rm --network none -v "$PWD:/src" -w /src/frontend-custom node:22-bookworm sh -c 'npm test -- src/views/knowledge/processingOverview.test.ts && npm run type-check && npm run build'
docker build -f frontend-custom/Dockerfile.local -t archive-intelligence-ui:pdf-unlimited-20261010 frontend-custom
