#!/usr/bin/env bash
# Build and capture rollback assets only. Never stop/recreate a live service.
set -euo pipefail
cd "$(dirname "$0")/../.."
repo_root="$PWD"
test "$(git branch --show-current)" = frontend-custom
test "$(git -C frontend-custom branch --show-current)" = frontend-custom
git diff --quiet && git diff --cached --quiet
git -C frontend-custom diff --quiet && git -C frontend-custom diff --cached --quiet
release_tag="$(date +%Y%m%d-%H%M%S)"
release_dir="$repo_root/runtime/archival-citations/$release_tag"
mkdir -p "$release_dir"
chmod 700 "$release_dir"
umask 077
docker inspect WeKnora-app > "$release_dir/app-before.json"
docker inspect WeKnora-frontend-custom-preview > "$release_dir/frontend-before.json"
backend_revision="$(git rev-parse HEAD)"
frontend_revision="$(git -C frontend-custom rev-parse HEAD)"
app_base="$(docker inspect -f '{{.Image}}' WeKnora-app)"
frontend_base="$(docker inspect -f '{{.Image}}' WeKnora-frontend-custom-preview)"
app_image="weknora-app:archival-citations-$release_tag"
frontend_image="archive-intelligence-ui:archival-citations-$release_tag"
rollback_image="weknora-app:archival-rollback-$release_tag"
docker tag "$app_base" "weknora-app:archival-base-$release_tag"
docker tag "$frontend_base" "archive-intelligence-ui:archival-base-$release_tag"
docker build -f deploy/archival-citations/Dockerfile.builder -t weknora-builder:archival-citations deploy/archival-citations > "$release_dir/builder.log" 2>&1
docker run --rm -v "$repo_root:/src" -w /src/frontend-custom node:22-bookworm \
 sh -c 'npm run type-check && npm test -- --test-reporter=dot src/utils/archivalCitations.test.ts src/utils/chatMarkdownRenderer.test.ts src/utils/archivalStream.test.ts && npm run build' > "$release_dir/frontend-build.log" 2>&1
docker run --rm -v "$repo_root:/src" -v "$repo_root/.go-modcache:/go/pkg/mod" -v weknora-go-build-cache:/root/.cache/go-build -w /src weknora-builder:archival-citations \
 sh -c "go test ./internal/archivalcitations ./internal/modelcontext ./internal/application/service/chat_pipeline ./internal/agent ./internal/agent/tools ./internal/handler/session ./internal/handler ./internal/router ./internal/container ./internal/database ./internal/archivemapping -count=1 && go build -buildvcs=false -o 'runtime/archival-citations/$release_tag/WeKnora' ./cmd/server" > "$release_dir/backend-build.log" 2>&1
mkdir -p "$release_dir/migrations/versioned" "$release_dir/migrations/sqlite"
cp migrations/versioned/000094_archival_citations.*.sql "$release_dir/migrations/versioned/"
cp migrations/sqlite/000015_archival_citations.*.sql "$release_dir/migrations/sqlite/"
# Stage only the custom dist. The original frontend is not read or overwritten.
cp -a frontend-custom/dist "$release_dir/dist"
docker build --build-arg "RUNTIME_BASE=weknora-app:archival-base-$release_tag" --build-arg "SOURCE_REVISION=$backend_revision" -f deploy/archival-citations/Dockerfile.app -t "$app_image" "$release_dir" > "$release_dir/app-image.log" 2>&1
docker build --build-arg "RUNTIME_BASE=weknora-app:archival-base-$release_tag" -f deploy/archival-citations/Dockerfile.rollback -t "$rollback_image" "$release_dir" > "$release_dir/rollback-image.log" 2>&1
for candidate in "$app_image" "$rollback_image"; do
 docker run --rm --user appuser --entrypoint sh "$candidate" -c 'test -r /app/migrations/versioned/000094_archival_citations.up.sql && test -x /app/migrations/versioned && test -x /app/WeKnora'
done
docker build --build-arg "RUNTIME_BASE=archive-intelligence-ui:archival-base-$release_tag" --build-arg "SOURCE_REVISION=$frontend_revision" -f deploy/archival-citations/Dockerfile.frontend -t "$frontend_image" "$release_dir" > "$release_dir/frontend-image.log" 2>&1
python3 - "$release_dir" "$backend_revision" "$frontend_revision" "$app_image" "$frontend_image" "$rollback_image" <<'PY'
import hashlib,json,pathlib,subprocess,sys
folder,backend,frontend,app,ui,rollback=sys.argv[1:]
p=pathlib.Path(folder)
def image_id(name):return subprocess.check_output(['docker','image','inspect','-f','{{.Id}}',name],text=True).strip()
before_app=json.loads((p/'app-before.json').read_text())[0]
before_ui=json.loads((p/'frontend-before.json').read_text())[0]
manifest={'state':'prepared_not_deployed','backend_commit':backend,'frontend_commit':frontend,'app_image':app,'app_image_id':image_id(app),'frontend_image':ui,'frontend_image_id':image_id(ui),'rollback_image':rollback,'rollback_image_id':image_id(rollback),'previous_app_id':before_app['Image'],'previous_frontend_id':before_ui['Image'],'binary_sha256':hashlib.sha256((p/'WeKnora').read_bytes()).hexdigest(),'feature_flag':'ARCHIVAL_CITATIONS_ENABLED','database_migration_applied':False}
(p/'release.json').write_text(json.dumps(manifest,ensure_ascii=False,indent=2)+'\n')
print(json.dumps(manifest,ensure_ascii=False,indent=2))
PY
printf 'Prepared only: %s\n' "$release_dir"
