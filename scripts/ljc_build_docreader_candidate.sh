#!/bin/bash
set -euo pipefail
cd "$(dirname "$0")/.."
printf 'running\n' > runtime/upstream/docreader-image.status
trap 'printf "%s\n" "$?" > runtime/upstream/docreader-image.status' EXIT
source_revision=$(git rev-parse HEAD)
image=weknora-docreader:ljc-upstream-eecfe4b3
docker build -f deploy/upstream-integration/Dockerfile.docreader --build-arg SOURCE_REVISION="$source_revision" -t "$image" .
docker run --rm --init -v "$PWD/testdata:/app/testdata:ro" --label ljc.operation=upstream-20261020 --entrypoint /app/.venv/bin/python "$image" -m unittest discover -s docreader/tests -v
