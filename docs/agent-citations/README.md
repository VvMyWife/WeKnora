# Agent citation stability (2026-10-22)

Authoritative checkout: `/home/ubuntu/RAG_Weknora-upstream-20261020`, branch `ljc/custom`. Commits use Author/Committer `ljc <ljc@local>`. Daily acceptance is http://10.20.0.15:18182; preserve18180. This feature does not add model repair/reanswer calls, migrations, historical rewrites, or model configuration changes.

## Stages

1. `217cd7b6b`: count raw, expanded and dropped citation tags per answer decoder; keep reasoning counters separate. Correlate logs with request/session/message/round. Finalization logs carry message identity. No full model/document text is logged.
2. `b5ee31de8`: at most one bounded model-only tool-result reminder per round, compaction source-association instructions, and request-local unchanged-response decode protection. Forged canonical tags remain rejected.
3. Lifecycle/regression tests and budget reservation for the reminder; see the commit log for exact revision.
4. Build and record acceptance before deployment. Backend restart requires fresh user authorization. No frontend rebuild is required for this backend-only feature; keep the tested notice-hiding18182 image.

## Run tests

```bash
cd /home/ubuntu/RAG_Weknora-upstream-20261020
git status --short
git log -3 --format='%h %an <%ae> / %cn <%ce> | %s'
docker run --rm -v "$PWD:/src" -v /home/ubuntu/RAG_Weknora/.go-modcache:/go/pkg/mod -v weknora-go-build-cache:/root/.cache/go-build -w /src weknora-builder:upstream-20261020 sh -ec 'go test -race ./internal/modelcontext ./internal/agent/compaction ./internal/archivalcitations; go test ./internal/agent ./internal/application/service/chat_pipeline'
docker run --rm -v "$PWD:/src" -v ljc-upstream-custom-node-modules:/src/frontend-custom/node_modules -w /src/frontend-custom weknora-frontend-builder:upstream-20261020 sh -ec 'npm test -- src/utils/citationGrouping.test.ts src/utils/archivalCitations.test.ts; npm run type-check'
```

Tests cover byte-split citation tags/EOF, reasoning isolation, malformed/unknown/history-only references, search/grep/read/graph outputs through SQLite snapshot persistence, parallel registrations, summary/history replay, repeated decode and untrusted canonical-tag rejection. Existing Wiki provenance and no-repair regressions also run. A simulated summary boundary is not a real-model compaction acceptance test.

## Diagnostics

```bash
docker logs --since 30m WeKnora-app 2>&1 | grep '\[Citations\]'
# Substitute the request ID shown by the conversation's request details:
docker logs --since 2h WeKnora-app 2>&1 | grep 'REQUEST_ID' | grep -E '\[Citations\]|LLM Usage|compaction'
```

`stage=decode` reports raw/expanded/malformed/unknown/ineligible/legacy/disabled counts for the answer channel and a separate reasoning count. Tag fragments count once, only after assembly/EOF. `stage=finalize` reports eligible evidence, unique saved citations, rejected canonical references, capture failures, review blocks, validation/save milliseconds and `repair_calls=0`. The request logger supplies the request ID; session/message/round connect the stages. Existing tool-start/result logs identify tools in the same request. Counts do not prove semantic entailment: inspect the cited snapshot during acceptance.

- `raw=0`, `valid=0`, evidence present: no machine-readable reference arrived in that channel.
- raw>0, expanded=0: inspect the specific reject counters; never guess a replacement.
- expanded>0 but valid=0: inspect downstream finalization/save/error records.
- reasoning_raw>0 alone is not an answer citation.

## Build backend only

Use persistent builder/cache with globally configured China sources. Do not replay the old paired frontend/backend release script.

