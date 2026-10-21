ARG BASE_IMAGE=weknora-app:ljc-inline-baseline-20261021
FROM ${BASE_IMAGE}
ARG SOURCE_REVISION
LABEL org.opencontainers.image.revision=${SOURCE_REVISION} ljc.operation=inline-citations-20261021
# Preserve every existing runtime customization, native library and migration.
COPY runtime/inline-citations/WeKnora /app/WeKnora
