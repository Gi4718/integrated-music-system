# Multi-stage build - Integrated Music System

# Stage 1: Build Go binaries
FROM golang:1.21-alpine AS go-builder

ENV GOPROXY=https://goproxy.cn,https://goproxy.io,direct
ENV GOSUMDB=off
ENV GOPATH=/tmp/gopath

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk add --no-cache git && \
    git config --global url."https://ghfast.top/https://github.com/".insteadOf "https://github.com/"

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o endfield-music-server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o endfield-music-cli ./cmd/cli

# Stage 2: Build Vue frontend
FROM node:20-alpine AS vue-builder

ENV NPM_CONFIG_REGISTRY=https://registry.npmmirror.com

WORKDIR /build
COPY web/package.json web/package-lock.json* ./
RUN npm install
COPY web/ .
RUN npm run build

# Stage 3: Runtime image
FROM alpine:3.19

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk add --no-cache \
        ca-certificates \
        nodejs \
        npm \
        supervisor \
        sqlite \
    && update-ca-certificates

RUN mkdir -p /data/config /data/ssl /data/downloads /data/db /app

COPY --from=go-builder /build/endfield-music-server /app/endfield-music-server
COPY --from=go-builder /build/endfield-music-cli /app/endfield-music-cli
COPY --from=vue-builder /build/dist /app/web/dist

ENV NPM_CONFIG_REGISTRY=https://registry.npmmirror.com
RUN npm install -g NeteaseCloudMusicApi && \
    mkdir -p /app/netease-api && \
    cp -r /usr/local/lib/node_modules/NeteaseCloudMusicApi/* /app/netease-api/

COPY docker/supervisord.conf /etc/supervisord.conf

WORKDIR /app
EXPOSE 33550 33551
CMD ["/usr/bin/supervisord", "-c", "/etc/supervisord.conf"]