```bash
cd /home/ubuntu/RAG_Weknora-upstream-20261020
test -z "$(git status --porcelain --untracked-files=no)"
REV=$(git rev-parse HEAD)
docker run --rm -e COMMIT_ID="$REV" -e VERSION=0.8.3 -e GOFLAGS=-buildvcs=false -v "$PWD:/src" -v /home/ubuntu/RAG_Weknora/.go-modcache:/go/pkg/mod -v weknora-go-build-cache:/root/.cache/go-build -w /src weknora-builder:upstream-20261020 sh -ec 'export GO_VERSION=$(go version | cut -d" " -f3); make build-prod GO_BUILD_TAGS=anydoc BINARY_NAME=runtime/inline-citations/WeKnora'
docker tag sha256:94f2d38aa0663ee9441157715007f72f5514a40d2b56ae06318b8908b2e366e8 weknora-app:ljc-agent-citations-baseline-20261022
test "$(docker image inspect --format '{{.Id}}' weknora-app:ljc-agent-citations-baseline-20261022)" = sha256:94f2d38aa0663ee9441157715007f72f5514a40d2b56ae06318b8908b2e366e8
docker build -f deploy/inline-citations/Dockerfile.app --build-arg BASE_IMAGE=weknora-app:ljc-agent-citations-baseline-20261022 --build-arg SOURCE_REVISION="$REV" -t "weknora-app:ljc-agent-citations-$REV" .
```

## Release and acceptance

Source is `eccda4115935f39e75fa527f2af6b1a6267792d0`. Tested, built and deployed image is `sha256:391885b29deae41638a817f831f0da506094022a51e7f26324b34b43125314e1` (`weknora-app:ljc-agent-citations-eccda4115`). The three race-tested packages and Agent/chat-pipeline regressions passed. The user granted this feature's restart authorization on2026-10-22; subsequent live acceptance is recorded below. Preparation initially stopped before deployment pending that authorization.

The app-only release helper has a read-only preflight and an explicit authorization gate:

```bash
cd /home/ubuntu/RAG_Weknora-upstream-20261020
python3 scripts/ljc_release_agent_citations.py --image sha256:391885b29deae41638a817f831f0da506094022a51e7f26324b34b43125314e1
# Only after the user approves this feature's restart window:
python3 scripts/ljc_release_agent_citations.py --image sha256:391885b29deae41638a817f831f0da506094022a51e7f26324b34b43125314e1 --apply --restart-authorized
```

It refuses a changed/already-promoted baseline, validates that only app image differs, saves a private before-compose, restarts app only, reloads existing custom nginx for new backend DNS, checks health/HTTP, verifies every other container ID/image unchanged, and rolls the app image back if promotion checks fail. Do not replay on an already promoted release.

Record the image ID/source revision, resolved production compose, current task counts and protected container identities first. Obtain a new backend restart window. Change only `services.app.image` in `/home/ubuntu/weknora-fork-migration/20261021-production/compose.production.json`, preserving every other resolved field. Keep a private rollback copy outside Git. Run `docker compose -f <production-compose> up -d --no-deps app`, check app health and nginx connectivity on18182; do not recreate database/docreader/model/18180/frontend containers.

Use separate acceptance conversations, not edits to existing user answers:

1. Intelligent mode: repeat “数字档案室建设与管理工作如何纳入相关人员岗位职责？” (original request `DW3uQVYnGdlo`) and “药品经营许可证”的定义与经营范围 (original request `nCq9bNvazSKE`).
2. Add single-document, multiple-document and paginated-read questions; inspect logs for the actual tool paths. Verify Wiki/graph evidence remains derived, not mislabeled direct PDF evidence.
3. Exercise compaction in an independent long conversation without changing production model windows; verify actual compaction logs, stable handles and a valid subsequent inline citation.
4. Compare quick QA and unrelated chat; no extra citation-only retrieval or repair/reanswer requests. Inspect nearby fact-to-chunk support, not merely a nonzero citation count.
5. On18182 click inline footnotes, copy full chunk, open the pinned original and source directory. Missing-binding notices remain hidden; genuine permission/version errors remain visible.

