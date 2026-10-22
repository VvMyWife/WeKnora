# Inline citations: generation, local validation and grouped display

## Baseline and delivery

- Active source: `/home/ubuntu/RAG_Weknora-upstream-20261020`, branch `ljc/custom`.
- Baseline HEAD: `60453bd8d`. Author and committer for new stages: `ljc <ljc@local>`.
- Backend baseline image: `sha256:76399c0e703a7118ec3d7b5f7cd5140388f0c147bad32c2df6b154b1f92c51e1`.
- Custom UI baseline image: `sha256:66b71bd73accc993f5c641dbeab4bc2bb8c999eae1f6050428c1fc147892043f`.
- Daily acceptance URL: http://10.20.0.15:18182. Do not change original `frontend/`, legacy18180, database migrations or document-processing queues.
- Runtime source/backup manifest: `docs/production-release-20261021/README.md`. The old `/home/ubuntu/RAG_Weknora` worktree is not the current release source.

## Why this change

Two real answers retrieved the supporting passage, but persisted respectively 2 citations/3 unbound lines and 0 citations/17 unbound lines. The latter incurred a separate approximately65-second, 2048-output-token model repair. Current logs cannot distinguish repair truncation from malformed/no binding output. Do not commit private answers or evidence text as test fixtures.

New behavior: emit short refs during answer generation, validate locally, never call a second citation model, and only consolidate presentation of already-bound sources. A valid source binding does not establish factual correctness. Missing refs are never guessed from adjacent citations, document names or text similarity.

## Stage ledger

0. Baseline and identity (this document).
1. Generation protocol and context-budget checks.
2. Local block bindings and diagnostics; remove repair calls.
3. Custom UI grouping, scope hints and review reasons.
4. Integrated acceptance and release instructions.

Each stage requires tests, explicit staged-file review, an independent commit and recorded hash. No history rewriting. Release/restart requires a new authorized window.

## Reproduce checks (server)

```bash
cd /home/ubuntu/RAG_Weknora-upstream-20261020
git status --short --branch
git var GIT_AUTHOR_IDENT
git var GIT_COMMITTER_IDENT
docker run --rm -v "$PWD:/src" -v /home/ubuntu/RAG_Weknora/.go-modcache:/go/pkg/mod -v weknora-go-build-cache:/root/.cache/go-build -w /src weknora-builder:upstream-20261020 go test ./internal/archivalcitations ./internal/modelcontext
```

Build/test/release instructions and measured results are completed in subsequent stages. Do not confuse a commit or successful build with a live18182 deployment.

## Stage results

- Stage0 baseline tests: `internal/archivalcitations` and `internal/modelcontext` passed; commit `c11b8f0bf`.
- Stage1 emits refs during factual sentence/list/table generation. Unknown context windows budget conservatively at32768 instead of claiming200K; declared model limits cap agent settings. This is a budgeting fallback, not provider capability detection. The current9B service was read-only verified as49152; its service configuration is unchanged. Explicitly declare model capacity in model settings for accurate sizing.
- Stage1 `f30712c49`: modelcontext/types/agent tests and4 frontend context-window tests passed.
- Stage2 removes the repair binder and all three QA/agent call sites. Finalization only reads registered evidence and writes snapshots/manifests. Missing refs stay unbound even if their literal text exists in evidence. Optional `bindings` and reason codes extend manifestv1 without a database migration. Paragraph soft breaks, list items and table rows are separate coverage units; an uncited factual tail remains reviewable. A local binding does not verify semantics inside the source.
- Stage2 `34d3d187a`: archivalcitations/modelcontext/agent/chat_pipeline passed. The first compile exposed two imports used only by the removed binder; fixed and all four packages rerun successfully.
- Stage3 groups verified archival refs only, preserving distinct chunk/version IDs. Paragraphs never merge across boundaries; same-level consecutive list items with identical source sets can share their final refs. Missing refs, review markers, code, nested lists, tables and source changes break groups. During streaming the last top-level branch stays untouched. The canonical answer/manifest and source directory are unchanged. Set build-time `VITE_CITATION_GROUPING_ENABLED=false` to disable only this projection; no repair call is re-enabled.
- Stage3 `6d0d616f3`:16 focused UI regressions and full Vue/TypeScript checking passed.

## Build, inspect and recover from interruption

