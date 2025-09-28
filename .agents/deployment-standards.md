# 部署规范

## 部署环境规范

### 环境分类
- **开发环境(dev)**: 本地开发和功能测试
- **测试环境(test)**: 集成测试和UAT测试
- **预发环境(staging)**: 生产前最后验证
- **生产环境(prod)**: 线上服务环境

### 配置管理
```yaml
# 环境配置文件命名
config/
├── config.yaml          # 默认配置
├── config-dev.yaml      # 开发环境
├── config-test.yaml     # 测试环境
├── config-staging.yaml  # 预发环境
└── config-prod.yaml     # 生产环境
```

## Docker部署规范

### Dockerfile规范
```dockerfile
# 多阶段构建
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制依赖文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源码并构建
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o go-nexus cmd/server/main.go

# 运行时镜像
FROM alpine:3.18

# 安装必要工具
RUN apk --no-cache add ca-certificates tzdata

# 创建用户和目录
RUN adduser -D -s /bin/sh nexus
USER nexus

WORKDIR /app

# 复制二进制文件
COPY --from=builder /app/go-nexus .
COPY --from=builder /app/config ./config

# 创建数据目录
RUN mkdir -p /var/lib/go-nexus

# 暴露端口
EXPOSE 8081

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8081/health || exit 1

# 启动命令
CMD ["./go-nexus", "server", "--config", "config/config.yaml"]
```

### Docker Compose配置
```yaml
# docker-compose.yml
version: '3.8'

services:
  go-nexus:
    build: .
    ports:
      - "8081:8081"
    volumes:
      - nexus-data:/var/lib/go-nexus
      - ./config:/app/config:ro
    environment:
      - GO_NEXUS_LOG_LEVEL=info
      - GO_NEXUS_DATABASE_DSN=postgres://nexus:password@postgres:5432/go_nexus?sslmode=disable
    depends_on:
      - postgres
      - redis
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--spider", "http://localhost:8081/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  postgres:
    image: postgres:13-alpine
    environment:
      POSTGRES_DB: go_nexus
      POSTGRES_USER: nexus
      POSTGRES_PASSWORD: password
    volumes:
      - postgres-data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    restart: unless-stopped

  redis:
    image: redis:6-alpine
    volumes:
      - redis-data:/data
    ports:
      - "6379:6379"
    restart: unless-stopped

volumes:
  nexus-data:
  postgres-data:
  redis-data:
```

## Kubernetes部署规范

### Namespace配置
```yaml
# namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: go-nexus
  labels:
    app: go-nexus
    env: production
```

### ConfigMap配置
```yaml
# configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: go-nexus-config
  namespace: go-nexus
data:
  config.yaml: |
    server:
      host: "0.0.0.0"
      port: 8081
      mode: "release"
    
    database:
      type: "postgresql"
      dsn: "postgres://nexus:$(DATABASE_PASSWORD)@postgres:5432/go_nexus?sslmode=require"
    
    storage:
      type: "s3"
      s3:
        endpoint: "$(S3_ENDPOINT)"
        bucket: "go-nexus-artifacts"
        access_key: "$(S3_ACCESS_KEY)"
        secret_key: "$(S3_SECRET_KEY)"
    
    log:
      level: "info"
      format: "json"
```

### Secret配置
```yaml
# secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: go-nexus-secret
  namespace: go-nexus
type: Opaque
data:
  database-password: <base64-encoded-password>
  jwt-secret: <base64-encoded-jwt-secret>
  s3-access-key: <base64-encoded-s3-access-key>
  s3-secret-key: <base64-encoded-s3-secret-key>
```

### Deployment配置
```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: go-nexus
  namespace: go-nexus
  labels:
    app: go-nexus
spec:
  replicas: 3
  selector:
    matchLabels:
      app: go-nexus
  template:
    metadata:
      labels:
        app: go-nexus
    spec:
      containers:
      - name: go-nexus
        image: laolishu/go-nexus:v0.1.0
        ports:
        - containerPort: 8081
        env:
        - name: DATABASE_PASSWORD
          valueFrom:
            secretKeyRef:
              name: go-nexus-secret
              key: database-password
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: go-nexus-secret
              key: jwt-secret
        - name: S3_ACCESS_KEY
          valueFrom:
            secretKeyRef:
              name: go-nexus-secret
              key: s3-access-key
        - name: S3_SECRET_KEY
          valueFrom:
            secretKeyRef:
              name: go-nexus-secret
              key: s3-secret-key
        - name: S3_ENDPOINT
          value: "https://s3.amazonaws.com"
        volumeMounts:
        - name: config
          mountPath: /app/config
          readOnly: true
        - name: data
          mountPath: /var/lib/go-nexus
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8081
          initialDelaySeconds: 30
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /ready
            port: 8081
          initialDelaySeconds: 5
          periodSeconds: 10
      volumes:
      - name: config
        configMap:
          name: go-nexus-config
      - name: data
        persistentVolumeClaim:
          claimName: go-nexus-data-pvc
```

