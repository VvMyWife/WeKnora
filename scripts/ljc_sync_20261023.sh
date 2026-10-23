#!/bin/bash
# Safe repeatable tests/builds only. Never deploys or migrates production.
set -euo pipefail
cd "$(dirname "$0")/.."
runtime=runtime/upstream-20261023
mkdir -p "$runtime"
job=${1:-status}
if [[ "$job" == status ]]; then
 git status --short --branch
 git log -4 --oneline
 for f in "$runtime"/*.status; do [[ ! -f "$f" ]] || printf '%s: %s\n' "$f" "$(cat "$f")"; done
 docker ps --filter label=ljc.operation=upstream-20261023 --format '{{.Names}} {{.Status}}'
 exit 0
fi
case "$job" in backend|custom-ui|official-ui|docreader|images) ;; *) echo 'Unknown safe job' >&2; exit 64;; esac
exec 9>"$runtime/$job.lock"
flock -n 9 || { echo "$job already running; inspect status, never replay"; exit 75; }
echo "$$" > "$runtime/$job.pid"
echo running > "$runtime/$job.status"
trap 'rc=$?; echo "$rc" > "$runtime/$job.status"; exit "$rc"' EXIT
builder=weknora-builder:upstream-20261020
mods=/home/ubuntu/RAG_Weknora/.go-modcache
if [[ "$job" == backend ]]; then
 hosts=()
 while read -r host; do
  [[ -z "$host" || "$host" == \#* ]] && continue
  hosts+=(--add-host "$host:93.184.216.34")
 done < deploy/upstream-integration/test-hosts.txt
 docker run --rm --name ljc-sync-20261023-backend-tests --label ljc.operation=upstream-20261023 "${hosts[@]}" -e GODEBUG=netdns=cgo -e LOG_FORMAT= -v "$PWD:/src" -v "$mods:/go/pkg/mod" -v weknora-go-build-cache:/root/.cache/go-build -w /src "$builder" sh -ec '
 go test ./internal/... ./cmd/server/... ./cmd/ljc-migrate/... -count=1 -timeout=5m
 (cd third_party/asynq && go test . ./internal/context -run "^TestUnlimited" -count=1)
 go test -tags anydoc ./internal/infrastructure/docparser/... -count=1
 go build -tags anydoc -o runtime/upstream-20261023/WeKnora-native-check ./cmd/server
 go build -o runtime/upstream-20261023/ljc-migrate ./cmd/ljc-migrate'
elif [[ "$job" == custom-ui || "$job" == official-ui ]]; then
 folder=frontend-custom; volume=ljc-upstream-custom-node-modules
 if [[ "$job" == official-ui ]]; then folder=frontend; volume=ljc-upstream-official-node-modules; fi
 docker run --rm --name "ljc-sync-20261023-$job-tests" --label ljc.operation=upstream-20261023 -e VITE_IS_DOCKER=true -e VITE_FRONTEND_COMMIT="$(git rev-parse HEAD)" -v "$PWD:/src" -v "$volume:/src/$folder/node_modules" -v /home/ubuntu/.npm:/root/.npm -w "/src/$folder" weknora-frontend-builder:upstream-20261020 sh -ec '
 test -x node_modules/.bin/vue-tsc || npm ci --prefer-offline --no-audit
 npm run type-check
 find src -type f \( -name "*.test.ts" -o -name "*.test.mjs" \) -print0 | xargs -0 npm test -- --test-concurrency=2
 npm run build'
elif [[ "$job" == docreader ]]; then
 docker build -f deploy/upstream-sync-20261023/Dockerfile.docreader --build-arg SOURCE_REVISION="$(git rev-parse HEAD)" -t weknora-docreader:ljc-sync-preflight-20261023 .
 docker run --rm --init --name ljc-sync-20261023-docreader-tests --label ljc.operation=upstream-20261023 -v "$PWD/testdata:/app/testdata:ro" --entrypoint /app/.venv/bin/python weknora-docreader:ljc-sync-preflight-20261023 -m unittest discover -s docreader/tests -v
else
 test -z "$(git status --porcelain --untracked-files=no)" || { echo 'Commit source before building'; exit 1; }
 git merge-base --is-ancestor bc1ade36 HEAD
 revision=$(git rev-parse HEAD); short=${revision:0:9}
 docker run --rm --name ljc-sync-20261023-app-build --label ljc.operation=upstream-20261023 -e COMMIT_ID="$revision" -e VERSION=0.8.3 -e BUILD_TIME="$(date -u +%FT%TZ)" -e GOFLAGS=-buildvcs=false -v "$PWD:/src" -v "$mods:/go/pkg/mod" -v weknora-go-build-cache:/root/.cache/go-build -w /src "$builder" sh -ec '
 export GO_VERSION=$(go version | cut -d" " -f3)
 make build-prod GO_BUILD_TAGS=anydoc BINARY_NAME=runtime/upstream-20261023/WeKnora
 go build -o runtime/upstream-20261023/ljc-migrate ./cmd/ljc-migrate'
 docker run --rm --name ljc-sync-20261023-ui-artifact --label ljc.operation=upstream-20261023 -e VITE_IS_DOCKER=true -e VITE_FRONTEND_COMMIT="$revision" -v "$PWD:/src" -v ljc-upstream-custom-node-modules:/src/frontend-custom/node_modules -w /src/frontend-custom weknora-frontend-builder:upstream-20261020 npm run build
 docker build -f deploy/upstream-sync-20261023/Dockerfile.app --build-arg SOURCE_REVISION="$revision" -t "weknora-app:ljc-sync-$short" .
 docker build -f deploy/upstream-sync-20261023/Dockerfile.frontend --build-arg SOURCE_REVISION="$revision" -t "weknora-ui-custom:ljc-sync-$short" .
 docker build -f deploy/upstream-sync-20261023/Dockerfile.docreader --build-arg SOURCE_REVISION="$revision" -t "weknora-docreader:ljc-sync-$short" .
 python3 - "$revision" "$short" <<'PY'
import json,subprocess,sys
from pathlib import Path
revision,short=sys.argv[1:]
images={}
for role,prefix in [('app','weknora-app'),('frontend','weknora-ui-custom'),('docreader','weknora-docreader')]:
 tag=prefix+':ljc-sync-'+short
 d=json.loads(subprocess.check_output(['docker','image','inspect',tag]))[0]
 assert d['Config']['Labels']['org.opencontainers.image.revision']==revision
 images[role]={'tag':tag,'id':d['Id']}
Path('runtime/upstream-20261023/artifact-manifest.json').write_text(json.dumps({'source_sha':revision,'upstream_sha':'bc1ade36b933ff7689f1659d670cfd5683c174c5','images':images,'production_deployed':False},indent=2))
print(json.dumps({'source':revision,'images':images,'production_deployed':False}))
PY
fi
