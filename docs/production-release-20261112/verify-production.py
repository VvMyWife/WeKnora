"""Verify the upgraded public entry point from an approved intranet client."""

import argparse
import json
import time
from datetime import datetime, timezone
from pathlib import Path
from urllib.error import HTTPError
from urllib.request import Request, urlopen


def call(url, payload=None, timeout=20):
    body = None if payload is None else json.dumps(payload).encode()
    headers = {"Content-Type": "application/json"} if body is not None else {}
    try:
        with urlopen(Request(url, data=body, headers=headers), timeout=timeout) as response:
            return response.status, json.load(response)
    except HTTPError as error:
        return error.code, {}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--origin", default="http://10.20.0.15:18182")
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    origin = args.origin.rstrip("/")
    base = f"{origin}/api/openai/v1"
    report = {"checked_at": datetime.now(timezone.utc).isoformat(), "origin": origin}

    try:
        status, models = call(f"{base}/models")
        report["models_status"] = status
        assert status == 200 and "policy-qa" in {row["id"] for row in models["data"]}
        status, _ = call(f"{origin}/api/v1/knowledge-bases")
        report["management_without_login_status"] = status
        assert status in {401, 403}

        start = time.monotonic()
        status, completion = call(
            f"{base}/chat/completions",
            {
                "model": "policy-qa",
                "messages": [{
                    "role": "user",
                    "content": "依据 GB/T 20988，简要概括第三级灾难恢复的数据备份要求，附引用依据。",
                }],
                "max_tokens": 1024,
            },
            timeout=310,
        )
        report["completion_status"] = status
        report["completion_seconds"] = round(time.monotonic() - start, 2)
        assert status == 200
        report["answer_chars"] = len(completion["choices"][0]["message"]["content"].strip())
        report["total_tokens"] = completion["usage"]["total_tokens"]
        assert report["answer_chars"] and report["total_tokens"] > 0

        citations = completion.get("citations", [])
        report["citation_count"] = len(citations)
        status, manifest = call(f"{base}/responses/{completion['id']}/citations")
        report["citation_manifest_status"] = status
        assert status == 200 and manifest["citations"] == citations

        report["citation_detail_passed"] = 0
        for citation in citations:
            path = citation["detail_url"]
            assert path.startswith("/api/openai/v1/")
            status, detail = call(f"{origin}{path}")
            assert status == 200 and detail.get("snapshot")
            report["citation_detail_passed"] += 1
        report["passed"] = True
    finally:
        args.output.write_text(json.dumps(report, ensure_ascii=False, indent=2) + "\n", encoding="utf-8", newline="\n")
        print(json.dumps(report, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