### Service配置
```yaml
# service.yaml
apiVersion: v1
kind: Service
metadata:
  name: go-nexus-service
  namespace: go-nexus
spec:
  selector:
    app: go-nexus
  ports:
  - protocol: TCP
    port: 8081
    targetPort: 8081
  type: ClusterIP
```

### Ingress配置
```yaml
# ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: go-nexus-ingress
  namespace: go-nexus
  annotations:
    kubernetes.io/ingress.class: "nginx"
    nginx.ingress.kubernetes.io/proxy-body-size: "1g"
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  tls:
  - hosts:
    - nexus.example.com
    secretName: go-nexus-tls
  rules:
  - host: nexus.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: go-nexus-service
            port:
              number: 8081
```

## Helm Chart规范

### Chart.yaml
```yaml
apiVersion: v2
name: go-nexus
description: A lightweight cloud-native repository manager
type: application
version: 0.1.0
appVersion: "0.1.0"
keywords:
- nexus
- repository
- artifacts
- maven
- npm
home: https://github.com/laolishu/go-nexus
sources:
- https://github.com/laolishu/go-nexus
maintainers:
- name: laolishu
  email: maintainer@example.com
```

### values.yaml
```yaml
# Default values for go-nexus
replicaCount: 3

image:
  repository: laolishu/go-nexus
  pullPolicy: IfNotPresent
  tag: "v0.1.0"

service:
  type: ClusterIP
  port: 8081

ingress:
  enabled: false
  className: "nginx"
  annotations: {}
  hosts:
  - host: nexus.local
    paths:
    - path: /
      pathType: Prefix
  tls: []

resources:
  requests:
    memory: "256Mi"
    cpu: "250m"
  limits:
    memory: "512Mi"
    cpu: "500m"

persistence:
  enabled: true
  storageClass: ""
  accessMode: ReadWriteOnce
  size: 10Gi

config:
  server:
    port: 8081
    mode: release
  
  database:
    type: sqlite
    dsn: "/var/lib/go-nexus/go-nexus.db"
  
  storage:
    type: filesystem
    basePath: "/var/lib/go-nexus"

postgresql:
  enabled: false
  auth:
    database: go_nexus
    username: nexus
    password: "changeme"

redis:
  enabled: false
  auth:
    enabled: false
```

## 监控和日志规范

### Prometheus监控
```yaml
# servicemonitor.yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: go-nexus-metrics
  namespace: go-nexus
spec:
  selector:
    matchLabels:
      app: go-nexus
  endpoints:
  - port: metrics
    path: /metrics
    interval: 30s
```

### 日志收集
```yaml
# fluentd配置
<source>
  @type tail
  path /var/log/containers/go-nexus-*.log
  pos_file /var/log/fluentd-go-nexus.log.pos
  tag kubernetes.go-nexus
  format json
  time_format %Y-%m-%dT%H:%M:%S.%NZ
</source>

<filter kubernetes.go-nexus>
  @type kubernetes_metadata
  @id filter_kube_metadata
</filter>

<match kubernetes.go-nexus>
  @type elasticsearch
  host elasticsearch.logging.svc.cluster.local
  port 9200
  index_name go-nexus
</match>
```

## 备份和恢复规范

### 数据备份策略
```bash
#!/bin/bash
# backup.sh

# 数据库备份
kubectl exec -n go-nexus postgres-0 -- pg_dump -U nexus go_nexus | \
    gzip > backup-$(date +%Y%m%d-%H%M%S).sql.gz

# 文件存储备份
kubectl exec -n go-nexus go-nexus-0 -- tar czf - /var/lib/go-nexus | \
    kubectl cp go-nexus/go-nexus-0:- backup-data-$(date +%Y%m%d-%H%M%S).tar.gz
```

### 灾难恢复流程
1. 恢复数据库数据
2. 恢复文件存储数据
3. 重新部署应用
4. 验证数据完整性