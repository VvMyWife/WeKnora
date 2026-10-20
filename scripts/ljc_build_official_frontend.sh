#!/bin/bash
set -euo pipefail
cd "$(dirname "$0")/.."
mkdir -p runtime/upstream
printf 'running\n' > runtime/upstream/official-frontend.status
trap 'printf "%s\n" "$?" > runtime/upstream/official-frontend.status' EXIT
if ! docker image inspect weknora-frontend-builder:upstream-20261020 >/dev/null 2>&1; then
 docker build -f deploy/upstream-integration/Dockerfile.frontend-builder -t weknora-frontend-builder:upstream-20261020 .
fi
docker volume create --label ljc.operation=upstream-20261020 ljc-upstream-official-node-modules >/dev/null
docker run --rm --name ljc-upstream-official-ui-build -v "$PWD:/src" -v ljc-upstream-official-node-modules:/src/frontend/node_modules -v /home/ubuntu/.npm:/root/.npm -w /src/frontend -e VITE_IS_DOCKER=true -e VITE_FRONTEND_COMMIT="$(git rev-parse --short=10 HEAD)" weknora-frontend-builder:upstream-20261020 sh -ec 'npm ci --prefer-offline --no-audit; npm run type-check; find src -name "*.test.ts" -print0 | xargs -0 npm test --; npm run build'

docker build -f deploy/upstream-integration/Dockerfile.frontend-runtime --build-arg SOURCE_REVISION="$(git rev-parse HEAD)" -t weknora-ui:upstream-eecfe4b3 .
