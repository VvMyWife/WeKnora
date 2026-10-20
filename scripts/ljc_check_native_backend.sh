#!/bin/bash
set -euo pipefail
cd "$(dirname "$0")/.."
printf 'running\n' > runtime/upstream/native-backend.status
trap 'printf "%s\n" "$?" > runtime/upstream/native-backend.status' EXIT
docker run --rm --name ljc-upstream-native-backend-tests --label ljc.operation=upstream-20261020 -e GODEBUG=netdns=cgo -v "$PWD:/src" -v /home/ubuntu/RAG_Weknora/.go-modcache:/go/pkg/mod -v weknora-go-build-cache:/root/.cache/go-build -w /src weknora-builder:upstream-20261020 sh -ec '
 (cd third_party/asynq && go test . ./internal/context -run "^TestUnlimited" -count=1)
 go test -tags anydoc ./internal/infrastructure/docparser/... -count=1
 go build -tags anydoc -o runtime/upstream/WeKnora-native-check ./cmd/server
 scripts/copy-licenses.sh runtime/upstream/licenses
 mkdir -p runtime/upstream/jieba-dict
 cp -a /go/pkg/mod/github.com/yanyiwu/gojieba@v1.4.7/deps/cppjieba/dict/. runtime/upstream/jieba-dict/
 '
