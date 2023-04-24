package grpc

import (
	"context"
	"errors"

	pb "github.com/nickzhog/devops-tool/internal/proto"
	"github.com/nickzhog/devops-tool/internal/server/server"
	"github.com/nickzhog/devops-tool/pkg/metric"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ pb.MetricsServer = (*MetricServer)(nil)

type MetricServer struct {
	srv server.Server

	pb.UnimplementedMetricsServer
}

func NewMetricServer(srv server.Server) *MetricServer {
	return &MetricServer{
		srv: srv,
	}
}

func (s *MetricServer) SetMetrics(ctx context.Context, in *pb.SetMetricsRequest) (*pb.SetMetricsResponse, error) {
	metrics := make([]metric.Metric, 0, len(in.Metrics))

	for _, pbmetric := range in.Metrics {
		var m metric.Metric

		switch pbmetric.Mtype.String() {
		case metric.GaugeType:
			m = metric.NewGaugeMetric(pbmetric.Id, pbmetric.Value)

		case metric.CounterType:
			m = metric.NewCounterMetric(pbmetric.Id, pbmetric.Delta)
		}

		m.Hash = pbmetric.Hash

		metrics = append(metrics, m)
	}

	err := s.srv.UpsertMany(ctx, metrics)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, err.Error())
	}

	return &pb.SetMetricsResponse{Ok: true}, nil
}

func (s *MetricServer) GetMetrics(ctx context.Context, in *pb.GetMetricsRequest) (*pb.GetMetricsResponse, error) {
	var response pb.GetMetricsResponse

	for _, pbMetric := range in.Request {
		m, err := s.srv.FindMetric(ctx, pbMetric.Id, pbMetric.Mtype.String())
		if err != nil {
			if errors.Is(err, metric.ErrNoResult) {
				return nil, status.Errorf(codes.NotFound, metric.ErrNoResult.Error())
			}
			return nil, status.Errorf(codes.Unknown, err.Error())
		}

		response.Metric = append(response.Metric,
			&pb.Metric{
				Id:    m.ID,
				Mtype: pb.MType(pb.MType_value[m.MType]),
				Value: *m.Value,
				Delta: *m.Delta,
			})
	}

	return &response, nil
}