Do not report live acceptance from unit tests/build/health alone. If the live model still omits references, use the decode counters to diagnose; do not fabricate or fill citations after generation.

## Rollback

Inspect current image/config and the private saved before-compose. Restore only app image `sha256:94f2d38aa0663ee9441157715007f72f5514a40d2b56ae06318b8908b2e366e8` in the active compose, then:

```bash
docker compose -f /home/ubuntu/weknora-fork-migration/20261021-production/compose.production.json up -d --no-deps app
docker exec WeKnora-frontend-custom-preview nginx -t
docker exec WeKnora-frontend-custom-preview nginx -s reload
```

Reload nginx only to refresh app DNS after recreation. Keep new evidence records; do not roll back with a destructive migration or restore an old database.

## Live acceptance and release record

- App-only guarded promotion succeeded. All87 other pre-existing containers retained their IDs/images, including18182 frontend,18180, databases and model services. App health is healthy. The custom frontend remains image `sha256:8242e2acfe8fb4b3720da0b06630150d890568801345900bf2b29673c44847b6`, source `872fe87b0`.
- Private before-compose/result: `/home/ubuntu/weknora-fork-migration/20261021-production/agent-citations-391885b29dea/`. The apply command above is historical and intentionally refuses replay after deployment.
- Frontend citation regressions:16 passed and typecheck passed; no frontend rebuild/restart. Backend tests include race checks on modelcontext/compaction/archivalcitations and Agent/chat-pipeline regressions.

| Mode/sample | Request | Session | Evidence binding |
| --- | --- | --- | --- |
| Intelligent: original岗位职责 question, no added request for citations | `rLIZ95I0VwtW` | `e62a62b9-3c52-4529-b1c9-882749ba091b` |5 raw refs,5 expanded,5 unique saved chunks across4 source versions;12 eligible evidence chunks; no rejects/capture failures; validation1ms/save13ms |
| Intelligent: original“介绍一下药品经营许可证”, including real long-document read/compaction | `YgUibXsVy64Y` | `c6d6e8ea-decf-46d1-b38b-09ac252af303` |19 raw refs,19 expanded,15 unique saved chunks;49 eligible evidence chunks; no rejects/capture failures; validation3ms/save12ms |
| Quick QA: GB/T20988 operation/maintenance requirements | `dHJwULJci3hg` | `46aed0a1-af41-463c-8960-e37edbc560f2` |4 saved chunks, inline citations visible |
| Intelligent: unrelated greeting, no retrieval requested | `p0l3rKiuIjkc` | `df545b78-61e5-4275-9ac5-cafa42bab040` |0 tool calls,0 citations, state not_applicable; no fabricated source/footer |

All samples have `repair_attempted=false`; intelligent finalization logs report `repair_calls=0`. The second request actually ran the existing context-compaction model call at14:51:58 before its final cited answer at14:53:27. This is not a citation-repair call, and its latency was not eliminated by this feature. Single/multiple-document results and full-document reading are covered by these actual tool paths; Wiki/graph provenance is additionally covered by automated regressions, not claimed as separately manually generated Wiki/graph answers.

Real18182 browser verified numbered inline footnotes after completion and reload; versioned chunk drawer; copied-original success toast; actual19-page original PDF rendering with honest no-anchor homepage fallback; source directory listing4 versions/5 chunks. Local screenshot artifacts stay outside Git. The original source really contains the cited岗位职责/年度考核 row; copying the pinned snapshot does not mix in model explanations.

Limits: these successful samples do not establish100% fact coverage or a controlled speed benchmark. The backend still recorded10 and48 reviewable blocks respectively; they are not automatically assigned nearby references. Finished/reloaded answers hide their notices as before. During typewriter playback a partial server-added notice span was briefly visible as text before the settled rendering; this frontend behavior was observed but the frontend was deliberately kept unchanged in this authorized backend-only release.
