FROM golang:1.26-bookworm
ENV GOPROXY=https://goproxy.cn GOSUMDB=off
RUN sed -i 's|http://deb.debian.org/debian-security|https://mirrors.tuna.tsinghua.edu.cn/debian-security|g; s|http://deb.debian.org/debian|https://mirrors.tuna.tsinghua.edu.cn/debian|g' /etc/apt/sources.list.d/debian.sources \
 && apt-get update && apt-get install -y --no-install-recommends libsqlite3-dev \
 && apt-get clean
