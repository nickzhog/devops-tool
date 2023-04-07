package grpc

import (
	"context"
	"errors"

	pb "github.com/nickzhog/devops-tool/internal/proto"
	"github.com/nickzhog/devops-tool/internal/server/config"
	"github.com/nickzhog/devops-tool/internal/server/service"
	"github.com/nickzhog/devops-tool/pkg/logging"
	"github.com/nickzhog/devops-tool/pkg/metric"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MetricServer struct {
	storage service.Storage
	logger  *logging.Logger
	cfg     *config.Config

	pb.UnimplementedMetricsServer
}

func NewMetricServer(logger *logging.Logger, cfg *config.Config, storage service.Storage) *MetricServer {
	return &MetricServer{
		logger:  logger,
		storage: storage,
		cfg:     cfg,
	}
}

func (s *MetricServer) SetMetric(ctx context.Context, req *pb.SetMetricRequest) (*pb.SetMetricResponse, error) {
	var response pb.SetMetricResponse

	for _, pbmetric := range req.Metrics {
		var m metric.Metric

		switch pbmetric.Mtype.String() {
		case metric.GaugeType:
			m = metric.NewGaugeMetric(pbmetric.Id, pbmetric.Value)

		case metric.CounterType:
			m = metric.NewCounterMetric(pbmetric.Id, pbmetric.Delta)
		}

		if s.cfg.Settings.Key != "" && pbmetric.Hash != m.GetHash(s.cfg.Settings.Key) {
			return nil, status.Errorf(codes.DataLoss, metric.ErrWrongHash.Error())
		}

		err := s.storage.UpsertMetric(ctx, m)
		if err != nil {
			return nil, status.Errorf(codes.Unknown, err.Error())
		}
	}

	response.Answer = "ok"
	return &response, nil
}

func (s *MetricServer) GetMetric(ctx context.Context, req *pb.GetMetricRequest) (*pb.GetMetricResponse, error) {
	var response pb.GetMetricResponse

	m, err := s.storage.FindMetric(ctx, req.Metric.Id, req.Metric.Mtype.String())
	if err != nil {
		if errors.Is(err, metric.ErrNoResult) {
			return nil, status.Errorf(codes.NotFound, metric.ErrNoResult.Error())
		}
		return nil, status.Errorf(codes.Unknown, err.Error())
	}

	response.Metric = &pb.Metric{
		Id:    m.ID,
		Mtype: pb.Metric_MType(pb.Metric_MType_value[m.MType]),
		Value: *m.Value,
		Delta: *m.Delta,
	}

	return &response, nil
}
