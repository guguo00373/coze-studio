# Coze Studio K8s 部署文档

## 一、Harbor 仓库配置

### 1. Docker 信任 Harbor（跳过证书验证）

Docker Desktop -> Settings -> Docker Engine，添加：

```json
{
  "insecure-registries": ["harbor.ubidirector.cn"]
}
```

或者修改 `%USERPROFILE%\.docker\daemon.json`，重启 Docker。

### 2. 登录 Harbor

```powershell
docker login harbor.ubidirector.cn -u admin
# 输入密码: admin@123
```

---

## 二、推送镜像到 Harbor

### 需要推送的镜像清单

| 应用 | 原始镜像 | Harbor 地址 |
|------|----------|-------------|
| 后端 | cozedev/coze-studio-server:latest | harbor.ubidirector.cn/coze/coze-studio-server:latest |
| 前端 | cozedev/coze-studio-web:latest | harbor.ubidirector.cn/coze/coze-studio-web:latest |
| MySQL | mysql:8.4.5 | harbor.ubidirector.cn/coze/mysql:8.4.5 |
| Redis | bitnamilegacy/redis:8.0 | harbor.ubidirector.cn/coze/redis:8.0 |
| ES | bitnamilegacy/elasticsearch:8.18.0 | harbor.ubidirector.cn/coze/elasticsearch:8.18.0 |
| MinIO | minio/minio:RELEASE.2025-06-13T11-33-47Z-cpuv1 | harbor.ubidirector.cn/coze/minio:RELEASE.2025-06-13T11-33-47Z-cpuv1 |
| Milvus | milvusdb/milvus:v2.5.10 | harbor.ubidirector.cn/coze/milvus:v2.5.10 |
| etcd | bitnamilegacy/etcd:3.5 | harbor.ubidirector.cn/coze/etcd:3.5 |
| NSQ | nsqio/nsq:v1.2.1 | harbor.ubidirector.cn/coze/nsq:v1.2.1 |
| RocketMQ | apache/rocketmq:5.3.2 | harbor.ubidirector.cn/coze/rocketmq:5.3.2 |
| busybox | busybox:latest | harbor.ubidirector.cn/coze/busybox:latest |
| curl | alpine/curl:8.12.1 | harbor.ubidirector.cn/coze/curl:8.12.1 |

### 一键推送脚本

```powershell
# 在 coze-studio 目录下执行
$Harbor = "harbor.ubidirector.cn"
$Project = "coze"

$images = @(
    "cozedev/coze-studio-server:latest"
    "cozedev/coze-studio-web:latest"
    "mysql:8.4.5"
    "bitnamilegacy/redis:8.0"
    "bitnamilegacy/elasticsearch:8.18.0"
    "minio/minio:RELEASE.2025-06-13T11-33-47Z-cpuv1"
    "milvusdb/milvus:v2.5.10"
    "bitnamilegacy/etcd:3.5"
    "nsqio/nsq:v1.2.1"
    "apache/rocketmq:5.3.2"
    "busybox:latest"
    "alpine/curl:8.12.1"
)

foreach ($img in $images) {
    $parts = $img -split "/"
    $name = if ($parts.Count -gt 2) { ($parts[1..($parts.Count-1)]) -join "/" } else { $parts[1] }
    $tagged = "$Harbor/$Project/$name"
    
    Write-Host ">>> $img -> $tagged" -ForegroundColor Yellow
    docker tag $img $tagged
    docker push $tagged
}

Write-Host "Done!" -ForegroundColor Green
```

### 验证推送结果

浏览器访问：https://harbor.ubidirector.cn -> coze 项目，确认所有镜像已上传。

---

## 三、修改 Helm Chart 配置

编辑 `helm/charts/opencoze/values.yaml`，修改以下内容：

### 1. 镜像地址替换

用 Harbor 地址替换所有镜像地址：

