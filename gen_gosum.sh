#!/bin/sh
cd /vol1/1000/docker/netmusic-downloader
docker run --rm \
  -v /vol1/1000/docker/netmusic-downloader:/app \
  -w /app \
  -e GOPROXY=https://goproxy.cn,direct \
  -e GOSUMDB=off \
  docker.1ms.run/library/golang:1.21-alpine \
  sh -c 'apk add --no-cache git && git config --global url."https://ghfast.top/https://github.com/".insteadOf "https://github.com/" && go mod tidy 2>&1'
