#!/usr/bin/env bash
set -euo pipefail
# Manual stop is respected by Restart=always. Do not kill a stale PID that may
# now belong to another process after a reboot.
sudo systemctl stop weknora-bge-m3.service
systemctl is-active weknora-bge-m3.service || true
