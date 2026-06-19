module github.com/murphy-hc/h-im/services/media

go 1.24

require (
	github.com/google/wire v0.7.0
	github.com/murphy-hc/h-im/gen/go v0.0.0
	github.com/murphy-hc/h-im/pkg/logger v0.0.0
	google.golang.org/grpc v1.68.0
)

require (
	golang.org/x/net v0.29.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240903143218-8af14fe29dc1 // indirect
	google.golang.org/protobuf v1.35.2 // indirect
)

replace (
	github.com/murphy-hc/h-im/gen/go => ../../gen/go
	github.com/murphy-hc/h-im/pkg/errcode => ../../pkg/errcode
	github.com/murphy-hc/h-im/pkg/logger => ../../pkg/logger
)
