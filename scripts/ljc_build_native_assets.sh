#!/bin/bash
set -euo pipefail
cd "$(dirname "$0")/.."
mkdir -p runtime/upstream
printf 'running\n' > runtime/upstream/native-assets.status
trap 'printf "%s\n" "$?" > runtime/upstream/native-assets.status' EXIT
for volume in ljc-upstream-cargo-registry ljc-upstream-anydoc-target ljc-upstream-browser-target; do
 docker volume create --label ljc.operation=upstream-20261020 "$volume" >/dev/null
done
docker run --rm --name ljc-upstream-native-build -v "$PWD:/src" -v ljc-upstream-cargo-registry:/usr/local/cargo/registry -v ljc-upstream-anydoc-target:/src/third_party/anydoc-go/target -v ljc-upstream-browser-target:/browser-target -v /home/ubuntu/.npm:/root/.npm -w /src weknora-rust-builder:upstream-20261020 bash -ec '
 bash scripts/build-anydoc-lib.sh
 commit=$(python3 -c "import json; print(json.load(open(\"scripts/browserskill-release.json\"))[\"daemon\"][\"source_commit\"])")
 source=/src/runtime/upstream/BrowserSkill
 if [ ! -d "$source/.git" ]; then git init "$source"; git -C "$source" remote add origin https://github.com/Tencent/BrowserSkill.git; fi
 if ! git -C "$source" cat-file -e "$commit^{commit}" 2>/dev/null; then timeout 300 git -c http.lowSpeedLimit=1024 -c http.lowSpeedTime=60 -C "$source" fetch --depth=1 origin "$commit"; fi
 git -C "$source" checkout --detach "$commit"
 test "$(git -C "$source" rev-parse HEAD)" = "$commit"
 cd "$source"
 npx --yes pnpm@10.17.0 install --frozen-lockfile
 npx --yes pnpm@10.17.0 ext:build:zip
 cargo build --locked --release -p bsk --target-dir /browser-target
 mkdir -p /src/runtime/upstream/browserskill
 install -m 755 /browser-target/release/bsk /src/runtime/upstream/browserskill/bsk
 cp apps/extension/dist/browser-skillextension-0.3.1-chrome.zip /src/runtime/upstream/browserskill/browser-skill-weknora-0.3.1.zip
 cp LICENSE /src/runtime/upstream/browserskill/BrowserSkill-LICENSE
 '