```yaml
mysql:
  image:
    repository: harbor.ubidirector.cn/coze/mysql
    tag: 8.4.5

redis:
  image:
    repository: harbor.ubidirector.cn/coze/redis
    tag: "8.0"

elasticsearch:
  image:
    repository: harbor.ubidirector.cn/coze/elasticsearch
    tag: 8.18.0

minio:
  image:
    repository: harbor.ubidirector.cn/coze/minio
    tag: RELEASE.2025-06-13T11-33-47Z-cpuv1

etcd:
  image:
    repository: harbor.ubidirector.cn/coze/etcd
    tag: 3.5

milvus:
  image:
    repository: harbor.ubidirector.cn/coze/milvus
    tag: v2.5.10

cozeServer:
  image:
    repository: harbor.ubidirector.cn/coze/coze-studio-server
    tag: latest
    pullPolicy: IfNotPresent

cozeWeb:
  image:
    repository: harbor.ubidirector.cn/coze/coze-studio-web
    tag: latest
    pullPolicy: IfNotPresent

rocketmq:
  namesrv:
    image:
      repository: harbor.ubidirector.cn/coze/rocketmq
      tag: 5.3.2
  broker:
    image:
      repository: harbor.ubidirector.cn/coze/rocketmq
      tag: 5.3.2

images:
  busybox: harbor.ubidirector.cn/coze/busybox:latest
  curl: harbor.ubidirector.cn/coze/curl:8.12.1
```

### 2. AI 模型配置（必改）

```yaml
cozeServer:
  env:
    # Ark 模型（火山引擎）
    ARK_EMBEDDING_BASE_URL: "https://ark.cn-beijing.volces.com/api/v3"
    ARK_EMBEDDING_API_KEY: "你的API Key"
    ARK_EMBEDDING_MODEL: "你的模型ID"
    
    # 内置对话模型
    BUILTIN_CM_TYPE: "openai"  # 或 ark
    BUILTIN_CM_OPENAI_BASE_URL: "https://api.openai.com/v1/chat/completions"
    BUILTIN_CM_OPENAI_API_KEY: "你的API Key"
    BUILTIN_CM_OPENAI_MODEL: "gpt-4o"
```

### 3. 存储配置（按需修改）

```yaml
mysql:
  persistence:
    storageClassName: "your-storage-class"  # 改成你的 StorageClass
    size: "50Gi"

elasticsearch:
  persistence:
    storageClassName: "your-storage-class"
    size: 50Gi

minio:
  persistence:
    storageClassName: "your-storage-class"
    size: "50Gi"
```

### 4. Service 类型（按需修改）

```yaml
cozeServer:
  service:
    type: ClusterIP  # 或 NodePort

cozeWeb:
  service:
    type: LoadBalancer  # 或 NodePort / ClusterIP + Ingress
```

---

## 四、部署到 K8s

### 1. 创建命名空间

```powershell
kubectl create namespace coze
```

### 2. 创建 Harbor 仓库 Secret（私有仓库需要）

```powershell
kubectl create secret docker-registry harbor-secret `
  --docker-server=harbor.ubidirector.cn `
  --docker-username=admin `
  --docker-password='admin@123' `
  -n coze
```

### 3. 安装 Helm Chart

```powershell
helm install opencoze ./helm/charts/opencoze `
  --namespace coze `
  -f ./helm/charts/opencoze/values.yaml
```

### 4. 检查部署状态

```powershell
# 查看 Pod 状态
kubectl get pods -n coze -w

# 查看 Service
kubectl get svc -n coze

# 查看日志
kubectl logs -f deployment/coze-server -n coze
kubectl logs -f deployment/coze-web -n coze
```

### 5. 访问系统

```powershell
# 查看分配的 IP
kubectl get svc -n coze coze-web

# 如果是 LoadBalancer 类型，等待 EXTERNAL-IP 分配
# 如果是 NodePort 类型，访问 <NodeIP>:<NodePort>
```

---

## 五、K8s 部署规范（必须遵守）

根据 K8s 平台负责人要求，部署时需遵守以下规范：

### 1. 环境变量通过 ConfigMap 提供

**不要**：在 values.yaml 或代码中硬编码环境变量
**应该**：创建 ConfigMap，通过 envFrom 引用

