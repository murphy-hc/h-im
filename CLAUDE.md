# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

> **⚠️ 约束：禁止自动执行 `git add` 和 `git commit`。** 所有 git 操作需用户明确指示。
>
> **⚠️ 约束：services/ 下所有微服务必须遵循 Kratos DDD 分层**，目录结构统一为 `cmd/` `internal/{server,service,biz,data,conf}/`，依赖方向 `server → service → biz ← data`。

## 项目概述

Go monorepo IM 系统，模块根：`github.com/murphy-hc/h-im`。基于 Kratos v2 + Wire DI + GORM + gRPC + Kafka。

| 组件 | 技术选型 |
|------|---------|
| 语言 / 框架 | Go 1.24, Kratos v2 |
| 数据库 / 缓存 / MQ | MySQL (GORM) / Redis / Kafka |
| 可观测性 | OpenTelemetry + Prometheus + 结构化日志 |
| 容器 | Docker (debian:bookworm-slim), 推送至阿里云 ACR |

## 常用命令

```bash
# 根目录
make build              # 构建全部服务
make build-{name}       # 构建单个服务 (e.g. make build-sequence)
make test               # 全部测试 (go test -race -short ./...)
make test-{name}        # 单个服务测试
make lint               # go vet + golangci-lint
make proto              # buf lint + buf generate (API protos)
make proto-conf         # protoc 生成所有 conf.pb.go
make docker-up          # 启动本地 MySQL + Redis + Kafka

# 单个服务内
cd services/{name}
make all                # proto → wire → tidy → build
make build              # 编译到 bin/server
make test               # 带 race detector 的测试
make wire               # 生成 wire_gen.go
```

## 架构概览

```
Client (WS) ←→ Gateway ──gRPC──→ Auth / User / Message / Contact / Group / Chatroom / Media / Push
                    │
                  Sequence (发号器, MySQL SELECT FOR UPDATE)
```

- **gateway**：唯一客户端入口，WS 8080 + gRPC 9200，完整 DDD 分层（含 data/）
- **message**：消息管线——落库 → Kafka 投递。gRPC 9102
- 详细架构、端口表、K8s 部署 → `docs/architecture.md`

## 模块约定

- 模块路径：`github.com/murphy-hc/h-im/{submodule}`
- API protos：`proto/him/{service}/v1/` → `gen/go/him/{service}/v1/`（Buf）
- 配置 protos：`services/{name}/internal/conf/conf.proto`（protoc）
- Wire DI：所有服务使用 `wireApp(bc *conf.Bootstrap, meter metric.Meter) (*kratos.App, func(), error)`
- gRPC 中间件链：`recovery → tracing → metadata → metrics`
- HTTP 端点：`/metrics` (Prometheus), `/ping` (健康检查)
- K8s 服务发现：`discovery:///{service}.default.svc.cluster.local:{port}`

## 共享包 (`pkg/`)

| 包 | 用途 |
|---|-----|
| `pkg/database` | GORM MySQL 连接工厂 |
| `pkg/kafka` | Kafka producer/consumer（stub，待实现） |
| `pkg/redis` | Redis 客户端封装 |
| `pkg/tracing` | OTLP trace exporter |
| `pkg/metrics` | Prometheus meter |
| `pkg/logger` | slog 日志工厂 |
| `pkg/jwt` | JWT 签发/校验 |
| `pkg/errcode` | 领域错误码 |
| `pkg/pagination` | 分页工具 |

## 按需参考

- 详细架构、消息流、端口表、K8s 部署细节 → `docs/architecture.md`
- 单个服务 DDD 分层、proto 文件、biz/data 接口等 → 直接读取 `services/{name}/internal/` 下的代码，分层结构统一
- 添加新 gRPC 方法流程 → `docs/architecture.md#adding-grpc-method`
