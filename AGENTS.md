# Repository maintenance
- Authoritative server checkout: /home/ubuntu/RAG_Weknora.
- Public fork: VvMyWife/WeKnora; upstream: Tencent/WeKnora.
- New branches use ljc/; new commits use author and committer name ljc.
- Preserve existing history. Commit each tested implementation stage separately.
- Integrate the user-approved fixed upstream/main target in full. frontend/ follows the official target; frontend-custom/ may now adapt all six modules as approved on 2026-10-21, retaining custom brand, layout, features and original preview. Preserve the historical freeze manifest as a baseline; pin reviewed candidate changes separately.
- Prior checkbox selection is superseded by full-sync approval. Resolve technical conflicts and retain custom behavior; ask only for incompatible business choices and custom frontend module design.
- Keep secrets, user data, databases, runtime logs, models, backups and build outputs outside Git.
- Use persistent China mirror configuration for dependency downloads and builds.
- After interruption, inspect Git/process/database state before continuing. Never blindly replay migrations or deployments.
- Document complete manual test, build, integration, deployment and rollback commands in README.

## Frontend entry points and release acceptance
- Default daily-use and acceptance URL: http://10.20.0.15:18182 (custom frontend).
- Keep http://10.20.0.15:18180 as the preserved legacy frontend for explicit comparison and troubleshooting only. Do not direct normal users there to verify new features.
- For user-requested frontend changes, deliver to the custom frontend serving port 18182; do not overwrite the preserved frontend on port 18180 or import upstream UI changes without explicit authorization. This convention does not override the existing UI freeze or authorize new UI changes by itself.
- These are separate frontend entry points to shared backend data, not independent knowledge bases. Neither entry point isolates uploads, deletions, or document-processing changes from the other.
- A Git commit or successful build is not a deployed feature. Before reporting a change as live, verify the actual image/assets served on port 18182 and exercise the relevant page there. Preserve the existing runtime configuration and rollback baseline, and obtain the required restart authorization.
- If a user cannot see a new feature, first check the accessed port, the live deployed version, and stale browser assets. Do not assume that refreshing alone will resolve it.

## Approved integration candidate exception (2026-10-20)
- User explicitly approved mapped-original type/PDF preview support in candidate28180. Exact file hashes are pinned in docs/upstream-integration/policy.json. This does not authorize production deployment or other UI module redesign.

- Production entry-point notes above describe the preserved running baseline. Current full-sync acceptance uses official28180 and custom28182 with isolated backend28081; never treat a candidate change as a production deployment.

## Approved custom candidate integration (2026-10-21)
- User approved all six frontend module adaptations following the presented recommendation. Technical compatibility fixes and locale corrections are authorized. Keep custom brand/layout and existing original preview. Validate in28182 only; production18182 remains unchanged.
- User requires GPU reranking; candidate GPU service uses cached BGE weights and an explicitly selected GPU with sufficient free memory. Do not stop or move other GPU services.


## Production release authorization and canonical frontend (2026-10-21)
- User explicitly authorized full production database/backend upgrade with backups and task checks, followed by18182 browser acceptance and stopping28182.
- 18182 is the only canonical daily frontend. Do not direct users to28182; do not recreate a28182 host binding after retirement.
- 18180 is retained only for legacy troubleshooting; it is not a daily or release acceptance endpoint.
- Production commands and recovery: docs/production-release-20261021/README.md. Prior candidate-only boundaries describe the pre-release period.

## Current full upstream synchronization (2026-10-23)
- Full official target bc1ade36 is approved for both frontend trees, preserving custom capabilities. Earlier UI selection and candidate port notes are historical.
- Current maintenance worktree: /home/ubuntu/RAG_Weknora-upstream-20261023; evidence: docs/upstream-sync-20261023/README.md.
- Preserve 20261020 runtime files because production mounts config; do not checkout new code there.
- This batch covers code and isolated candidates. Production128-to133 migration and switching need a new maintenance window. Keep28182 retired.

## User-approved GitHub delivery workflow (2026-10-26)
- Keep ljc/custom as the latest verified delivery branch. Develop major changes in isolated ljc/ feature branches and worktrees.
- After relevant full regression and required isolated runtime acceptance pass, merge the feature into ljc/custom and explicitly push origin ljc/custom:ljc/custom to GitHub. Preserve both histories; never force push or overwrite another branch.
- This user-approved delivery flow authorizes the completed feature merge/push; it does not authorize production migrations or restarts. Record deployed image revisions separately.
- Do not configure GitLab or an automatic push hook/cron. Preserve push.default=nothing and explicit branch selection.
- Detailed manual commands and interruption checks: docs/custom-delivery/README.md.
