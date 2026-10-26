FROM weknora-app:archive-base-20261026
ARG SOURCE_REVISION
LABEL org.opencontainers.image.revision=${SOURCE_REVISION} ljc.operation=archive-document-20261026
COPY runtime/archive-document/WeKnora /app/WeKnora
