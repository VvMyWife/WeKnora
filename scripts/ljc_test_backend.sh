#!/bin/bash
set -euo pipefail
cd "$(dirname "$0")/.."
mkdir -p runtime/upstream
exec 9>runtime/upstream/backend-tests.lock
flock -n 9 || { echo "Backend tests already running; inspect state before retrying" >&2; exit 75; }
printf 'running\n' > runtime/upstream/backend-tests.status
trap 'printf "%s\n" "$?" > runtime/upstream/backend-tests.status' EXIT
hosts=()
while read -r host; do
 [[ -z "$host" || "$host" == \#* ]] && continue
 hosts+=(--add-host "$host:93.184.216.34")
done < deploy/upstream-integration/test-hosts.txt
docker run --rm "${hosts[@]}" --label ljc.operation=upstream-20261020 --label ljc.stage=backend-tests --name ljc-upstream-backend-tests-$(date +%s) -e GODEBUG=netdns=cgo --dns "${LJC_TEST_DNS:-61.139.2.69}" --dns 223.5.5.5 -v "$PWD:/src" -v /home/ubuntu/RAG_Weknora/.go-modcache:/go/pkg/mod -v weknora-go-build-cache:/root/.cache/go-build -w /src weknora-builder:upstream-20261020 sh -ec '
 go test ./internal/... ./cmd/server/... ./cmd/ljc-migrate/... -count=1
 (cd client && go test ./... -count=1)
 (cd cli && go test ./... -count=1)
 go build -o runtime/upstream/WeKnora ./cmd/server
 go build -o runtime/upstream/ljc-migrate ./cmd/ljc-migrate
 '
