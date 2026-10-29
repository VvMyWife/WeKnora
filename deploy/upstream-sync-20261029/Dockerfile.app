ARG BASE_IMAGE=weknora-app:ljc-sync-base-20261023
FROM ${BASE_IMAGE}
ARG SOURCE_REVISION
LABEL org.opencontainers.image.revision=${SOURCE_REVISION} ljc.operation=upstream-20261029
COPY runtime/upstream-20261029/WeKnora /app/WeKnora
COPY runtime/upstream-20261029/ljc-migrate /app/ljc-migrate
COPY migrations /app/migrations-next-sync
RUN mv /app/migrations /app/migrations-before-sync-20261023 && mv /app/migrations-next-sync /app/migrations
COPY config /app/config
COPY THIRD_PARTY_NOTICES.md /app/THIRD_PARTY_NOTICES.md
