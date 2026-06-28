# h-im

Go 微服务即时通讯（IM）系统，基于 Kratos v2 + Wire DI + GORM + gRPC + Kafka。

## 架构

```
Client (WebSocket) ←→ Gateway ──gRPC──→ User / Message / Contact / Group / Chatroom / Media / Push
                    │
                  Sequence (发号器)
```

| 服务 | 端口 | 职责 |
|------|------|------|
| **gateway** | WS:8080, gRPC:9200, HTTP:8180 | WebSocket 接入、消息路由、在线管理 |
| **user** | gRPC:9100, HTTP:8181 | 注册/登录、JWT、在线状态、Token 校验 |
| **message** | gRPC:9102, HTTP:8183 | 消息落库、Kafka 消费、推送触发 |
| **group** | gRPC:9104, HTTP:8185 | 群组 CRUD、成员管理、禁言 |
| **chatroom** | gRPC:9105, HTTP:8186 | 聊天室管理、消息历史 |
| **contact** | gRPC:9106, HTTP:8187 | 好友管理、黑名单 |
| **push** | gRPC:9107, HTTP:8188 | FCM / APNs 推送 |
| **media** | gRPC:9108, HTTP:8189 | 阿里云 OSS 上传（服务端 + 预签名 URL） |
| **sequence** | gRPC:9101, HTTP:8182 | 分布式 ID 发号器 |

## 技术栈

| 组件 | 技术 |
|------|------|
| 语言 | Go 1.24 |
| 框架 | Kratos v2 (DDD 四层架构) |
| DI | Google Wire |
| 数据库 | MySQL (GORM) |
| 缓存 | Redis |
| 消息队列 | Kafka (segmentio/kafka-go) |
| 可观测性 | OpenTelemetry + Prometheus |
| 容器 | Docker (debian:bookworm-slim) |

## 快速开始

### 前置条件

- Go 1.24+
- Docker & Docker Compose
- Make

### 启动基础设施

```bash
make docker-up   # 启动 MySQL + Redis + Kafka
```

### 编译 & 运行

```bash
# 编译全部服务
make build

# 编译单个服务
make build-gateway

# 运行测试
make test

# 代码检查
make lint

# 生成 proto
make proto
```

### 单个服务开发

```bash
cd services/gateway
make all      # proto → wire → tidy → build
make build    # 仅编译
make test     # 运行测试
make wire     # 重新生成 wire_gen.go
```

## 项目结构

```
h-im/
├── proto/                    # Protobuf 定义
│   └── him/{service}/v1/    # 各服务 API
├── gen/go/                   # 生成的 Go 代码
├── pkg/                      # 共享包
│   ├── database/             # GORM MySQL 工厂
│   ├── kafka/                # Kafka producer/consumer
│   ├── redis/                # Redis 客户端
│   ├── tracing/              # OTLP trace exporter
│   ├── metrics/              # Prometheus meter
│   ├── jwt/                  # JWT 签发/校验
│   ├── errcode/              # 领域错误码
│   └── pagination/           # 分页工具
├── services/                 # 微服务
│   └── {name}/
│       ├── cmd/server/       # 入口 + wire
│       └── internal/
│           ├── server/       # 传输层（路由、中间件）
│           ├── service/      # 组装层（gRPC/HTTP handler）
│           ├── biz/          # 业务层（UseCase、接口定义）
│           ├── data/         # 数据层（DB/Redis/Kafka 实现）
│           └── conf/         # 配置定义
├── configs/                  # 开发环境配置
├── docker/                   # Docker Compose
├── deploy/                   # K8s 部署清单
└── docs/                     # 文档
```

## 配置

### 环境变量（生产部署）

| 变量 | 用途 | 服务 |
|------|------|------|
| `JWT_SECRET` | JWT 签名密钥 | user |
| `FCM_CREDENTIALS` | Firebase 服务账号 JSON 路径 | push |
| `APNS_KEY_ID` | Apple APNs Key ID | push |
| `APNS_TEAM_ID` | Apple Team ID | push |
| `APNS_KEY_PATH` | .p8 私钥路径 | push |
| `APNS_BUNDLE_ID` | App Bundle ID | push |
| `MEDIA_SECRET` | 媒体服务共享密钥 | media |

### 配置文件

开发环境配置在 `services/*/configs/config.yaml`。生产环境通过 K8s ConfigMap/Secret 挂载到 `/app/configs/`。

## API 端点

- **gRPC**: 各服务端口见上表
- **HTTP**: `/ping` (健康检查), `/metrics` (Prometheus)
- **WebSocket**: `ws://gateway:8080/ws?app_id={}&user_id={}&token={}&device_id={}`

## 消息流

```
Client → Gateway (WS) → Kafka → Message Service → DB
                                    ├→ Gateway (推送在线用户)
                                    └→ Push Service (离线推送)
```

| Topic | 用途 |
|-------|------|
| `him.private.message` | 私聊消息 |
| `him.chatroom.message` | 聊天室消息 |
| `him.group.message` | 群组消息 |

## 部署

```bash
# Docker 镜像构建
docker build -t him-gateway -f services/gateway/Dockerfile .

# K8s 部署
kubectl apply -f deploy/k8s/
```

## 文档

- [架构设计](docs/architecture.md)
- [设计文档](docs/superpowers/specs/)

## License

MIT
