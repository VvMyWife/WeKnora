#!/bin/bash
set -euo pipefail
cd "$(dirname "$0")/.."
test ! -f "$(git rev-parse --git-path MERGE_HEAD)"
source_revision=$(git rev-parse HEAD)
image=${LJC_BACKEND_IMAGE:-weknora-app:ljc-upstream-eecfe4b3}
printf 'running\n' > runtime/upstream/backend-image.status
trap 'printf "%s\n" "$?" > runtime/upstream/backend-image.status' EXIT
docker run --rm --label ljc.operation=upstream-20261020 -e COMMIT_ID="$source_revision" -e VERSION=0.8.3 -e BUILD_TIME="$(date -u +%FT%TZ)" -e GOFLAGS=-buildvcs=false -v "$PWD:/src" -v /home/ubuntu/RAG_Weknora/.go-modcache:/go/pkg/mod -v weknora-go-build-cache:/root/.cache/go-build -w /src weknora-builder:upstream-20261020 sh -ec '
 export GO_VERSION=$(go version | cut -d" " -f3)
 make build-prod GO_BUILD_TAGS=anydoc BINARY_NAME=runtime/upstream/WeKnora
 go build -o runtime/upstream/ljc-migrate ./cmd/ljc-migrate
 '
docker build -f deploy/upstream-integration/Dockerfile.app-runtime --build-arg SOURCE_REVISION="$source_revision" -t "$image" .
docker run --rm --entrypoint /app/ljc-migrate -e LJC_SQLITE_PATH=/tmp/ljc-smoke.sqlite "$image" -mode status
docker run --rm --entrypoint /opt/weknora/browserskill/bsk "$image" --version
