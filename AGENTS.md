# Repository maintenance
- Authoritative server checkout: /home/ubuntu/RAG_Weknora.
- Public fork: VvMyWife/WeKnora; upstream: Tencent/WeKnora.
- New branches use ljc/; new commits use author and committer name ljc.
- Preserve existing history. Commit each tested implementation stage separately.
- Integrate the user-approved fixed upstream/main target in full. frontend/ follows the official target; frontend-custom/ stays frozen until the user confirms a module adaptation.
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
