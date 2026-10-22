# Hide missing-binding notices (2026-10-22)

Presentation only: hide current `引用待核验` and historical `未绑定来源` marker spans in chat, and remove the missing-binding footer/status. Keep actual footnotes, evidence drawers, source directory and genuine access/storage/version errors. Canonical answers and manifests remain unchanged. Group before removing notices, so unsupported content remains a consolidation boundary. No model request, backend build/restart, data migration or historical rewrite.

## Stage 1: source and tests

```bash
cd /home/ubuntu/RAG_Weknora-upstream-20261020
git status --short
git config user.name
git config user.email
docker run --rm -v "$PWD:/src" -v ljc-upstream-custom-node-modules:/src/frontend-custom/node_modules -w /src/frontend-custom weknora-frontend-builder:upstream-20261020 sh -ec 'npm test -- src/utils/citationGrouping.test.ts src/utils/archivalCitations.test.ts; npm run type-check'
python3 -m py_compile scripts/ljc_release_citation_notice_ui.py
# Inspect only the changed files, stage explicitly, commit with ljc identity.
git diff --check
git log -1 --format='%h %an <%ae> / %cn <%ce>'
```

## Stage 2: frontend build and guarded promotion

Existing persistent builder/dependency volume uses the configured China mirror; no new dependency is required. Record the source commit, never rebuild the backend for this UI patch.

```bash
cd /home/ubuntu/RAG_Weknora-upstream-20261020
REV=$(git rev-parse HEAD)
docker run --rm -e VITE_IS_DOCKER=true -e VITE_FRONTEND_COMMIT="$REV" -v "$PWD:/src" -v ljc-upstream-custom-node-modules:/src/frontend-custom/node_modules -w /src/frontend-custom weknora-frontend-builder:upstream-20261020 sh -ec 'find src -name "*.test.ts" -print0 | xargs -0 npm test -- --test-concurrency=2; npm run build'
docker build -f deploy/inline-citations/Dockerfile.frontend --build-arg BASE_IMAGE=sha256:79ab0a372e3858ac37dbabc10156e0e81b3a340e06cfde0107f56a200a3e0c79 --build-arg SOURCE_REVISION="$REV" -t "weknora-ui-custom:ljc-notice-$REV" .
IMAGE=$(docker image inspect --format '{{.Id}}' "weknora-ui-custom:ljc-notice-$REV")
python3 scripts/ljc_release_citation_notice_ui.py --image "$IMAGE"
python3 scripts/ljc_release_citation_notice_ui.py --image "$IMAGE" --apply
```

The release helper rejects replay, snapshots a private rollback compose, changes only `services.frontend.image`, recreates only frontend, verifies nginx/HTTP and checks every other container ID/image unchanged. No backend/model/18180 restart.

## Manual browser acceptance

1. Open http://10.20.0.15:18182 and log in; refresh any already open page.
2. Open existing conversation `b1992132-a82e-4b3f-8a77-492ed370d45b`: introduction and two fact items remain, no inline/footer pending-review notice; both numbered footnotes and the source directory remain.
3. Click each footnote: evidence drawer and copy/original-preview controls still work.
4. Open historical conversation `85386798-f6bd-47e8-8330-00bbe20d6a35`: old unbound notices are absent; valid links and conservative grouping remain. Refresh and repeat.
5. Open no-valid-citation sample `06d60348-461b-4aea-9d38-6ebfe8b3c7e1`: answer remains, no empty citation warning/footer; never manufacture a footnote.
6. Verify actual source access/deletion/version errors still show when evidence is unavailable. The new automated tests preserve these states; do not delete production sources to simulate them.

## Rollback

Inspect the recorded private `citation-notice-ui-<image-prefix>/result.json` and compare current image/config first. Do not replay a previous backend release script. Restore only the previous frontend image:

```bash
cd /home/ubuntu/RAG_Weknora-upstream-20261020
python3 - <<'PY'
import json
from pathlib import Path
p=Path('/home/ubuntu/weknora-fork-migration/20261021-production/compose.production.json')
c=json.loads(p.read_text())
c['services']['frontend']['image']='sha256:79ab0a372e3858ac37dbabc10156e0e81b3a340e06cfde0107f56a200a3e0c79'
p.write_text(json.dumps(c,indent=2)+'\n')
PY
docker compose -f /home/ubuntu/weknora-fork-migration/20261021-production/compose.production.json up -d --no-deps frontend
docker exec WeKnora-frontend-custom-preview nginx -t
```

Retain canonical evidence records. No destructive rollback migration.
