module github.com/murphy-hc/h-im/services/user

go 1.24.0

toolchain go1.24.13

require (
	github.com/google/wire v0.7.0
	github.com/murphy-hc/h-im/gen/go v0.0.0
	github.com/murphy-hc/h-im/pkg/logger v0.0.0
	google.golang.org/grpc v1.75.0
)

require (
	go.opentelemetry.io/otel/metric v1.38.0 // indirect
	go.opentelemetry.io/otel/sdk v1.38.0 // indirect
	go.opentelemetry.io/otel/trace v1.38.0 // indirect
	golang.org/x/net v0.43.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	golang.org/x/text v0.29.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250825161204-c5933d9347a5 // indirect
	google.golang.org/protobuf v1.36.8 // indirect
)

replace (
	github.com/murphy-hc/h-im/gen/go => ../../gen/go
	github.com/murphy-hc/h-im/pkg/errcode => ../../pkg/errcode
	github.com/murphy-hc/h-im/pkg/jwt => ../../pkg/jwt
	github.com/murphy-hc/h-im/pkg/logger => ../../pkg/logger
	github.com/murphy-hc/h-im/pkg/postgres => ../../pkg/postgres
)
