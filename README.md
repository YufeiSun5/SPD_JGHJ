# Gin + MQTT + PostgreSQL 项目

这是一个基于 Go 语言的 Web API 项目，集成了以下技术栈：

- **Gin**: 高性能的 HTTP Web 框架
- **MQTT**: 轻量级消息传输协议
- **PostgreSQL**: 关系型数据库
- **GORM**: Go 语言 ORM 库

## 项目结构

```
.
├── main.go              # 主程序入口
├── config/              # 配置管理
│   └── config.go
├── database/            # 数据库连接
│   └── database.go
├── mqtt/                # MQTT 客户端
│   └── mqtt.go
├── handlers/            # HTTP 处理器
│   └── handlers.go
├── .env.example         # 环境变量示例
├── go.mod              # Go 模块文件
└── README.md           # 项目说明
```

## 快速开始

### 1. 克隆项目

```bash
git clone <your-repo-url>
cd gin-mqtt-pgsql
```

### 2. 安装依赖

```bash
go mod tidy
```

### 3. 配置环境变量

复制 `.env.example` 为 `.env` 并修改相应配置：

```bash
cp .env.example .env
```

### 4. 启动 PostgreSQL 数据库

使用 Docker 快速启动 PostgreSQL：

```bash
docker run --name postgres-db -e POSTGRES_PASSWORD=password -e POSTGRES_DB=gin_app -p 5432:5432 -d postgres:13
```

### 5. 启动 MQTT Broker

使用 Docker 启动 Mosquitto MQTT Broker：

```bash
docker run -it -p 1883:1883 -p 9001:9001 eclipse-mosquitto
```

### 6. 运行项目

```bash
go run main.go
```

服务器将在 `http://localhost:8080` 启动。

## API 端点

### 基础端点

- `GET /` - 欢迎页面
- `GET /health` - 健康检查
- `GET /status` - 系统状态
- `GET /api/v1/ping` - Ping 测试

### MQTT 端点

- `POST /api/v1/mqtt/publish` - 发布 MQTT 消息

#### 发布 MQTT 消息示例

```bash
curl -X POST http://localhost:8080/api/v1/mqtt/publish \
  -H "Content-Type: application/json" \
  -d '{
    "topic": "test/topic",
    "message": "Hello MQTT!"
  }'
```

## 环境变量

| 变量名 | 描述 | 默认值 |
|--------|------|--------|
| `SERVER_PORT` | 服务器端口 | `8080` |
| `GIN_MODE` | Gin 运行模式 | `debug` |
| `DB_HOST` | 数据库主机 | `localhost` |
| `DB_PORT` | 数据库端口 | `5432` |
| `DB_USER` | 数据库用户名 | `postgres` |
| `DB_PASSWORD` | 数据库密码 | `password` |
| `DB_NAME` | 数据库名称 | `gin_app` |
| `DB_SSLMODE` | SSL 模式 | `disable` |
| `MQTT_BROKER` | MQTT Broker 地址 | `localhost` |
| `MQTT_PORT` | MQTT 端口 | `1883` |
| `MQTT_CLIENT_ID` | MQTT 客户端 ID | `gin-app-client` |
| `MQTT_USERNAME` | MQTT 用户名 | (空) |
| `MQTT_PASSWORD` | MQTT 密码 | (空) |

## 开发

### 添加新的 API 端点

1. 在 `handlers/` 目录下创建新的处理器函数
2. 在 `main.go` 中注册路由
3. 根据需要更新数据库模型

### 数据库迁移

使用 GORM 的自动迁移功能：

```go
// 在 database/database.go 中添加
func AutoMigrate() error {
    return DB.AutoMigrate(&YourModel{})
}
```

### MQTT 订阅

在 `mqtt/mqtt.go` 中添加订阅逻辑：

```go
func init() {
    // 订阅主题
    Subscribe("your/topic", func(client MQTT.Client, msg MQTT.Message) {
        // 处理接收到的消息
        log.Printf("Received: %s", msg.Payload())
    })
}
```

## 部署

### 使用 Docker

创建 `Dockerfile`：

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
CMD ["./main"]
```

构建和运行：

```bash
docker build -t gin-mqtt-pgsql .
docker run -p 8080:8080 gin-mqtt-pgsql
```

## 许可证

MIT License




























