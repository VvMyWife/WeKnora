#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
unit=weknora-bge-m3.service
if ! systemctl cat "$unit" >/dev/null 2>&1; then
  echo "Install $unit first; see deploy/embedding/README.md" >&2
  exit 1
fi
sudo systemctl start "$unit"
for attempt in $(seq 1 180); do
  if curl --max-time 3 -fsS http://127.0.0.1:18082/health; then
    printf '\nBGE-M3 ready (managed by %s)\n' "$unit"
    exit 0
  fi
  sleep 1
done
echo "BGE-M3 is not ready; inspect logs/bge-m3.log and systemctl status $unit" >&2
exit 1
