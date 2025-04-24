package utils

import (
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func CreateGRPCClient(URI string) *grpc.ClientConn {
	conn, err := grpc.NewClient(
		URI,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(200*1024*1024)),
	)
	if err != nil {
		log.Err(err).Msg("failed creating systems service gRPC client")
		return nil
	}

	return conn
}
