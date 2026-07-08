package data

import (
	"context"

	pb "github.com/murphy-hc/h-im/gen/go/him/sequence/v1"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

// NewSequenceClient creates a gRPC client for the sequence service.
func NewSequenceClient() (pb.SequenceServiceClient, func(), error) {
	conn, err := grpc.DialInsecure(context.Background(),
		grpc.WithEndpoint("discovery:///sequence.default.svc.cluster.local:9108"),
	)
	if err != nil {
		return nil, nil, err
	}
	return pb.NewSequenceServiceClient(conn), func() { conn.Close() }, nil
}