```yaml
# 示例：创建 ConfigMap
apiVersion: v1
kind: ConfigMap
metadata:
  name: coze-config
  namespace: coze
data:
  LOG_LEVEL: "debug"
  STORAGE_TYPE: "minio"
  ES_VERSION: "v8"
  VECTOR_STORE_TYPE: "milvus"
  EMBEDDING_TYPE: "openai"
  BUILTIN_CM_TYPE: "openai"

# Deployment 中引用
envFrom:
  - configMapRef:
      name: coze-config
```

### 2. 密码使用 Secret

**不要**：明文写入配置文件、代码、values.yaml
**应该**：创建 Secret，通过 envFrom 或 volumeMount 引用

```yaml
# 示例：创建 Secret
apiVersion: v1
kind: Secret
metadata:
  name: coze-secret
  namespace: coze
type: Opaque
stringData:
  MYSQL_ROOT_PASSWORD: "your-password"
  MYSQL_PASSWORD: "your-password"
  MINIO_ROOT_PASSWORD: "your-password"
  OPENAI_API_KEY: "sk-xxx"

# Deployment 中引用
envFrom:
  - secretRef:
      name: coze-secret
```

```powershell
# 或通过命令行创建
kubectl create secret generic coze-secret \
  --from-literal=MYSQL_ROOT_PASSWORD='your-password' \
  --from-literal=MYSQL_PASSWORD='your-password' \
  --from-literal=MINIO_ROOT_PASSWORD='your-password' \
  --from-literal=OPENAI_API_KEY='sk-xxx' \
  -n coze
```

### 3. 数据使用 Volume，不要用本地目录

**不要**：hostPath、emptyDir（持久化数据）
**应该**：使用 PVC + StorageClass

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: mysql-data
  namespace: coze
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: "your-storage-class"  # 改成你的 StorageClass
  resources:
    requests:
      storage: 50Gi
```

### 4. 组件间通过域名通信

**不要**：localhost、127.0.0.1、硬编码 IP
**应该**：使用 K8s Service 名称（Helm chart 已自动处理）

```
# 同一命名空间内直接用 Service 名称
coze-mysql:3306
coze-redis:6379
coze-elasticsearch:9200
coze-minio:9000
coze-milvus:19530
coze-etcd:2379
coze-rocketmq-namesrv:9876
```

### values.yaml 改造建议

基于以上规范，values.yaml 中的 `cozeServer.env` 应该清空敏感信息，改为引用 ConfigMap 和 Secret：

```yaml
cozeServer:
  env: []  # 清空，改用 envFrom
  
  envFrom:
    - configMapRef:
        name: coze-config
    - secretRef:
        name: coze-secret
```

---

## 六、K8s 资源规划

| 组件 | CPU Request | Memory Request | Storage | 副本数 |
|------|-------------|----------------|---------|--------|
| MySQL | 2核 | 4Gi | 50Gi | 1 |
| Elasticsearch | 2核 | 4Gi | 50Gi | 1 |
| Milvus | 2核 | 4Gi | 20Gi | 1 |
| MinIO | 4核 | 8Gi | 50Gi | 1 |
| Redis | 1核 | 1Gi | 50Gi | 1 |
| etcd | 1核 | 1Gi | 20Gi | 1 |
| RocketMQ Namesrv | 1核 | 2Gi | 20Gi | 1 |
| RocketMQ Broker | 4核 | 8Gi | 20Gi | 1 |
| coze-server | 2核 | 4Gi | - | 1 |
| coze-web | 1核 | 512Mi | - | 1 |
| **总计** | **20核** | **35.5Gi** | **280Gi** | |

---

## 六、常见问题

### 1. Pod 一直 Pending

```powershell
kubectl describe pod <pod-name> -n coze
# 检查是否资源不足或 StorageClass 不存在
```

### 2. 镜像拉取失败

```powershell
kubectl get events -n coze --sort-by='.lastTimestamp'
# 检查是否需要 harbor-secret
```

### 3. MySQL 启动失败

```powershell
kubectl logs pod/coze-mysql-0 -n coze
# 检查 schema.sql 是否正确挂载
```

### 4. Elasticsearch 不健康

```powershell
kubectl logs pod/coze-elasticsearch-0 -n coze
# 通常需要等待 1-2 分钟完成初始化
```

---

## 七、卸载

```powershell
helm uninstall opencoze -n coze
kubectl delete namespace coze
```
