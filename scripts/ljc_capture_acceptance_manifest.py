#!/usr/bin/env python3
"""Capture candidate artifact identities without exporting environment or secrets."""
import argparse
import datetime
import hashlib
import json
import subprocess
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]


def run(*args):
    return subprocess.check_output(args, text=True, cwd=ROOT).strip()


def capture(private_dir):
    boundary = json.loads(run("python3", "scripts/ljc_integration_selfcheck.py",
                              "--private-dir", str(private_dir)))
    policy = json.loads((ROOT / "docs/upstream-integration/policy.json").read_text())
    manifest = {
        "captured_at": datetime.datetime.now(datetime.timezone.utc).isoformat(),
        "integration_head": run("git", "rev-parse", "HEAD"),
        "upstream_target": policy["upstream_target"],
        "production_deployed": False,
        "maintenance_branch_advanced": run("git", "rev-parse", "ljc/custom") != policy["custom_baseline"],
        "artifacts": [],
        "static_assets": {},
        "evidence": [],
        "production_boundary": boundary,
    }
    for component in ("app", "official-ui", "custom-ui", "docreader", "neo4j"):
        name = "ljc-upstream-acceptance-" + component
        container = json.loads(run("docker", "inspect", name))[0]
        image = json.loads(run("docker", "image", "inspect", container["Image"]))[0]
        labels = image["Config"].get("Labels") or {}
        manifest["artifacts"].append({
            "component": component, "container": name,
            "image_tag": container["Config"]["Image"],
            "image_id": container["Image"], "created": image["Created"],
            "labels": {k: v for k, v in labels.items()
                       if k.startswith(("org.opencontainers.image.", "ljc."))},
            "state": container["State"]["Status"],
            "acceptance": ("review-only; frozen custom UI; pending module decision and locale fixes"
                           if component == "custom-ui" else "candidate; not production"),
        })
    for frontend in ("frontend", "frontend-custom"):
        directory = ROOT / frontend / "dist"
        if not directory.is_dir():
            continue
        files = sorted(p for p in directory.rglob("*") if p.is_file())
        digest = hashlib.sha256()
        for path in files:
            row = path.relative_to(directory).as_posix() + " " + hashlib.sha256(path.read_bytes()).hexdigest()
            digest.update((row + "\n").encode())
        manifest["static_assets"][frontend] = {
            "file_count": len(files), "tree_sha256": digest.hexdigest(),
            "scope": "workspace dist; container identity is recorded separately",
        }
    for path in sorted(private_dir.glob("*results.json")):
        manifest["evidence"].append({
            "file": path.name, "sha256": hashlib.sha256(path.read_bytes()).hexdigest(),
            "bytes": path.stat().st_size,
        })
    return manifest


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--private-dir", type=Path,
                        default=Path("/home/ubuntu/weknora-fork-migration/20261020-upstream"))
    parser.add_argument("--output", type=Path,
                        default=ROOT / "docs/upstream-integration/artifact-manifest.json")
    args = parser.parse_args()
    data = capture(args.private_dir)
    args.output.write_text(json.dumps(data, ensure_ascii=False, indent=2) + "\n")
    print("Candidate artifact manifest saved:", args.output)
    print("Production boundary verified; custom UI remains review-only.")


if __name__ == "__main__":
    main()
