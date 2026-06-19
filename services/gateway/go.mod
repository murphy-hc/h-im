module github.com/murphy-hc/h-im/services/gateway

go 1.24

require (
	github.com/google/wire v0.7.0
	github.com/gorilla/websocket v1.5.3
	github.com/murphy-hc/h-im/pkg/logger v0.0.0
)

replace (
	github.com/murphy-hc/h-im/pkg/jwt => ../../pkg/jwt
	github.com/murphy-hc/h-im/pkg/logger => ../../pkg/logger
)