The build-only script inherits the exact current production image, preserving native libraries, migrations, proxy configuration and other customizations. It refuses a changed baseline and uses new image tags. It uses existing persistent China-configured builders/dependency caches, not temporary dependency installs.

```bash
cd /home/ubuntu/RAG_Weknora-upstream-20261020
bash -n scripts/ljc_build_inline_citations.sh
git diff --check
bash scripts/ljc_build_inline_citations.sh
# In another terminal, inspect progress; do not start another build:
docker ps --filter name=ljc-inline --format '{{.Names}} {{.Status}}'
tail -30 runtime/inline-citations/$(git rev-parse HEAD)/app.log
tail -30 runtime/inline-citations/$(git rev-parse HEAD)/ui.log
cat runtime/inline-citations/$(git rev-parse HEAD)/build.status
```

The script holds a flock and reports independent app/UI exit codes. After interruption, check processes, logs and status first. Never automatically replay a production restart or migration. Completed build does not deploy.

## Manual acceptance

1. Open18182, sign in, and use an independent sample conversation with a mapped PDF. Ask a narrowly scoped question whose answer is present in supplied evidence.
2. Verify factual prose/list/table refs are inline; same-source paragraph/list refs consolidate only in presentation, with a hover/focus scope hint. Change to another chunk or version and verify it stays separate.
3. Click and keyboard-activate the remaining footnotes. Verify exact snapshot, copy actions, known-page PDF positioning, continuous scrolling, and honest no-page fallback. Verify the source directory contains only actually cited chunks.
4. Verify missing-source content is not assigned nearby evidence. Open review reasons; pure allowlisted transitions/headings do not create spurious warnings.
5. Refresh and compare IDs, numbers, content and source directory. Old valid answers may consolidate visually but must not be rewritten or automatically rebound.
6. Inspect server logs: `[Citations] stage=finalize` includes validation/save milliseconds, registered/valid/invalid/review counts and `repair_calls=0`. Model usage after answer generation must not contain a citation-repair call. Compare generation time separately from local validation, with the same model, question and similar retrieved evidence.

## Deployment and rollback (authorization required)

This change requires no database migration. Keep the upgraded production database and current compose; do not invoke old migration94 tools or switch to the pre-upgrade application.

Before release, obtain a new restart window and check running processing jobs. Save a private copy of the active production compose and its resolved image IDs under the private release directory (never Git). Update **only** app/frontend image references to the two successful build IDs, keeping every existing environment, mount, network and runtime setting. Validate the compose before recreating only app/frontend. Do not restart docreader, model services or databases.

```bash
docker compose -f /home/ubuntu/weknora-fork-migration/20261021-production/compose.production.json config --quiet
# Only after recording the matching app/frontend image pair in that compose:
docker compose -f /home/ubuntu/weknora-fork-migration/20261021-production/compose.production.json up -d --no-deps app frontend
docker exec WeKnora-frontend-custom-preview nginx -t
docker exec WeKnora-frontend-custom-preview nginx -s reload
curl --fail http://10.20.0.15:18081/health
curl --fail -I http://10.20.0.15:18182/
docker inspect WeKnora-app WeKnora-frontend-custom-preview --format '{{.Name}} {{.Image}} {{index .Config.Labels "org.opencontainers.image.revision"}}'
docker logs --since 10m WeKnora-app 2>&1 | grep '\[Citations\]'
```

Keep18180 unchanged. If acceptance fails, restore only the saved app/frontend image pair (or the two baseline IDs above) and recreate those services with the same upgraded database/configuration. Reload nginx and recheck health. Do not reverse migrations, delete evidence, reset the branch or remove database/queue volumes. To disable only grouping, rebuild with `VITE_CITATION_GROUPING_ENABLED=false`; do not reintroduce automatic model repair.

Release state: user authorized immediate restart; the first release is live on18182. App/custom-UI source revision `97c809a99`, release tooling `f56529616`. Only app/frontend images changed; protected container IDs and upgraded database ledger were verified unchanged. Final acceptance follow-up is recorded below.

Build preparation exposed root-owned existing `runtime/`; the build helper now creates/grants ownership only to its new `runtime/inline-citations` directory using the existing builder. It does not recursively change ownership of the source tree or any production data. The failed first attempt did not start a build or touch running containers.

## Post-interruption acceptance (2026-10-22)

