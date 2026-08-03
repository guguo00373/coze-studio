# Coze Studio K8s 部署

## 一、部署前准备

### 1. 查看 StorageClass

```bash
kubectl get sc
```

### 2. 修改 values.yaml

把所有 `storageClassName: ""` 改成你集群的 StorageClass 名称：

```yaml
mysql:
  persistence:
    storageClassName: "你的StorageClass"

elasticsearch:
  persistence:
    storageClassName: "你的StorageClass"

minio:
  persistence:
    storageClassName: "你的StorageClass"

# ... 其他组件同理
```

---

## 二、部署命令

```bash
# 1. 创建命名空间
kubectl create namespace coze

# 2. 创建 Harbor 认证
kubectl create secret docker-registry harbor-secret \
  --docker-server=harbor.ubidirector.cn \
  --docker-username=admin \
  --docker-password='admin@123' \
  -n coze

# 3. 部署
helm install opencoze ./opencoze-harbor \
  --namespace coze \
  -f ./opencoze-harbor/values.yaml

# 4. 查看状态
kubectl get pods -n coze
kubectl get svc -n coze
```

---

## 三、访问系统

```bash
# 获取访问地址
kubectl get svc -n coze coze-web
```

---

## 四、常用命令

```bash
# 查看日志
kubectl logs -f deployment/coze-server -n coze
kubectl logs -f deployment/coze-web -n coze

# 升级
helm upgrade opencoze ./opencoze-harbor \
  --namespace coze \
  -f ./opencoze-harbor/values.yaml

# 卸载
helm uninstall opencoze -n coze
kubectl delete namespace coze
```

---

## 五、故障排查

```bash
# 查看 Pod 状态
kubectl get pods -n coze

# 查看事件
kubectl get events -n coze --sort-by='.lastTimestamp'

# 查看描述
kubectl describe pod <pod-name> -n coze
```

# 其他
## 模型思考模式关闭



```
要在界面上配置，需要在模型管理里：
1. 进入 模型管理 → 找到 Qwen/Qwen3.6-27B
2. 点击编辑，看是否有 思考能力 的开关
3. 如果没有，就只能手动改数据库
手动关闭（如果界面上没有）：
kubectl exec -it -n coze opencoze-mysql-0 -- mysql -u coze -pcoze123 opencoze -e "UPDATE model_instance SET capability = JSON_SET(capability, '$.reasoning', false) WHERE id=100002;"
然后重启server：
kubectl rollout restart deployment/opencoze-server -n coze
先看看界面上模型编辑有没有这个开关，没有的话再用SQL改。
```

