#!/usr/bin/env python3
"""Read-only frontend freeze and interrupted-integration checks."""
import argparse
import hashlib
import json
from pathlib import Path
import subprocess


def git(root, *args):
    return subprocess.check_output(["git", "-C", str(root), *args], text=True).strip()


def inspect(root):
    root = Path(root).resolve()
    baseline = json.loads((root / "docs/fork-maintenance/baseline.json").read_text())
    freeze = json.loads((root / "docs/fork-maintenance/frontend-freeze.json").read_text())
    failures = []
    expected = {item["path"]: item["sha256"] for item in freeze["files"]}
    for name, digest in expected.items():
        path = root / name
        if not path.is_file() or hashlib.sha256(path.read_bytes()).hexdigest() != digest:
            failures.append(name)
    tracked = set(git(root, "ls-files", "frontend-custom").splitlines())
    failures.extend(sorted(tracked - set(expected)))
    failures.extend(sorted(set(expected) - tracked))
    original = git(root, "diff", "--name-only", baseline["backend_head"], "--", "frontend")
    policy_path = root / "docs/upstream-integration/policy.json"
    if policy_path.is_file():
        policy = json.loads(policy_path.read_text())
        for ref in policy.get("official_frontend_candidates", []):
            candidate = git(root, "diff", "--name-only", ref, "--", "frontend")
            overrides = policy.get("official_frontend_overrides", {})
            paths = set(candidate.splitlines())
            if paths.issubset(overrides) and all(
                (root / name).is_file()
                and hashlib.sha256((root / name).read_bytes()).hexdigest() == digest
                for name, digest in overrides.items()
            ):
                original = ""
                break
    failures.extend(original.splitlines())
    pending = []
    for name in ("CHERRY_PICK_HEAD", "MERGE_HEAD", "REVERT_HEAD", "rebase-merge", "rebase-apply", "sequencer"):
        path = Path(git(root, "rev-parse", "--git-path", name))
        if not path.is_absolute():
            path = root / path
        if path.exists():
            pending.append(name)
    return {
        "root": str(root),
        "branch": git(root, "branch", "--show-current"),
        "head": git(root, "rev-parse", "HEAD"),
        "frontend_frozen": not failures,
        "changed_frontend_paths": sorted(set(failures)),
        "pending_git_operations": pending,
        "working_tree": git(root, "status", "--porcelain"),
        "next_action": "Inspect conflicts and current HEAD; continue or abort the existing Git operation. Never replay automatically."
        if pending else "Resolve frontend changes before integrating." if failures else
        "Ready for approved integration; inspect status before beginning.",
    }


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--repo", type=Path, default=Path(__file__).resolve().parents[1])
    args = parser.parse_args()
    result = inspect(args.repo)
    print(json.dumps(result, ensure_ascii=False, indent=2))
    return 3 if result["pending_git_operations"] else 2 if not result["frontend_frozen"] else 0


if __name__ == "__main__":
    raise SystemExit(main())
