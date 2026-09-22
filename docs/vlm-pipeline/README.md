# WeKnora VLM evidence pipeline

## Scope

This deployment makes the built-in `paddleocr_vl` parser the first matching rule
for newly uploaded `PDF`, `JPG`, `JPEG`, and `PNG` files in the FluScri tenant.
It does not reparse, overwrite, or otherwise alter historical knowledge-base
documents.

The parser routing is stored per knowledge base. Future files matching the four
extensions take the PaddleOCR-VL path; other types retain their existing
WeKnora rules.

## Evidence handling

When a VLM returns OCR text, WeKnora now removes complete `<think>`, `<analysis>`,
and `<reasoning>` blocks before text is stored or embedded. An unterminated or
leftover reasoning tag rejects that OCR result instead of allowing model control
text into the knowledge base.

The asynchronous document task uses WeKnora's existing retry policy. If all
retries fail because of rejected VLM output, the document is marked failed and
disabled (quarantined). It is not silently indexed with unsafe or incomplete
content.

`image_ocr` chunks are included in the chunk-reading tool. For exact table fields
such as names or ID numbers, the agent instruction requires reading that source
chunk rather than relying on a document caption or generated summary.

## Private audit records

Set `WEKNORA_VLM_AUDIT_DIR=/data/vlm_audit` in the app container and mount
`/home/ubuntu/RAG_Weknora/data/vlm_audit` there. Each record includes raw and
cleaned OCR output for diagnosis, is stored beneath a hashed knowledge ID, and
uses private permissions (`0700` directory, `0600` file). These records may
contain personal information: keep the mount local, restrict server access, and
delete records only under the applicable retention policy.

## Verify source tests

From the repository on the server:

```bash
cd /home/ubuntu/RAG_Weknora
docker run --rm -e GOPROXY=https://goproxy.cn,direct \
  -v /home/ubuntu/RAG_Weknora:/src -w /src golang:1.26-bookworm \
  /usr/local/go/bin/go test ./internal/application/service ./internal/agent/tools
```

## Deploy and verify

The deployment override is `docker-compose.vlm-clean.yml`.

```bash
cd /home/ubuntu/RAG_Weknora
mkdir -p data/vlm_audit
chmod 700 data/vlm_audit
docker compose -f docker-compose.yml -f docker-compose.vlm-clean.yml up -d --no-deps app
curl -fsS http://127.0.0.1:18081/health
```

For a user-facing functional check, create a *new* disposable knowledge base,
upload a new table image, wait until parsing completes, then ask for a precise
table value. In the chunk viewer, confirm an `image_ocr` source chunk contains
the exact original value before treating it as answer evidence. Do not use the
existing `0826test` history for this validation.

## Rollback

The release uses a versioned app image tag. To return to the previous image,
change only the `app.image` value in `docker-compose.vlm-clean.yml` to the prior
known-good tag and repeat the deploy command. Parser routing and historic
documents stay untouched.
