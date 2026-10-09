#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
if [ -f logs/bge-m3.pid ]; then kill "$(cat logs/bge-m3.pid)" 2>/dev/null || true; rm -f logs/bge-m3.pid; fi
