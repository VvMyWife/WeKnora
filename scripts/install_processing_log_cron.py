#!/usr/bin/env python3
"""Install or remove this project's exporter job without replacing other crontab entries."""

import argparse
import subprocess


MARKER = "# weknora-processing-logs"
JOB = (
    "*/5 * * * * /usr/bin/python3 "
    "/home/ubuntu/RAG_Weknora/scripts/export_processing_logs.py --once "
    ">> /home/ubuntu/RAG_Weknora/data/processing-logs/exporter.log 2>&1 "
    + MARKER
)


def updated_crontab(existing: str, remove: bool) -> str:
    lines = [line for line in existing.splitlines() if MARKER not in line]
    if not remove:
        lines.append(JOB)
    return "\n".join(lines) + ("\n" if lines else "")


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--remove", action="store_true")
    args = parser.parse_args()
    current = subprocess.run(["crontab", "-l"], capture_output=True, text=True)
    if current.returncode not in (0, 1):
        raise SystemExit(current.stderr)
    updated = updated_crontab(current.stdout, args.remove)
    subprocess.run(["crontab", "-"], input=updated, text=True, check=True)
    print("Removed" if args.remove else "Installed", MARKER)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
