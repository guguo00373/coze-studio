# Coze Studio Helm Chart (Harbor 版本)

## 目录结构

```
opencoze-harbor/
├── Chart.yaml                    # Chart 元信息
├── values.yaml                   # 配置文件（已修改为 Harbor 镜像地址）
├── templates/
│   ├── _helpers.tpl              # 模板辅助函数
│   ├── configmap.yaml            # 通用 ConfigMap（新增）
│   ├── secret.yaml               # 通用 Secret（新增）
│   ├── deployment.yaml           # coze-server 部署
│   ├── service.yaml              # coze-server 服务
│   ├── coze-web-*.yaml           # 前端相关
│   ├── mysql-*.yaml              # MySQL 相关
│   ├── redis-*.yaml              # Redis 相关
│   ├── elasticsearch-*.yaml      # Elasticsearch 相关
│   ├── minio-*.yaml              # MinIO 相关
│   ├── etcd-*.yaml               # etcd 相关
│   ├── milvus-*.yaml             # Milvus 相关
│   └── rocketmq-*.yaml           # RocketMQ 相关
└── files/                        # 初始化脚本
    └── mysql/
        └── schema.sql
```

## 部署前准备

### 1. 配置 Harbor 仓库 Secret

```bash
kubectl create namespace coze

kubectl create secret docker-registry harbor-secret \
  --docker-server=harbor.ubidirector.cn \
  --docker-username=admin \
  --docker-password='admin@123' \
  -n coze
```

### 2. 修改敏感信息

编辑 `values.yaml` 中的 `secret.stringData`：

```yaml
secret:
  enabled: true
  stringData:
    MYSQL_ROOT_PASSWORD: "your_mysql_root_password"
    MYSQL_PASSWORD: "your_mysql_password"
    MINIO_ROOT_PASSWORD: "your_minio_password"
    OPENAI_API_KEY: "sk-your_openai_api_key"
    OPENAI_EMBEDDING_API_KEY: "sk-your_openai_api_key"
    BUILTIN_CM_OPENAI_API_KEY: "sk-your_openai_api_key"
```

### 3. 配置存储类

修改所有 `storageClassName` 为你的 K8s StorageClass：

```bash
# 查看可用的 StorageClass
kubectl get sc
```

将 `values.yaml` 中的 `storageClassName: ""` 改为实际的 StorageClass 名称。

## 部署命令

```bash
# 安装
helm install opencoze ./opencoze-harbor \
  --namespace coze \
  -f ./opencoze-harbor/values.yaml

# 升级
helm upgrade opencoze ./opencoze-harbor \
  --namespace coze \
  -f ./opencoze-harbor/values.yaml

# 卸载
helm uninstall opencoze -n coze
```

## 验证部署

```bash
# 查看 Pod 状态
kubectl get pods -n coze

# 查看 Service
kubectl get svc -n coze

# 查看日志
kubectl logs -f deployment/coze-server -n coze
kubectl logs -f deployment/coze-web -n coze
```

## 访问系统

```bash
# 查看分配的 IP
kubectl get svc -n coze coze-web

# 浏览器访问 http://<EXTERNAL-IP>
```

## 镜像清单

| 组件 | Harbor 地址 |
|------|-------------|
| 后端 | harbor.ubidirector.cn/coze/coze-studio-server:latest |
| 前端 | harbor.ubidirector.cn/coze/coze-studio-web:latest |
| MySQL | harbor.ubidirector.cn/coze/mysql:8.4.5 |
| Redis | harbor.ubidirector.cn/coze/redis:8.0 |
| Elasticsearch | harbor.ubidirector.cn/coze/elasticsearch:8.18.0 |
| MinIO | harbor.ubidirector.cn/coze/minio:RELEASE.2025-06-13T11-33-47Z-cpuv1 |
| Milvus | harbor.ubidirector.cn/coze/milvus:v2.5.10 |
| etcd | harbor.ubidirector.cn/coze/etcd:3.5 |
| NSQ | harbor.ubidirector.cn/coze/nsq:v1.2.1 |
| RocketMQ | harbor.ubidirector.cn/coze/rocketmq:5.3.2 |
| busybox | harbor.ubidirector.cn/coze/busybox:latest |
| curl | harbor.ubidirector.cn/coze/curl:8.12.1 |

## K8s 规范说明

本 Chart 遵循以下 K8s 部署规范：

1. **ConfigMap** - 非敏感环境变量通过 ConfigMap 提供
2. **Secret** - 密码和 API Key 使用 Secret 存储
3. **Volume** - 数据持久化使用 PVC，不使用本地目录
4. **域名通信** - 组件间通过 K8s Service 名称通信

## 故障排查

### Pod 一直 Pending

```bash
kubectl describe pod <pod-name> -n coze
# 检查是否资源不足或 StorageClass 不存在
```

### 镜像拉取失败

```bash
kubectl get events -n coze --sort-by='.lastTimestamp'
# 检查是否需要 harbor-secret
```

### MySQL 启动失败

```bash
kubectl logs pod/coze-mysql-0 -n coze
# 检查 schema.sql 是否正确挂载
```
