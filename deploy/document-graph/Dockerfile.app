ARG RUNTIME_BASE
FROM ${RUNTIME_BASE}
ARG SOURCE_REVISION
LABEL org.opencontainers.image.revision=${SOURCE_REVISION}
COPY --chmod=0755 WeKnora /app/WeKnora
