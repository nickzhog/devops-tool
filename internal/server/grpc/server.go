package grpc

import (
	"context"
	"net"

	pb "github.com/nickzhog/devops-tool/internal/proto"
	"github.com/nickzhog/devops-tool/internal/server/config"
	"github.com/nickzhog/devops-tool/internal/server/service"
	"github.com/nickzhog/devops-tool/pkg/logging"
	"google.golang.org/grpc"
)

func Serve(ctx context.Context, logger *logging.Logger, cfg *config.Config, storage service.Storage) {
	srv := grpc.NewServer()
	pb.RegisterMetricsServer(srv, NewMetricServer(logger, cfg, storage))
	go func() {
		listen, err := net.Listen("tcp", cfg.Settings.PortGRPC)
		if err != nil {
			logger.Fatal(err)
		}
		if err = srv.Serve(listen); err != nil && err != grpc.ErrServerStopped {
			logger.Fatalf("grpc listen:%+s\n", err)
		}
	}()

	logger.Tracef("grpc server started")

	<-ctx.Done()

	srv.GracefulStop()

	logger.Tracef("grpc server stopped")
}
