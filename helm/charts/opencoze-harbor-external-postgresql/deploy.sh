#!/bin/bash
# Coze Studio K8s 一键部署脚本 (外部 PostgreSQL)
# 使用方法: bash deploy.sh [namespace]

set -e

NAMESPACE=${1:-coze}
CHART_DIR="$(dirname "$0")"

echo "=========================================="
echo "  Coze Studio K8s 部署 (外部 PostgreSQL)"
echo "=========================================="
echo ""

# PostgreSQL 配置
PG_HOST="postgres.ai-middleware.svc"
PG_PORT="5432"
PG_SUPER_USER="navigator"
PG_SUPER_PASSWORD="ai-platform-pg-pwd-2026"
PG_USER="opencoze_user"
PG_PASSWORD="root"
PG_DATABASE="opencoze"

# Harbor 配置
HARBOR_SERVER="harbor.ubidirector.cn"
HARBOR_USER="admin"
HARBOR_PASSWORD="admin@123"

# 检查工具
echo "🔍 检查环境..."
if ! command -v kubectl &> /dev/null; then
    echo "❌ kubectl 未安装"
    exit 1
fi

if ! command -v helm &> /dev/null; then
    echo "❌ helm 未安装"
    exit 1
fi

if ! command -v psql &> /dev/null; then
    echo "⚠️  psql 未安装，将跳过数据库初始化（需要手动执行）"
    PSQL_AVAILABLE=false
else
    PSQL_AVAILABLE=true
fi

# 检查 K8s 连接
if ! kubectl cluster-info &> /dev/null; then
    echo "❌ 无法连接 K8s 集群"
    exit 1
fi

echo "✅ 环境检查通过"
echo ""

# 步骤1: 创建数据库和用户
echo "📦 步骤1/5: 创建数据库和用户"
echo "----------------------------------------"
echo "请在 PostgreSQL 上执行以下 SQL:"
echo ""
echo "psql -U $PG_SUPER_USER -h $PG_HOST"
echo ""
echo "CREATE DATABASE $PG_DATABASE;"
echo "CREATE USER $PG_USER WITH PASSWORD '$PG_PASSWORD';"
echo "GRANT ALL PRIVILEGES ON DATABASE $PG_DATABASE TO $PG_USER;"
echo "\\c $PG_DATABASE"
echo "GRANT ALL ON SCHEMA public TO $PG_USER;"
echo "ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO $PG_USER;"
echo "ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO $PG_USER;"
echo ""
read -p "数据库创建完成后按 Enter 继续..."
echo ""

# 步骤2: 初始化 Schema
echo "📦 步骤2/5: 初始化数据库 Schema"
echo "----------------------------------------"
if [ "$PSQL_AVAILABLE" = true ]; then
    echo "正在执行 Schema 初始化..."
    psql -U $PG_USER -h $PG_HOST -d $PG_DATABASE -f "$CHART_DIR/files/postgres/schema.sql"
    echo "✅ Schema 初始化完成"
else
    echo "请手动执行以下命令:"
    echo "psql -U $PG_USER -h $PG_HOST -d $PG_DATABASE -f $CHART_DIR/files/postgres/schema.sql"
    read -p "Schema 初始化完成后按 Enter 继续..."
fi
echo ""

# 步骤3: 创建命名空间和 Secret
echo "📦 步骤3/5: 创建命名空间和 Secret"
echo "----------------------------------------"
echo "创建命名空间: $NAMESPACE"
kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -

echo "创建 Harbor 认证..."
kubectl create secret docker-registry harbor-secret \
  --docker-server=$HARBOR_SERVER \
  --docker-username=$HARBOR_USER \
  --docker-password=$HARBOR_PASSWORD \
  -n $NAMESPACE \
  --dry-run=client -o yaml | kubectl apply -f -

echo "✅ 命名空间和 Secret 创建完成"
echo ""

# 步骤4: 检查 StorageClass
echo "📦 步骤4/5: 检查 StorageClass"
echo "----------------------------------------"
echo "可用的 StorageClass:"
kubectl get sc
echo ""
echo "当前配置使用: coze-storage"
echo "如果需要修改，请编辑 values.yaml 中的 storageClassName"
read -p "确认 StorageClass 正确后按 Enter 继续..."
echo ""

# 步骤5: 安装 Helm Chart
echo "📦 步骤5/5: 安装 Helm Chart"
echo "----------------------------------------"
echo "正在安装 Coze Studio..."
helm install opencoze $CHART_DIR \
  --namespace $NAMESPACE \
  -f $CHART_DIR/values.yaml \
  --wait --timeout 10m

echo ""
echo "=========================================="
echo "  部署完成！"
echo "=========================================="
echo ""

# 验证部署
echo "📊 部署状态:"
echo ""
echo "Pod 状态:"
kubectl get pods -n $NAMESPACE
echo ""
echo "Service 状态:"
kubectl get svc -n $NAMESPACE
echo ""

# 获取访问地址
echo "🌐 访问地址:"
EXTERNAL_IP=$(kubectl get svc -n $NAMESPACE opencoze-web -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null)
if [ -z "$EXTERNAL_IP" ]; then
    echo "LoadBalancer IP 分配中，请稍后执行:"
    echo "  kubectl get svc -n $NAMESPACE opencoze-web"
    echo ""
    echo "或者使用 NodePort 方式访问:"
    NODE_PORT=$(kubectl get svc -n $NAMESPACE opencoze-web -o jsonpath='{.spec.ports[0].nodePort}' 2>/dev/null)
    if [ -n "$NODE_PORT" ]; then
        echo "  http://<任意节点IP>:$NODE_PORT"
    fi
else
    echo "  http://$EXTERNAL_IP"
fi
echo ""

# 常用命令
echo "📝 常用命令:"
echo ""
echo "查看日志:"
echo "  kubectl logs -f deployment/opencoze-server -n $NAMESPACE"
echo "  kubectl logs -f deployment/opencoze-web -n $NAMESPACE"
echo ""
echo "升级配置:"
echo "  helm upgrade opencoze $CHART_DIR -n $NAMESPACE -f $CHART_DIR/values.yaml"
echo ""
echo "卸载:"
echo "  helm uninstall opencoze -n $NAMESPACE"
echo "  kubectl delete namespace $NAMESPACE"
echo ""
