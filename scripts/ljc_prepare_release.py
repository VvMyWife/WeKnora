#!/usr/bin/env python3
"""Prepare an isolated backend-only image context; never deploy or migrate."""
import datetime
import hashlib
import json
from pathlib import Path
import shutil
import subprocess
import uuid

from ljc_status import inspect

root = Path(__file__).resolve().parents[1]
state = inspect(root)
if state["pending_git_operations"] or not state["frontend_frozen"] or state["working_tree"]:
    raise SystemExit("Commit and verify the candidate before preparing a release")
binary = root / "runtime/fork/WeKnora"
if not binary.is_file():
    raise SystemExit("Build runtime/fork/WeKnora first")
name = "release-" + datetime.datetime.now().strftime("%Y%m%d-%H%M%S") + "-" + uuid.uuid4().hex[:8]
dest = root / "runtime/fork" / name
dest.mkdir(mode=0o700)
shutil.copy2(binary, dest / "WeKnora")
shutil.copytree(root / "migrations", dest / "migrations")
manifest = {
    "state": "prepared_not_deployed",
    "backend_commit": state["head"],
    "binary_sha256": hashlib.sha256(binary.read_bytes()).hexdigest(),
    "frontend_frozen": True,
    "database_migrations_executed": False,
}
(dest / "release.json").write_text(json.dumps(manifest, indent=2) + "\n")
print(dest)
