package agent

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"sync"

	pb "github.com/nickzhog/devops-tool/internal/proto"

	"github.com/nickzhog/devops-tool/internal/agent/config"
	grpcclient "github.com/nickzhog/devops-tool/internal/agent/grpc_client"
	"github.com/nickzhog/devops-tool/pkg/encryption"
	"github.com/nickzhog/devops-tool/pkg/logging"
	"github.com/nickzhog/devops-tool/pkg/metric"
)

var _ Agent = (*agent)(nil)

type Agent interface {
	UpdateMetrics()
	SendMetricsHTTP(ctx context.Context)
	ImportMetrics([]metric.Metric) error
	ExportMetrics() []metric.Metric
}

type agent struct {
	cfg    *config.Config
	logger *logging.Logger

	publicKey *rsa.PublicKey

	grpcClient pb.MetricsClient

	gaugeMetrics   map[string]float64
	counterMetrics map[string]int64
	mutex          *sync.RWMutex
}

func NewAgent(cfg *config.Config, logger *logging.Logger) *agent {
	agent := &agent{
		cfg:            cfg,
		logger:         logger,
		mutex:          new(sync.RWMutex),
		gaugeMetrics:   make(map[string]float64),
		counterMetrics: make(map[string]int64),
	}

	if cfg.Settings.CryptoKey != "" {
		pubKey, err := encryption.NewPublicKey(cfg.Settings.CryptoKey)
		if err != nil {
			logger.Fatal(err)
		}

		agent.publicKey = pubKey
	}

	if cfg.Settings.AddressGRPC != "" {
		agent.grpcClient = grpcclient.NewClient(cfg.Settings.AddressGRPC)
	}

	return agent
}

func (a *agent) UpdateMetrics() {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	setGaugeMetrics(a.gaugeMetrics)

	a.counterMetrics["PollCount"]++
}

func (a *agent) SendMetricsHTTP(ctx context.Context) {
	var url string
	var answer []byte
	var err error

	metrics := a.ExportMetrics()

	jsonMetrics, err := json.Marshal(metrics)
	if err != nil {
		a.logger.Error(err)
	}

	a.mutex.RLock()
	defer a.mutex.RUnlock()

	for k, v := range a.gaugeMetrics {
		url = fmt.Sprintf("%s/update/gauge/%s/%v", a.cfg.Settings.Address, k, v)

		a.sendRequest(ctx, url, nil)

		////

		url = fmt.Sprintf("%s/update", a.cfg.Settings.Address)

		metric := metric.NewGaugeMetric(k, v)
		metric.Hash = metric.GetHash(a.cfg.Settings.Key)

		body, _ := json.Marshal(metric)
		answer, err = a.sendRequest(ctx, url, body)
	}

	for k, v := range a.counterMetrics {
		url = fmt.Sprintf("%s/update/counter/%s/%v", a.cfg.Settings.Address, k, v)

		a.sendRequest(ctx, url, nil)

		////

		url = fmt.Sprintf("%s/update", a.cfg.Settings.Address)

		metric := metric.NewCounterMetric(k, v)
		metric.Hash = metric.GetHash(a.cfg.Settings.Key)

		body, _ := json.Marshal(metric)
		answer, err = a.sendRequest(ctx, url, body)
	}

	a.logger.Tracef("metrics sended to: %s, last err: %v, last answer: %s", a.cfg.Settings.Address, err, answer)

	url = fmt.Sprintf("%s/updates/", a.cfg.Settings.Address)
	_, err = a.sendRequest(ctx, url, jsonMetrics)
	if err != nil {
		a.logger.Error(err)
	}
}

func (a *agent) SendMetricsGRPC(ctx context.Context) error {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	var request pb.SetMetricsRequest
	for k, v := range a.gaugeMetrics {
		metric := metric.NewGaugeMetric(k, v)
		metric.Hash = metric.GetHash(a.cfg.Settings.Key)

		request.Metrics = append(request.Metrics, &pb.Metric{
			Id:    metric.ID,
			Mtype: pb.MType_gauge,
			Value: *metric.Value,
			Hash:  metric.Hash,
		})
	}
	for k, v := range a.counterMetrics {
		metric := metric.NewCounterMetric(k, v)
		metric.Hash = metric.GetHash(a.cfg.Settings.Key)

		request.Metrics = append(request.Metrics, &pb.Metric{
			Id:    metric.ID,
			Mtype: pb.MType_counter,
			Delta: *metric.Delta,
			Hash:  metric.Hash,
		})
	}
	_, err := a.grpcClient.SetMetrics(ctx, &request)
	if err != nil {
		a.logger.Error(err)
		return err
	}

	return nil
}

func (a *agent) ImportMetrics(metrics []metric.Metric) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	for _, m := range metrics {
		switch m.MType {
		case metric.CounterType:
			a.counterMetrics[m.ID] = *m.Delta
		case metric.GaugeType:
			a.gaugeMetrics[m.ID] = *m.Value
		default:
			return fmt.Errorf("wrong metric type: %s", m.MType)
		}
	}

	return nil
}

func (a *agent) ExportMetrics() []metric.Metric {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	metrics := make([]metric.Metric, 0, len(a.counterMetrics)+len(a.gaugeMetrics))
	for k, v := range a.gaugeMetrics {
		m := metric.NewGaugeMetric(k, v)
		if a.cfg.Settings.Key != "" {
			m.Hash = m.GetHash(a.cfg.Settings.Key)
		}

		metrics = append(metrics, m)
	}
	for k, v := range a.counterMetrics {
		m := metric.NewCounterMetric(k, v)
		if a.cfg.Settings.Key != "" {
			m.Hash = m.GetHash(a.cfg.Settings.Key)
		}

		metrics = append(metrics, m)
	}

	return metrics
}
