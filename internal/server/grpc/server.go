package grpc

import (
	"context"
	"net"

	pb "github.com/nickzhog/devops-tool/internal/proto"
	"github.com/nickzhog/devops-tool/internal/server/config"
	"github.com/nickzhog/devops-tool/internal/server/server"
	"google.golang.org/grpc"
)

func Serve(ctx context.Context, srv server.Server, cfg *config.Config) {
	gRPCsrv := grpc.NewServer()

	if cfg.Settings.TrustedSubnet != "" {
		_, ipNet, err := net.ParseCIDR(cfg.Settings.TrustedSubnet)
		if err != nil {
			srv.Logger.Fatal(err)
		}
		gRPCsrv = grpc.NewServer(
			grpc.UnaryInterceptor(
				NewIPinterceptor(ipNet, srv.Logger),
			),
		)
	}

	pb.RegisterMetricsServer(gRPCsrv, NewMetricServer(srv))
	go func() {
		listen, err := net.Listen("tcp", cfg.Settings.AddressGRPC)
		if err != nil {
			srv.Logger.Fatal(err)
		}
		if err = gRPCsrv.Serve(listen); err != nil && err != grpc.ErrServerStopped {
			srv.Logger.Fatalf("grpc listen:%+s\n", err)
		}
	}()

	srv.Logger.Tracef("grpc server started")

	<-ctx.Done()

	gRPCsrv.GracefulStop()

	srv.Logger.Tracef("grpc server stopped")
}
