#!/usr/bin/env python3
"""Export readable, per-knowledge-base diagnostics from WeKnora's live data.

Read-only against PostgreSQL and Docker logs. The only writes are under
data/processing-logs. Designed for a five-minute cron run; flock prevents overlap.
"""

from __future__ import annotations

import argparse
import collections
import datetime as dt
import fcntl
import json
import os
from pathlib import Path
import re
import subprocess
import sys
import tempfile


ROOT = Path(__file__).resolve().parents[1]
OUT = ROOT / "data" / "processing-logs"
UUID = re.compile(r"^[0-9a-fA-F-]{36}$")
FAILURE = re.compile(r"\[(KnowledgeUploadFailure|KnowledgeProcessFailure)\]")
KB_ID = re.compile(r"\bkb_id=([0-9a-fA-F-]{36})\b")
KNOWLEDGE_ID = re.compile(r"\bknowledge_id=([0-9a-fA-F-]{36})\b")


def one_line(value: object, limit: int = 900) -> str:
    text = " ".join(str(value or "").split())
    return text[:limit] + ("…" if len(text) > limit else "")


def psql_rows(sql: str) -> list[dict]:
    command = [
        "docker", "exec", "WeKnora-postgres", "psql", "-X", "-A", "-t",
        "-v", "ON_ERROR_STOP=1", "-U", "weknora", "-d", "weknora", "-c", sql,
    ]
    result = subprocess.run(command, capture_output=True, text=True, check=True)
    return [json.loads(line) for line in result.stdout.splitlines() if line.strip()]


def load_database() -> tuple[dict[str, str], dict[str, dict], dict[str, list[dict]], dict[str, list[dict]]]:
    bases = {
        row["id"]: row["name"] + ("（已删除）" if row["deleted_at"] else "")
        for row in psql_rows("""
            SELECT row_to_json(t) FROM (
              SELECT id, name, deleted_at FROM knowledge_bases
            ) AS t
        """)
    }
    docs = {
        row["id"]: row
        for row in psql_rows("""
            SELECT row_to_json(t) FROM (
              SELECT id, knowledge_base_id, folder_path, file_name,
                     parse_status, summary_status,
                     left(coalesce(error_message, ''), 1000) AS error_message
              FROM knowledges WHERE deleted_at IS NULL
            ) AS t
        """)
    }
    failures: dict[str, list[dict]] = collections.defaultdict(list)
    for row in psql_rows("""
        WITH latest AS (
          SELECT knowledge_id, max(attempt) AS attempt
          FROM knowledge_processing_spans GROUP BY knowledge_id
        ), ranked AS (
          SELECT k.knowledge_base_id, k.id AS knowledge_id,
                 k.folder_path, k.file_name, s.attempt, s.name,
                 s.error_code, s.error_message,
                 left(coalesce(s.error_detail, ''), 1500) AS error_detail,
                 s.updated_at,
                 row_number() OVER (
                   PARTITION BY k.knowledge_base_id ORDER BY s.updated_at DESC, s.id DESC
                 ) AS rank,
                 count(*) OVER (PARTITION BY k.knowledge_base_id) AS total_failures
          FROM knowledge_processing_spans s
          JOIN latest l ON l.knowledge_id = s.knowledge_id AND l.attempt = s.attempt
          JOIN knowledges k ON k.id = s.knowledge_id
          WHERE s.status = 'failed' AND k.deleted_at IS NULL
        )
        SELECT row_to_json(t) FROM (
          SELECT * FROM ranked WHERE rank <= 500
        ) AS t
    """):
        failures[row["knowledge_base_id"]].append(row)
    stages: dict[str, list[dict]] = collections.defaultdict(list)
    for row in psql_rows("""
        WITH latest AS (
          SELECT knowledge_id, max(attempt) AS attempt
          FROM knowledge_processing_spans GROUP BY knowledge_id
        ), classified AS (
          SELECT k.knowledge_base_id, s.status,
                 CASE
                   WHEN s.name LIKE 'postprocess.graph.%' THEN 'postprocess.graph'
                   WHEN s.name IN ('docreader', 'chunking', 'embedding', 'multimodal',
                                   'postprocess', 'postprocess.summary', 'postprocess.wiki')
                     THEN s.name
                 END AS stage
          FROM knowledge_processing_spans s
          JOIN latest l ON l.knowledge_id = s.knowledge_id AND l.attempt = s.attempt
          JOIN knowledges k ON k.id = s.knowledge_id
          WHERE k.deleted_at IS NULL
        )
        SELECT row_to_json(t) FROM (
          SELECT knowledge_base_id, stage, status, count(*) AS count
          FROM classified WHERE stage IS NOT NULL
          GROUP BY knowledge_base_id, stage, status
        ) AS t
    """):
        stages[row["knowledge_base_id"]].append(row)
    return bases, docs, failures, stages


def document_path(doc: dict) -> str:
    return "/".join(part for part in (doc.get("folder_path"), doc.get("file_name")) if part) or doc["id"]


