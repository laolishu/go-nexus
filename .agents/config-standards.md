# 配置规范

## 配置文件格式

### 主配置文件
- 使用 YAML 格式，文件名：`config.yaml`
- 支持环境变量覆盖：`GO_NEXUS_` 前缀

### 配置文件结构
```yaml
# 服务配置
server:
  host: "0.0.0.0"
  port: 8081
  mode: "release"  # debug, release, test
  read_timeout: "30s"
  write_timeout: "30s"
  max_header_bytes: 1048576

# 数据库配置
database:
  type: "sqlite"  # sqlite, postgresql
  dsn: "data/go-nexus.db"
  max_open_conns: 25
  max_idle_conns: 10
  conn_max_lifetime: "5m"

# 存储配置
storage:
  type: "filesystem"  # filesystem, s3
  base_path: "/var/lib/go-nexus"
  
# Redis缓存配置（可选）
cache:
  type: "memory"  # memory, redis
  redis:
    addr: "localhost:6379"
    password: ""
    db: 0
    pool_size: 10

# 日志配置
log:
  level: "info"  # debug, info, warn, error
  format: "json"  # json, text
  output: "stdout"  # stdout, file
  file_path: "/var/log/go-nexus/app.log"
  max_size: 100  # MB
  max_backups: 3
  max_age: 28  # days

# 安全配置
security:
  jwt_secret: "your-secret-key"
  jwt_expire: "24h"
  enable_https: false
  cert_file: ""
  key_file: ""

# 插件配置
plugins:
  enabled: ["maven", "npm"]
  path: "/var/lib/go-nexus/plugins"
```

### 环境变量覆盖规则
- `GO_NEXUS_SERVER_PORT=8080` 覆盖 `server.port`
- `GO_NEXUS_DATABASE_DSN=host=localhost` 覆盖 `database.dsn`
- 嵌套配置使用下划线分隔

### 配置验证
- 启动时进行配置验证
- 必填项检查
- 数据类型验证
- 范围值检查

## 配置加载顺序
1. 默认配置（代码中定义）
2. 配置文件（`config.yaml`）
3. 环境变量覆盖
4. 命令行参数覆盖