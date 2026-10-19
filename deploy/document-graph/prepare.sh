#!/usr/bin/env bash
# Prepare only: never restart production or execute database migrations.
set -euo pipefail
cd "$(dirname "$0")/../.."
test "$(git branch --show-current)" = frontend-custom
test "$(git -C frontend-custom branch --show-current)" = frontend-custom
git diff --quiet && git diff --cached --quiet
git -C frontend-custom diff --quiet && git -C frontend-custom diff --cached --quiet
release_tag="$(date +%Y%m%d-%H%M%S)"
release_dir="$PWD/runtime/document-graph/$release_tag"
mkdir -p "$release_dir"
chmod 700 "$release_dir"
umask 077
docker inspect WeKnora-app > "$release_dir/app-before.json"
docker inspect WeKnora-frontend-custom-preview > "$release_dir/frontend-before.json"
docker inspect WeKnora-frontend > "$release_dir/original-before.json"
backend_revision="$(git rev-parse HEAD)"
frontend_revision="$(git -C frontend-custom rev-parse HEAD)"
app_base="$(docker inspect -f '{{.Image}}' WeKnora-app)"
ui_base="$(docker inspect -f '{{.Image}}' WeKnora-frontend-custom-preview)"
docker run --rm -v "$PWD:/src" -v "$PWD/.go-modcache:/go/pkg/mod" -v weknora-go-build-cache:/root/.cache/go-build -w /src weknora-builder:archival-citations \
 sh -c "go test ./internal/application/repository ./internal/application/service ./internal/handler ./internal/router -count=1 && go build -buildvcs=false -o 'runtime/document-graph/$release_tag/WeKnora' ./cmd/server" > "$release_dir/backend-build.log" 2>&1
docker run --rm -v "$PWD:/src" -w /src/frontend-custom node:22-bookworm \
 sh -c 'npm run type-check && npm test -- src/views/knowledge/wiki/documentGraphScope.test.ts && npm run build' > "$release_dir/frontend-build.log" 2>&1
cp -a frontend-custom/dist "$release_dir/dist"
app_image="weknora-app:document-graph-$release_tag"
ui_image="archive-intelligence-ui:document-graph-$release_tag"
docker build --build-arg "RUNTIME_BASE=$app_base" --build-arg "SOURCE_REVISION=$backend_revision" -f deploy/document-graph/Dockerfile.app -t "$app_image" "$release_dir" > "$release_dir/app-image.log" 2>&1
docker build --build-arg "RUNTIME_BASE=$ui_base" --build-arg "SOURCE_REVISION=$frontend_revision" -f deploy/document-graph/Dockerfile.frontend -t "$ui_image" "$release_dir" > "$release_dir/frontend-image.log" 2>&1
python3 - "$release_dir" "$backend_revision" "$frontend_revision" "$app_image" "$ui_image" <<'PY'
import hashlib,json,pathlib,subprocess,sys
folder,backend,frontend,app,ui=sys.argv[1:]
p=pathlib.Path(folder)
def image_id(name): return subprocess.check_output(['docker','image','inspect','-f','{{.Id}}',name],text=True).strip()
old_app=json.loads((p/'app-before.json').read_text())[0]
old_ui=json.loads((p/'frontend-before.json').read_text())[0]
manifest={'state':'prepared_not_deployed','backend_commit':backend,'frontend_commit':frontend,'app_image':app,'app_image_id':image_id(app),'frontend_image':ui,'frontend_image_id':image_id(ui),'previous_app_id':old_app['Image'],'previous_frontend_id':old_ui['Image'],'binary_sha256':hashlib.sha256((p/'WeKnora').read_bytes()).hexdigest(),'database_migrations':False}
(p/'release.json').write_text(json.dumps(manifest,ensure_ascii=False,indent=2)+'\n')
print(json.dumps(manifest,ensure_ascii=False,indent=2))
PY
printf 'Prepared only: %s\n' "$release_dir"