- Production selfcheck passed; source worktree clean at `f56529616`, both running images still match the approved first release.
- Full UI suite:1486 passed,0 failed; Go archivalcitations/modelcontext/types/agent/chat_pipeline passed. Typecheck/build passed.
- Browser: login, historical five-item list consolidation, evidence drawer, actual22-page PDF rendering and physical page2 navigation passed. The old source has no trustworthy physical anchors; the viewer correctly declares this and opens page1 before manual navigation. Historical answers are not rewritten.
- Quick-QA sample `06d60348-461b-4aea-9d38-6ebfe8b3c7e1` / `icxsM3cmhuux`: answer model9408 input/120 output tokens; local validation0ms/save22ms/repair_calls0. It did not retrieve the target table text into the visible four-chunk prompt and produced no valid refs; this is retained as a failed coverage case, not advertised as a successful citation answer.
- An intelligent-mode follow-up exposed early compaction at the unknown32K fallback. Confirmed real9B vLLM window49152, then updated **only** WeKnora's existing9B model `parameters.context_window` through its settings form to49152. Hash of all other parameter fields stayed `a40d623066253525e1fab471c2607475`; no service/model restart or other parameter change. Existing in-flight requests keep their old budget; new requests use the declared capacity.
- Follow-up source `51935395e`: short literal inline-ref generation example and six interpolated context-help locales. All1486 UI tests, Go regressions, typecheck and both image builds passed again. Release tooling `5612cd298` promoted only app/frontend images after checking no active document processing. Protected service IDs unchanged; no migration or model-service restart.

## Final release and measured acceptance

Current immutable images (source `51935395ed18967f60b2017536ef469d02734271`):

- App: `sha256:94f2d38aa0663ee9441157715007f72f5514a40d2b56ae06318b8908b2e366e8`.
- Custom UI18182: `sha256:79ab0a372e3858ac37dbabc10156e0e81b3a340e06cfde0107f56a200a3e0c79`.
- Private deployment result/rollback compose: `/home/ubuntu/weknora-fork-migration/20261021-production/inline-citations-51935395e/`. Do not commit this directory or blindly rerun release scripts after promotion.
- Public machine-readable release summary: `release.json`. Source revision is the binary/UI build revision; later documentation-only commits do not imply rebuilt application images.

Measured real-browser samples, not a controlled speed benchmark:

| Mode | Request | Result | Citation finalization |
| --- | --- | --- | --- |
| Intelligent | `SXZFQv4PuvEw` | About9seconds,1 search tool,3 canonical list bindings to one chunk,1 grouped final footnote; no context-compaction call in this sample | validation0ms,save59ms,repair_calls0 |
| Quick QA | `N2moRF73QXeu` |2 factual list items with2 distinct chunk footnotes; same PDF does not merge the chunks | validation0ms,save14ms,repair_calls0 |

Both samples retain1 reviewable introduction mentioning a standard but missing its own generated ref. This warning is intentional and is **not** silently covered by list citations. Do not claim100% sentence coverage, automatic truth verification or elimination of all model/retrieval latency. The earlier failed table-retrieval sample remains in the independent acceptance conversation for diagnosis; retrieval/chunk expansion ranking was not changed by this work.

On the deployed follow-up UI, verified mouse and Enter activation of inline footnotes, successful original-text copy toast, versioned source drawer, historical grouped display, reload persistence and correct48K label for the declared49152 window. The model editor now explains32768/32K fallback instead of200K. Original PDF22-page rendering/manual page navigation was visually checked; missing source-page anchors still explicitly fall back to page1. No private original content or screenshots are committed.

Final production selfcheck passed: upgraded ledger128+4 clean, app/frontend healthy, GPU reranker healthy, retired28182 still not listening,18180 preserved. **Release is live on18182; generation-time citation and presentation acceptance passed with the explicitly retained review warnings above.**

To reproduce the guarded promotion before a future release (do not rerun on an already promoted pair):

```bash
cd /home/ubuntu/RAG_Weknora-upstream-20261020
python3 -m py_compile scripts/ljc_release_inline_citations.py
python3 scripts/ljc_release_inline_citations.py
# Only with a newly authorized restart window and matching pinned baseline:
python3 scripts/ljc_release_inline_citations.py --apply --restart-authorized
python3 scripts/ljc_production_selfcheck.py
git log -8 --format='%h %an <%ae> | %cn <%ce> | %s'
```

For a later change, update/pin the source revision, current before-images and new tested image pair in the release helper/overlay in its own tested Git stage. It intentionally refuses unexpected images or replay of an attempted release.
