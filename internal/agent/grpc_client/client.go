package grpcclient

import (
	pb "github.com/nickzhog/devops-tool/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewClient(port string) pb.MetricsClient {
	conn, err := grpc.Dial(port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}

	c := pb.NewMetricsClient(conn)

	return c
}
