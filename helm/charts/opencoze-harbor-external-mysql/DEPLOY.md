# Coze Studio K8s 部署指南

## 一、环境准备

### 1.1 前置条件

- K8s 集群 (v1.20+)
- Helm 3.x
- kubectl 已配置
- Harbor 仓库可访问

### 1.2 节点配置要求

| 环境 | CPU | 内存 | 存储 | 说明 |
|------|-----|------|------|------|
| 开发/测试 | 2核 | 4Gi | 50Gi | 最小配置，部分组件可关闭 |
| 生产环境 | 8核 | 16Gi | 200Gi | 推荐配置 |
| 高可用 | 16核+ | 32Gi+ | 500Gi+ | 多副本部署 |

> **说明**：默认 values.yaml 中大部分组件未设置 resources，K8s 会按实际需求调度。可根据实际情况调整。

### 1.3 网络要求

- 节点可访问 Harbor 仓库 `harbor.ubidirector.cn`
- 节点之间网络互通

---

## 二、部署步骤

### 2.1 上传 Helm Chart 到目标机器

```bash
# 方式1: scp 传输
scp -r opencoze-harbor user@<target-ip>:/home/user/

# 方式2: 打包传输
tar -czf opencoze-harbor.tar.gz opencoze-harbor/
scp opencoze-harbor.tar.gz user@<target-ip>:/home/user/
# 在目标机器解压
tar -xzf opencoze-harbor.tar.gz
```

### 2.2 创建命名空间

```bash
kubectl create namespace coze
```

### 2.3 配置 Harbor 仓库认证

```bash
kubectl create secret docker-registry harbor-secret \
  --docker-server=harbor.ubidirector.cn \
  --docker-username=admin \
  --docker-password='admin@123' \
  -n coze
```

### 2.4 修改配置文件

编辑 `values.yaml`，按需修改以下配置：

#### 2.4.1 修改敏感信息（必须）

```yaml
secret:
  enabled: true
  stringData:
    MYSQL_ROOT_PASSWORD: "修改为你的MySQL root密码"
    MYSQL_PASSWORD: "修改为你的MySQL密码"
    MINIO_ROOT_PASSWORD: "修改为你的MinIO密码"
    OPENAI_API_KEY: "sk-你的OpenAI-API-Key"
    OPENAI_EMBEDDING_API_KEY: "sk-你的OpenAI-API-Key"
    BUILTIN_CM_OPENAI_API_KEY: "sk-你的OpenAI-API-Key"
```

#### 2.4.2 修改存储类（必须）

查看集群可用的 StorageClass：

```bash
kubectl get sc
```

将 `values.yaml` 中所有 `storageClassName: ""` 改为实际值：

```yaml
mysql:
  persistence:
    storageClassName: "your-storage-class"  # 改成你的

elasticsearch:
  persistence:
    storageClassName: "your-storage-class"

minio:
  persistence:
    storageClassName: "your-storage-class"

# ... 其他组件同理
```

#### 2.4.3 修改 Service 类型（按需）

```yaml
cozeServer:
  service:
    type: LoadBalancer  # 或 NodePort

cozeWeb:
  service:
    type: LoadBalancer  # 或 NodePort
```

#### 2.4.4 修改 AI 模型配置（按需）

```yaml
cozeServer:
  env:
    # 如果使用 OpenAI
    EMBEDDING_TYPE: "openai"
    BUILTIN_CM_TYPE: "openai"
    OPENAI_EMBEDDING_BASE_URL: "https://api.openai.com/v1/embeddings"
    OPENAI_EMBEDDING_MODEL: "text-embedding-3-large"
    
    # 如果使用火山引擎 Ark
    # EMBEDDING_TYPE: "ark"
    # BUILTIN_CM_TYPE: "ark"
    # ARK_EMBEDDING_BASE_URL: "https://ark.cn-beijing.volces.com/api/v3"
    # ARK_EMBEDDING_MODEL: "你的模型ID"
```

### 2.5 安装 Helm Chart

```bash
helm install opencoze ./opencoze-harbor \
  --namespace coze \
  -f ./opencoze-harbor/values.yaml
```

### 2.6 验证部署

```bash
# 查看 Pod 状态（等待所有 Pod 变为 Running）
kubectl get pods -n coze -w

# 查看 Service
kubectl get svc -n coze

# 查看部署状态
helm status opencoze -n coze
```

---

## 三、访问系统

### 3.1 获取访问地址

```bash
kubectl get svc -n coze coze-web
```

### 3.2 访问方式

#### LoadBalancer 类型

```bash
# 等待 EXTERNAL-IP 分配
kubectl get svc -n coze coze-web -w

# 浏览器访问
http://<EXTERNAL-IP>
```

#### NodePort 类型

```bash
# 获取 NodePort
kubectl get svc -n coze coze-web -o jsonpath='{.spec.ports[0].nodePort}'

# 浏览器访问
http://<任意节点IP>:<NodePort>
```

#### ClusterIP + Ingress

```yaml
# values.yaml 中配置 Ingress
ingress:
  enabled: true
  className: "nginx"
  hosts:
    - host: coze.yourdomain.com
      paths:
        - path: /
          pathType: Prefix
```

---

## 四、常用命令

### 4.1 查看日志

