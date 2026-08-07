#!/bin/bash
# 快速迭代前端：在 coze-fe-debug 容器里构建 app，把 dist 同步进运行中的
# coze-web (nginx) 静态目录。无需重新构建 Docker 镜像。
# 构建约 1.5~3 分钟，同步后浏览器强制刷新（Ctrl+Shift+R）即可看到效果。
#
# 用法：bash scripts/dev-sync-frontend.sh
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
APP_DIR="$REPO_ROOT/frontend/apps/coze-studio"
PACKAGES_DIR="$REPO_ROOT/frontend/packages"
WEB_CONTAINER="${WEB_CONTAINER:-coze-web}"
DEBUG_IMAGE="${DEBUG_IMAGE:-coze-fe-debug:latest}"
CONTAINER="fe-sync"

cleanup() {
  docker rm -f "$CONTAINER" >/dev/null 2>&1 || true
}
trap cleanup EXIT

echo "==> 启动构建容器 ..."
docker rm -f "$CONTAINER" >/dev/null 2>&1 || true
docker run -d --name "$CONTAINER" "$DEBUG_IMAGE" sleep 3600 >/dev/null

echo "==> 同步最新源码（排除 node_modules/dist）..."
tar -C "$APP_DIR" \
  --exclude=node_modules --exclude=dist --exclude=coverage \
  -cf - . | docker exec -i "$CONTAINER" tar -C /app/frontend/apps/coze-studio -xf -
tar -C "$PACKAGES_DIR" \
  --exclude=node_modules --exclude=dist --exclude=__tests__ --exclude=coverage \
  -cf - . | docker exec -i "$CONTAINER" tar -C /app/frontend/packages -xf -

echo "==> 构建 app ..."
docker exec "$CONTAINER" bash -c \
  'cd /app/frontend/apps/coze-studio && NO_COLOR=1 IS_OPEN_SOURCE=true npm run build' \
  2>&1 | tail -4

echo "==> 同步 dist 到 coze-web nginx ..."
docker exec "$CONTAINER" tar -C /app/frontend/apps/coze-studio/dist -cf - . \
  | docker exec -i "$WEB_CONTAINER" tar -C /usr/share/nginx/html -xf -

echo "==> 完成！请浏览器 Ctrl+Shift+R 强刷查看。"
