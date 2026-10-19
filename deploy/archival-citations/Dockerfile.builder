FROM weknora-builder:enrichment-retry-20261016
# Persistent build defaults, not per-download ad-hoc overrides.
ENV GOPROXY=https://goproxy.cn,direct \
    GOSUMDB=sum.golang.google.cn
