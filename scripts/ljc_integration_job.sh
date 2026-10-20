#!/bin/bash
# Idempotent launcher for repeatable tests/builds only. Never migrations or deploys.
set -euo pipefail
cd "$(dirname "$0")/.."
job=${1:?name required}
case "$job" in
 backend-tests) script=scripts/ljc_test_backend.sh;;
 backend-image) script=scripts/ljc_build_backend_candidate.sh;;
 official-frontend) script=scripts/ljc_build_official_frontend.sh;;
 custom-frontend) script=scripts/ljc_build_custom_frontend.sh;;
 native-assets) script=scripts/ljc_build_native_assets.sh;;
 native-backend) script=scripts/ljc_check_native_backend.sh;;
 docreader-image) script=scripts/ljc_build_docreader_candidate.sh;;
 *) echo "Unknown safe job: $job" >&2;exit 64;;
esac
mkdir -p runtime/upstream/jobs
exec 8>"runtime/upstream/jobs/$job.lock"
flock -n 8 || { echo "$job already running";exit 75; }
printf '%s\n' "$$" >"runtime/upstream/jobs/$job.pid"
printf 'running\n' >"runtime/upstream/jobs/$job.status"
trap 'rc=$?; printf "%s\n" "$rc" >"runtime/upstream/jobs/$job.status"; exit "$rc"' EXIT
bash "$script"
