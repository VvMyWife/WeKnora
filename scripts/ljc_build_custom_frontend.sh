#!/bin/bash
set -euo pipefail
cd "$(dirname "$0")/.."
revision=${LJC_SOURCE_REVISION:-$(git rev-parse HEAD)}
printf 'running\n' > runtime/upstream/custom-frontend.status
trap 'printf "%s\n" "$?" > runtime/upstream/custom-frontend.status' EXIT
docker volume create --label ljc.operation=upstream-20261020 ljc-upstream-custom-node-modules >/dev/null
docker run --rm --name ljc-upstream-custom-ui-build -v "$PWD:/src" -v ljc-upstream-custom-node-modules:/src/frontend-custom/node_modules -v /home/ubuntu/.npm:/root/.npm -w /src/frontend-custom -e VITE_IS_DOCKER=true -e VITE_FRONTEND_COMMIT="$revision" weknora-frontend-builder:upstream-20261020 sh -ec 'npm ci --prefer-offline --no-audit; npm run type-check; find src -name "*.test.ts" -print0 | xargs -0 npm test -- --test-concurrency=2; npm run build'
docker build -f deploy/upstream-integration/Dockerfile.custom-ui-runtime --build-arg SOURCE_REVISION="$revision" -t weknora-ui-custom:ljc-upstream-eecfe4b3 .
