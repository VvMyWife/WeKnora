ARG RUNTIME_BASE
FROM ${RUNTIME_BASE}
ARG SOURCE_REVISION
LABEL org.opencontainers.image.revision=${SOURCE_REVISION}
COPY WeKnora /app/WeKnora
COPY migrations/ /app/migrations/
RUN chmod 0755 /app/WeKnora \
 && find /app/migrations -type d -exec chmod 0755 {} + \
 && find /app/migrations -type f -exec chmod 0644 {} +
