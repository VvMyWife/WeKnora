FROM docker.m.daocloud.io/library/golang:1.26-bookworm@sha256:ee3901c61e423732605a51f1b8157c849ec01194ea0a274213b5202b2c11e7bf
ENV GOPROXY=https://goproxy.cn GOSUMDB=sum.golang.google.cn \
    PIP_INDEX_URL=https://pypi.tuna.tsinghua.edu.cn/simple \
    UV_DEFAULT_INDEX=https://pypi.tuna.tsinghua.edu.cn/simple \
    NPM_CONFIG_REGISTRY=https://registry.npmmirror.com
RUN find /etc/apt -type f \( -name '*.sources' -o -name '*.list' \) -exec sed -i 's@http://deb.debian.org/debian@http://mirrors.tuna.tsinghua.edu.cn/debian@g;s@http://security.debian.org/debian-security@http://mirrors.tuna.tsinghua.edu.cn/debian-security@g' {} + \
 && apt-get update && apt-get install -y --no-install-recommends python3-venv python3-pip libsqlite3-dev libgomp1 \
 && pip3 install --break-system-packages uv \
 && rm -rf /var/lib/apt/lists/*
WORKDIR /src
