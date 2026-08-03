#!/bin/bash
# Coze Studio 一键部署脚本 (Windows 版本使用 Git Bash)
# 使用方法: bash deploy.sh [namespace]

set -e

NAMESPACE=${1:-coze}
CHART_DIR="$(dirname "$0")"

echo "=========================================="
echo "  Coze Studio K8s 一键部署"
echo "=========================================="
echo ""

# 检查 kubectl
if ! command -v kubectl &> /dev/null; then
    echo "❌ kubectl 未安装"
    exit 1
fi

# 检查 helm
if ! command -v helm &> /dev/null; then
    echo "❌ helm 未安装"
    exit 1
fi

# 检查 K8s 连接
if ! kubectl cluster-info &> /dev/null; then
    echo "❌ 无法连接 K8s 集群"
    exit 1
fi

echo "✅ 环境检查通过"
echo ""

# 步骤1: 创建命名空间
echo "📦 步骤1/4: 创建命名空间 [$NAMESPACE]"
kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -

# 步骤2: 创建 Harbor Secret
echo "🔐 步骤2/4: 创建 Harbor 仓库认证"
kubectl create secret docker-registry harbor-secret \
  --docker-server=harbor.ubidirector.cn \
  --docker-username=admin \
  --docker-password='admin@123' \
  -n $NAMESPACE \
  --dry-run=client -o yaml | kubectl apply -f -

# 步骤3: 安装 Helm Chart
echo "🚀 步骤3/4: 安装 Coze Studio"
helm install opencoze $CHART_DIR \
  --namespace $NAMESPACE \
  -f $CHART_DIR/values.yaml \
  --wait --timeout 10m

# 步骤4: 验证部署
echo "✅ 步骤4/4: 验证部署"
echo ""
echo "Pod 状态:"
kubectl get pods -n $NAMESPACE
echo ""
echo "Service 状态:"
kubectl get svc -n $NAMESPACE
echo ""

# 获取访问地址
echo "=========================================="
echo "  部署完成！"
echo "=========================================="
echo ""
echo "获取访问地址:"
EXTERNAL_IP=$(kubectl get svc -n $NAMESPACE coze-web -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null)
if [ -z "$EXTERNAL_IP" ]; then
    echo "LoadBalancer IP 分配中，请稍后执行: kubectl get svc -n $NAMESPACE coze-web"
    echo ""
    echo "或者使用 NodePort 方式访问:"
    NODE_PORT=$(kubectl get svc -n $NAMESPACE coze-web -o jsonpath='{.spec.ports[0].nodePort}')
    echo "  http://<任意节点IP>:${NODE_PORT}"
else
    echo "  http://${EXTERNAL_IP}"
fi
echo ""
echo "查看日志:"
echo "  kubectl logs -f deployment/coze-server -n $NAMESPACE"
echo "  kubectl logs -f deployment/coze-web -n $NAMESPACE"
echo ""
echo "卸载命令:"
echo "  helm uninstall opencoze -n $NAMESPACE"
echo "  kubectl delete namespace $NAMESPACE"
