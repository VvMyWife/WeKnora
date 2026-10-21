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
