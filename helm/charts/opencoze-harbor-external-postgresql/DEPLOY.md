# Coze Studio K8s 部署完整指南 (内部 PostgreSQL)

## 一、环境准备

前置条件：

- K8s 集群 v1.20+
- Helm 3.x
- kubectl 已配置
- Harbor 仓库可访问
- 集群存在可用 StorageClass

本 Chart 默认部署内部 PostgreSQL，不需要提前准备外部数据库。

## 二、内部 PostgreSQL

默认配置在 `values.yaml`：

```yaml
postgresql:
  enabled: true
  internal: true
  image:
    repository: harbor.ubidirector.cn/coze/postgres
    tag: "15"
  host: ""
  port: "5432"
  user: "opencoze"
  password: "opencoze123"
  database: "opencoze"
  initJob:
    enabled: false
```

内部 PostgreSQL 会通过 StatefulSet 部署，Service 名称为：

```text
<release-fullname>-postgresql
```

数据库 Schema 通过 `files/postgres/schema.sql` 挂载到：

```text
/docker-entrypoint-initdb.d/init.sql
```

PostgreSQL 首次初始化数据目录时会自动执行该脚本。

## 三、部署前配置

### 1. 修改 StorageClass

查看集群可用 StorageClass：

```bash
kubectl get sc
```

如果不是 `coze-storage`，修改 `values.yaml` 中所有 `storageClassName`。

### 2. 修改敏感信息

修改 `values.yaml`：

```yaml
secret:
  stringData:
    DATABASE_PASSWORD: "your_postgresql_password"
    MINIO_ROOT_PASSWORD: "your_minio_password"
    MODEL_API_KEY_0: "sk-your-key"
    OPENAI_EMBEDDING_API_KEY: "sk-your-key"
    BUILTIN_CM_OPENAI_API_KEY: "sk-your-key"
```

如果修改了 `DATABASE_PASSWORD`，也同步修改：

```yaml
postgresql:
  password: "your_postgresql_password"
```

### 3. 创建 Harbor Secret

```bash
kubectl create namespace coze

kubectl create secret docker-registry harbor-secret \
  --docker-server=harbor.ubidirector.cn \
  --docker-username=<username> \
  --docker-password='<password>' \
  -n coze
```

## 四、安装

```bash
cd /root/guguo/opencoze-harbor-external-postgresql

helm install opencoze . \
  --namespace coze \
  -f values.yaml
```

部署前可先渲染检查：

```bash
helm template opencoze . -n coze -f values.yaml
```

## 五、验证

查看 Pod 和 Service：

```bash
kubectl get pods -n coze
kubectl get svc -n coze
kubectl get statefulset -n coze
```

验证 PostgreSQL Schema：

```bash
kubectl get statefulset -n coze | grep postgresql
kubectl exec -it statefulset/opencoze-opencoze-pg-postgresql -n coze -- \
  psql -U opencoze -d opencoze -c "\dt"
```

如果 release 名称不同，先用 `kubectl get statefulset -n coze | grep postgresql` 确认实际名称。

## 六、访问系统

查看前端 Service：

```bash
kubectl get svc -n coze | grep web
```

如果是 `LoadBalancer`，访问 `EXTERNAL-IP`。如果是 `NodePort`，访问任意节点 IP 加 NodePort。

## 七、常用命令

```bash
# 后端日志
kubectl logs -f deployment/opencoze-opencoze-pg-server -n coze

# 前端日志
kubectl logs -f deployment/opencoze-opencoze-pg-web -n coze

# PostgreSQL 日志
kubectl logs -f statefulset/opencoze-opencoze-pg-postgresql -n coze

# 升级
helm upgrade opencoze . -n coze -f values.yaml

# 卸载 Helm release
helm uninstall opencoze -n coze
```

注意：卸载 Helm release 不会自动删除 PVC。需要彻底清理数据时再手动删除 PVC。

## 八、故障排查

### Pod Pending

```bash
kubectl describe pod <pod-name> -n coze
```

常见原因：StorageClass 不存在、资源不足、PVC 无法绑定。

### 镜像拉取失败

```bash
kubectl get events -n coze --sort-by='.lastTimestamp'
```

检查 `harbor-secret`、Harbor 地址和镜像 tag 是否存在。

### 数据库连接失败

```bash
kubectl exec -it statefulset/opencoze-opencoze-pg-postgresql -n coze -- \
  psql -U opencoze -d opencoze -c "select 1"

kubectl get deployment opencoze-opencoze-pg-server -n coze -o yaml | grep -A 40 "DATABASE_"
```

确认 `DATABASE_HOST` 指向内部 PostgreSQL Service。
