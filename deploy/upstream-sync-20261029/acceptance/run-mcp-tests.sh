#!/bin/bash
set -euo pipefail
cd "$(dirname "$0")/../../.."
docker run --rm --init --label ljc.operation=upstream-20261029 -v "$PWD:/src" -v ljc-sync-20261029-mcp-venv:/proofenv -v ljc-mcp-pip-cache:/root/.cache/pip -w /src/mcp-server python:3.11-slim sh -ec '
 test -x /proofenv/bin/python
 /proofenv/bin/python -m pytest tests -q
 /proofenv/bin/python -m pytest test_chat_async_preflight.py test_chat_sse.py -q
'
