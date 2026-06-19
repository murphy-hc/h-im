module github.com/murphy-hc/h-im/services/sequence

go 1.25.0

require (
	github.com/google/wire v0.7.0
	github.com/jackc/pgx/v5 v5.10.0
	github.com/murphy-hc/h-im/gen/go v0.0.0
	github.com/murphy-hc/h-im/pkg/logger v0.0.0
	google.golang.org/grpc v1.68.0
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	golang.org/x/net v0.29.0 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.29.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240903143218-8af14fe29dc1 // indirect
	google.golang.org/protobuf v1.35.2 // indirect
)

replace (
	github.com/murphy-hc/h-im/gen/go => ../../gen/go
	github.com/murphy-hc/h-im/pkg/logger => ../../pkg/logger
	github.com/murphy-hc/h-im/pkg/snowflake => ../../pkg/snowflake
)