```bash
# 后端服务日志
kubectl logs -f deployment/coze-server -n coze

# 前端服务日志
kubectl logs -f deployment/coze-web -n coze

# MySQL 日志
kubectl logs -f pod/coze-mysql-0 -n coze

# Elasticsearch 日志
kubectl logs -f pod/coze-elasticsearch-0 -n coze
```

### 4.2 进入容器

```bash
# 进入后端容器
kubectl exec -it deployment/coze-server -n coze -- /bin/sh

# 进入 MySQL 容器
kubectl exec -it pod/coze-mysql-0 -n coze -- /bin/bash
```

### 4.3 升级配置

```bash
# 修改 values.yaml 后执行
helm upgrade opencoze ./opencoze-harbor \
  --namespace coze \
  -f ./opencoze-harbor/values.yaml
```

### 4.4 卸载

```bash
helm uninstall opencoze -n coze
kubectl delete namespace coze
```

---

## 五、故障排查

### 5.1 Pod 一直 Pending

```bash
kubectl describe pod <pod-name> -n coze
# 常见原因：
# - 资源不足（CPU/Memory）
# - StorageClass 不存在
# - 节点亲和性不满足
```

### 5.2 镜像拉取失败

```bash
kubectl get events -n coze --sort-by='.lastTimestamp'
# 常见原因：
# - harbor-secret 未创建
# - Harbor 仓库不可访问
# - 镜像不存在
```

### 5.3 Pod CrashLoopBackOff

```bash
kubectl logs <pod-name> -n coze --previous
# 常见原因：
# - 配置错误
# - 依赖服务未就绪
# - 数据库连接失败
```

### 5.4 服务无法访问

```bash
# 检查 Service endpoints
kubectl get endpoints -n coze

# 检查 Pod 是否就绪
kubectl get pods -n coze

# 测试集群内访问
kubectl run curl-test --image=harbor.ubidirector.cn/coze/curl:8.12.1 --rm -it -- curl http://coze-web:80
```

---

## 六、组件说明

### 6.1 架构图

```
                    ┌─────────────┐
                    │   Ingress   │
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │   coze-web  │  (Nginx 前端)
                    │   :80       │
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │ coze-server │  (Go 后端)
                    │   :8888     │
                    └──────┬──────┘
                           │
    ┌──────────────────────┼──────────────────────┐
    │                      │                      │
┌───▼───┐  ┌───────────────▼───────┐  ┌──────────▼──────────┐
│ MySQL │  │     Elasticsearch     │  │       MinIO         │
│ :3306 │  │        :9200          │  │      :9000          │
└───────┘  └───────────────────────┘  └─────────────────────┘

┌───────┐  ┌───────────────────────┐  ┌─────────────────────┐
│ Redis │  │        Milvus         │  │      RocketMQ       │
│ :6379 │  │       :19530          │  │      :9876          │
└───────┘  └───────────────────────┘  └─────────────────────┘
```

### 6.2 镜像清单

| 组件 | Harbor 地址 | 用途 |
|------|-------------|------|
| coze-studio-server | harbor.ubidirector.cn/coze/coze-studio-server:latest | 后端服务 |
| coze-studio-web | harbor.ubidirector.cn/coze/coze-studio-web:latest | 前端服务 |
| mysql | harbor.ubidirector.cn/coze/mysql:8.4.5 | 数据库 |
| redis | harbor.ubidirector.cn/coze/redis:8.0 | 缓存 |
| elasticsearch | harbor.ubidirector.cn/coze/elasticsearch:8.18.0 | 搜索引擎 |
| minio | harbor.ubidirector.cn/coze/minio:RELEASE.2025-06-13T11-33-47Z-cpuv1 | 对象存储 |
| milvus | harbor.ubidirector.cn/coze/milvus:v2.5.10 | 向量数据库 |
| etcd | harbor.ubidirector.cn/coze/etcd:3.5 | 配置中心 |
| rocketmq | harbor.ubidirector.cn/coze/rocketmq:5.3.2 | 消息队列 |
| busybox | harbor.ubidirector.cn/coze/busybox:latest | 初始化工具 |
| curl | harbor.ubidirector.cn/coze/curl:8.12.1 | 初始化工具 |

---

## 七、K8s 规范说明

本部署遵循以下规范：

1. **ConfigMap** - 非敏感环境变量通过 ConfigMap 提供
2. **Secret** - 密码和 API Key 使用 Secret 存储
3. **Volume** - 数据持久化使用 PVC，不使用本地目录
4. **域名通信** - 组件间通过 K8s Service 名称通信

---

## 八、资源规划

### 存储需求

| 组件 | Storage |
|------|---------|
| MySQL | 50Gi |
| Elasticsearch | 50Gi |
| Milvus | 20Gi |
| MinIO | 50Gi |
| Redis | 50Gi |
| etcd | 20Gi |
| RocketMQ | 40Gi |
| **总计** | **280Gi** |

### CPU/内存

大部分组件未设置 resources limits，实际需求取决于使用量。建议：

- **开发/测试**：2核 4Gi 可运行（参考官方说明）
- **生产环境**：根据实际使用情况调整

如需限制资源，可在 `values.yaml` 中为各组件添加 `resources` 配置。

---

## 九、安全建议

1. 修改所有默认密码
2. 使用 HTTPS 访问（配置 Ingress TLS）
3. 限制 Service 暴露类型（优先使用 ClusterIP + Ingress）
4. 定期备份数据库
5. 启用审计日志
