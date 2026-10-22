#!/bin/bash
# Build only. Never restarts production, touches compose or migrates data.
set -euo pipefail
cd "$(dirname "$0")/.."
exec 9>/tmp/ljc-inline-citations-build.lock
flock -n 9 || { echo 'Another inline-citation build is active'; exit 1; }
test -z "$(git status --porcelain --untracked-files=no)" || { echo 'Commit source before build'; exit 1; }
revision=$(git rev-parse HEAD)
release="runtime/inline-citations/$revision"
# Existing runtime/ was created by the persistent root build container. Create
# and grant ownership only of this feature's new directory, not the workspace.
docker run --rm -u 0 -e HOST_UID="$(id -u)" -e HOST_GID="$(id -g)" -v "$PWD/runtime:/runtime" weknora-builder:upstream-20261020 sh -ec 'mkdir -p /runtime/inline-citations; chown "$HOST_UID:$HOST_GID" /runtime/inline-citations'
mkdir -p "$release"
expected_app=sha256:76399c0e703a7118ec3d7b5f7cd5140388f0c147bad32c2df6b154b1f92c51e1
expected_ui=sha256:66b71bd73accc993f5c641dbeab4bc2bb8c999eae1f6050428c1fc147892043f
freeze_base() {
  local name=$1 expected=$2 snapshot=$3 actual
  actual=$(docker inspect -f '{{.Image}}' "$name")
  local service=app
  test "$name" != WeKnora-frontend-custom-preview || service=frontend
  local released
  released=$(python3 -c 'import json,sys; print(json.load(open("deploy/inline-citations/compose.release.json"))["services"][sys.argv[1]]["image"])' "$service")
  test "$actual" = "$expected" || test "$actual" = "$released" || { echo "Baseline changed for $name; inspect before rebuilding"; exit 1; }
  if docker image inspect "$snapshot" >/dev/null 2>&1; then
    test "$(docker image inspect -f '{{.Id}}' "$snapshot")" = "$expected" || exit 1
  else docker image tag "$expected" "$snapshot"; fi
}
freeze_base WeKnora-app "$expected_app" weknora-app:ljc-inline-baseline-20261021
freeze_base WeKnora-frontend-custom-preview "$expected_ui" weknora-ui-custom:ljc-inline-baseline-20261021
build_app() {
  docker run --rm --name ljc-inline-app-build -e COMMIT_ID="$revision" -e VERSION=0.8.3 -e BUILD_TIME="$(date -u +%FT%TZ)" -e GOFLAGS=-buildvcs=false -v "$PWD:/src" -v /home/ubuntu/RAG_Weknora/.go-modcache:/go/pkg/mod -v weknora-go-build-cache:/root/.cache/go-build -w /src weknora-builder:upstream-20261020 sh -ec '
    export GO_VERSION=$(go version | cut -d" " -f3)
    go test ./internal/archivalcitations ./internal/modelcontext ./internal/types ./internal/agent ./internal/application/service/chat_pipeline
    make build-prod GO_BUILD_TAGS=anydoc BINARY_NAME=runtime/inline-citations/WeKnora
  '
  docker build -f deploy/inline-citations/Dockerfile.app --build-arg SOURCE_REVISION="$revision" -t "weknora-app:ljc-inline-${revision:0:9}" .
}
build_ui() {
  docker run --rm --name ljc-inline-ui-build -e VITE_IS_DOCKER=true -e VITE_FRONTEND_COMMIT="$revision" -v "$PWD:/src" -v ljc-upstream-custom-node-modules:/src/frontend-custom/node_modules -w /src/frontend-custom weknora-frontend-builder:upstream-20261020 sh -ec '
    npm run type-check
    find src -name "*.test.ts" -print0 | xargs -0 npm test -- --test-concurrency=2
    npm run build
  '
  docker build -f deploy/inline-citations/Dockerfile.frontend --build-arg SOURCE_REVISION="$revision" -t "weknora-ui-custom:ljc-inline-${revision:0:9}" .
}
printf 'source=%s\napp_base=%s\nui_base=%s\n' "$revision" "$expected_app" "$expected_ui" > "$release/build.status"
build_app > "$release/app.log" 2>&1 & app_pid=$!
build_ui > "$release/ui.log" 2>&1 & ui_pid=$!
echo "Building $revision; logs: $release; app_pid=$app_pid ui_pid=$ui_pid"
app_rc=0; ui_rc=0
wait "$app_pid" || app_rc=$?
wait "$ui_pid" || ui_rc=$?
printf 'app_exit=%s\nui_exit=%s\n' "$app_rc" "$ui_rc" >> "$release/build.status"
test "$app_rc" = 0 && test "$ui_rc" = 0
docker image inspect "weknora-app:ljc-inline-${revision:0:9}" "weknora-ui-custom:ljc-inline-${revision:0:9}" --format '{{.Id}} {{index .Config.Labels "org.opencontainers.image.revision"}}'
echo 'Build complete. Not deployed; restart authorization still required.'
