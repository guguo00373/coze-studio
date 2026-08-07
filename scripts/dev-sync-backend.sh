#!/bin/bash
# 快速迭代后端：本地编译 Go 二进制，热替换进运行中的 coze-server 容器，再重启。
# 无需重新构建 Docker 镜像。首次运行会填充构建缓存（较慢），之后每次 ~1 分钟。
#
# 用法：bash scripts/dev-sync-backend.sh
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BACKEND_DIR="$REPO_ROOT/backend"
OUT_DIR="${SYNC_OUT_DIR:-/tmp/opencode}"
GOCACHE_VOL="${GOCACHE_VOL:-coze-gocache}"

echo "==> 编译后端二进制（使用持久化构建缓存）..."
docker run --rm \
  -v "$BACKEND_DIR":/app/backend \
  -v /root/go/pkg/mod:/go/pkg/mod \
  -v "$GOCACHE_VOL":/root/.cache/go-build \
  -v "$OUT_DIR":/out \
  -w /app/backend \
  -e GOPROXY=https://goproxy.cn,direct \
  -e GOSUMDB=off \
  golang:1.24-alpine sh -c 'apk add --no-cache git >/dev/null 2>&1; go build -o /out/opencoze main.go'

echo "==> 热替换二进制并重启 coze-server ..."
docker cp "$OUT_DIR/opencoze" coze-server:/app/opencoze
docker restart coze-server >/dev/null

echo "==> 等待服务就绪 ..."
for _ in $(seq 1 20); do
  if curl -sf -o /dev/null http://localhost:8889/health; then
    echo "==> coze-server 已就绪（health 200）"
    exit 0
  fi
  sleep 2
done
echo "!! 服务未就绪，请检查 docker logs coze-server"
exit 1
