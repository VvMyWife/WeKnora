# WeKnora fork maintenance

Authority: /home/ubuntu/RAG_Weknora. Isolation: /home/ubuntu/RAG_Weknora-fork.

- main: upstream mirror only; never deploy it automatically.
- ljc/custom: custom maintenance branch.
- ljc/integrate-YYYYMMDD-topic: explicitly selected upstream changes.
- Existing original and custom frontend source are frozen.
- New commits use ljc; existing private history is not rewritten.
- Publish only after complete candidate-history review and validation.
- Database migration compatibility must be validated on an isolated restore.

The initial source has 47 custom commits after official v0.8.0; retain those existing functional commits rather than flattening them. Private backups stay outside the public repository.
