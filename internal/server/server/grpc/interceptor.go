package grpc

import (
	"context"
	"net"

	"github.com/nickzhog/devops-tool/pkg/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func NewIPinterceptor(trustedSubnet *net.IPNet, logger *logging.Logger) func(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.InvalidArgument, "missing metadata")
		}

		ips, ok := md["x-real-ip"]
		if !ok || len(ips) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing client IP")
		}

		ip := net.ParseIP(ips[0])
		if ip == nil {
			return nil, status.Error(codes.InvalidArgument, "invalid client IP")
		}

		if trustedSubnet.Contains(ip) {
			return handler(ctx, req)
		}

		logger.Tracef("wrong ip: %s", ip)
		return nil, status.Error(codes.PermissionDenied, "client IP is not allowed")
	}
}