def format_report(kb_id: str, name: str, docs: list[dict], failures: list[dict], stages: list[dict], now: str) -> str:
    parse = collections.Counter(d.get("parse_status") or "未设置" for d in docs)
    summary = collections.Counter(d.get("summary_status") or "未设置" for d in docs)
    bad_docs = [d for d in docs if d.get("parse_status") == "failed" or d.get("summary_status") == "failed"]
    lines = [
        f"知识库：{name}", f"ID：{kb_id}", f"更新时间：{now}", f"文档总数：{len(docs)}", "",
        "解析状态：" + "，".join(f"{k} {v}" for k, v in sorted(parse.items())),
        "摘要状态：" + "，".join(f"{k} {v}" for k, v in sorted(summary.items())),
        "", "处理阶段（最近一次尝试）：",
    ]
    by_stage: dict[str, dict[str, int]] = collections.defaultdict(dict)
    for item in stages:
        by_stage[item["stage"]][item["status"]] = item["count"]
    for stage, statuses in sorted(by_stage.items()):
        lines.append(f"- {stage}：" + "，".join(f"{status} {count}" for status, count in sorted(statuses.items())))
    lines.extend(["", f"当前失败文档：{len(bad_docs)}"])
    for doc in bad_docs[:500]:
        lines.append(
            f"- {document_path(doc)} | 解析={doc.get('parse_status')} | "
            f"摘要={doc.get('summary_status')} | {one_line(doc.get('error_message'))}"
        )
    if len(bad_docs) > 500:
        lines.append(f"（其余 {len(bad_docs) - 500} 份未展开）")
    total_failures = failures[0]["total_failures"] if failures else 0
    lines.extend(["", f"最近一次处理尝试中的失败步骤：{total_failures}（最多展示最近 500 条）"])
    for row in failures:
        lines.append(
            f"- {row['updated_at']} | {document_path(row)} | "
            f"尝试#{row['attempt']} | {row['name']} | {row['error_code']} | "
            f"{one_line(row['error_message'])}"
        )
        if row.get("error_detail") and row["error_detail"] != row.get("error_message"):
            lines.append(f"    原因：{one_line(row['error_detail'], 1500)}")
    lines.extend(["", "新增失败事件详见本目录 events.log；完整处理轨迹仍保存在数据库中。", ""])
    return "\n".join(lines)


def atomic_write(path: Path, content: str) -> None:
    fd, temp_name = tempfile.mkstemp(prefix=".report-", dir=path.parent)
    try:
        os.fchmod(fd, 0o640)
        with os.fdopen(fd, "w", encoding="utf-8") as stream:
            stream.write(content)
        os.replace(temp_name, path)
    finally:
        if os.path.exists(temp_name):
            os.unlink(temp_name)


def append_docker_failures(docs: dict[str, dict], known_bases: set[str]) -> int:
    cursor_path = OUT / ".docker-log-cursor"
    cursor = cursor_path.read_text(encoding="ascii").strip() if cursor_path.exists() else ""
    since = cursor or "15m"
    process = subprocess.Popen(
        ["docker", "logs", "--timestamps", "--since", since, "WeKnora-app"],
        stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True,
        errors="replace", bufsize=1,
    )
    assert process.stdout is not None
    latest = cursor
    count = 0
    for line in process.stdout:
        timestamp, separator, message = line.partition(" ")
        if not separator or not timestamp.startswith("20") or timestamp <= cursor:
            continue
        latest = max(latest, timestamp)
        if not FAILURE.search(message):
            continue
        kb_match = KB_ID.search(message)
        doc_match = KNOWLEDGE_ID.search(message)
        doc = docs.get(doc_match.group(1)) if doc_match else None
        kb_id = kb_match.group(1) if kb_match else (doc or {}).get("knowledge_base_id")
        if not kb_id or kb_id not in known_bases or not UUID.fullmatch(kb_id):
            continue
        folder = OUT / kb_id
        folder.mkdir(mode=0o750, exist_ok=True)
        with (folder / "events.log").open("a", encoding="utf-8") as target:
            os.chmod(target.name, 0o640)
            path = f" | 文件={document_path(doc)}" if doc else ""
            target.write(f"{timestamp}{path} | {one_line(message, 2200)}\n")
        count += 1
    if process.wait() != 0:
        raise RuntimeError("docker logs failed; cursor was not advanced")
    if latest != cursor:
        atomic_write(cursor_path, latest + "\n")
    return count


def export_once() -> tuple[int, int]:
    os.umask(0o027)
    OUT.mkdir(parents=True, exist_ok=True, mode=0o750)
    with (OUT / ".export.lock").open("w") as lock:
        try:
            fcntl.flock(lock, fcntl.LOCK_EX | fcntl.LOCK_NB)
        except BlockingIOError:
            print("Previous export still running; next scheduled run will retry.")
            return 0, 0
        bases, docs, failures, stages = load_database()
        grouped: dict[str, list[dict]] = collections.defaultdict(list)
        for doc in docs.values():
            grouped[doc["knowledge_base_id"]].append(doc)
        now = dt.datetime.now(dt.timezone(dt.timedelta(hours=8))).isoformat(timespec="seconds")
        index = [f"WeKnora 知识库处理日志 | 更新于 {now}", "路径：data/processing-logs/<知识库ID>/", ""]
        for kb_id, name in sorted(bases.items(), key=lambda item: item[1]):
            folder = OUT / kb_id
            folder.mkdir(mode=0o750, exist_ok=True)
            kb_docs = grouped.get(kb_id, [])
            atomic_write(folder / "status.txt", format_report(kb_id, name, kb_docs, failures[kb_id], stages[kb_id], now))
            index.append(f"{name} | {kb_id} | {len(kb_docs)} 份文档")
        atomic_write(OUT / "INDEX.txt", "\n".join(index) + "\n")
        event_count = append_docker_failures(docs, set(bases))
        return len(bases), event_count


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--once", action="store_true", help="Run one export (also the cron mode)")
    parser.parse_args()
    try:
        bases, events = export_once()
    except Exception as exc:
        print(f"Processing log export failed: {exc}", file=sys.stderr)
        return 1
    print(f"Exported {bases} knowledge bases; appended {events} failure events.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
