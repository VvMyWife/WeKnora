FROM weknora-builder:archival-citations
ENV GOPROXY=https://goproxy.cn \
    GOSUMDB=sum.golang.google.cn \
    PIP_INDEX_URL=https://pypi.tuna.tsinghua.edu.cn/simple \
    UV_DEFAULT_INDEX=https://pypi.tuna.tsinghua.edu.cn/simple \
    NPM_CONFIG_REGISTRY=https://registry.npmmirror.com
