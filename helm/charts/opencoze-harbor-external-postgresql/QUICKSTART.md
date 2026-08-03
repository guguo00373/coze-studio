# Coze Studio K8s 部署 (内部 PostgreSQL)

## 快速开始

### 1. 修改配置

默认使用 Chart 内部 PostgreSQL StatefulSet，无需提前准备外部数据库。

确认 `values.yaml` 中配置：

```yaml
postgresql:
  enabled: true
  internal: true
  user: "opencoze"
  password: "opencoze123"
  database: "opencoze"

secret:
  stringData:
    DATABASE_PASSWORD: "opencoze123"
    MODEL_API_KEY_0: "sk-your-key"
    OPENAI_EMBEDDING_API_KEY: "sk-your-key"
    BUILTIN_CM_OPENAI_API_KEY: "sk-your-key"
```

如果服务器 StorageClass 不是 `coze-storage`，先修改 `values.yaml` 中所有 `storageClassName`。

### 2. 创建命名空间和 Harbor Secret

```bash
kubectl create namespace coze

kubectl create secret docker-registry harbor-secret \
  --docker-server=harbor.ubidirector.cn \
  --docker-username=admin \
  --docker-password='admin@123' \
  -n coze
```

### 3. 部署

```bash
helm install opencoze ./opencoze-harbor-external-postgresql \
  --namespace coze \
  -f values.yaml
```

### 4. 验证

```bash
kubectl get pods -n coze
kubectl get svc -n coze
```

内部 PostgreSQL 连接测试：

```bash
kubectl exec -it statefulset/opencoze-opencoze-pg-postgresql -n coze -- \
  psql -U opencoze -d opencoze -c "\dt"
```

如果 release 名或 chart fullname 不同，先用下面命令确认 PostgreSQL StatefulSet 名称：

```bash
kubectl get statefulset -n coze | grep postgresql
```

## 常用命令

```bash
# 查看日志
kubectl logs -f deployment/opencoze-opencoze-pg-server -n coze

# 升级
helm upgrade opencoze ./opencoze-harbor-external-postgresql -n coze -f values.yaml

# 卸载
helm uninstall opencoze -n coze
```

注意：卸载 Helm release 不会自动删除 PVC。需要彻底清理数据时再手动删除 PVC。
