package storagefile

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/nickzhog/devops-tool/internal/server/config"
	"github.com/nickzhog/devops-tool/internal/server/service"
	"github.com/nickzhog/devops-tool/pkg/logging"
	"github.com/nickzhog/devops-tool/pkg/metric"
)

var _ StorageFile = (*storageFile)(nil)

type StorageFile interface {
	StartUpdate(ctx context.Context)
}

type storageFile struct {
	file     *os.File
	interval time.Duration
	logger   *logging.Logger
	storage  service.Storage
}

func NewStorageFile(ctx context.Context, cfg *config.Config, logger *logging.Logger, storage service.Storage) StorageFile {
	if cfg.Settings.Restore {
		err := importFromFile(ctx, cfg.Settings.StoreFile, storage)
		if err != nil {
			logger.Tracef("err: %v", err)
		}
	}

	file, err := os.OpenFile(cfg.Settings.StoreFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		logger.Fatal(err)
	}

	return &storageFile{
		interval: cfg.Settings.StoreInterval,
		logger:   logger,
		file:     file,
		storage:  storage,
	}
}

func (s *storageFile) StartUpdate(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	for {
		select {
		case <-ticker.C:
			err := s.updateFile(ctx)
			if err != nil {
				s.logger.Error(err)
			}

		case <-ctx.Done():
			s.logger.Traceln("storage file update stopped")
			return
		}
	}
}

func (s *storageFile) updateFile(ctx context.Context) (err error) {
	err = s.file.Truncate(0)
	if err != nil {
		return
	}
	_, err = s.file.Seek(0, 0)
	if err != nil {
		return
	}

	metrics, err := s.storage.ExportMetrics(ctx)
	if err != nil {
		return
	}

	jsonData, err := json.Marshal(metrics)
	if err != nil {
		return
	}

	_, err = s.file.Write(jsonData)

	// s.logger.Traceln("storage file updated")

	return
}

func importFromFile(ctx context.Context, file string, storage service.Storage) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	var metrics []metric.Metric
	err = json.Unmarshal(data, &metrics)
	if err != nil {
		return err
	}

	for _, v := range metrics {
		_, err := storage.FindMetric(ctx, v.ID, v.MType)
		if err != nil {
			if !errors.Is(err, metric.ErrNoResult) {
				return err
			}
			err = storage.UpsertMetric(ctx, v)
			if err != nil {
				return err
			}
			continue
		}
	}

	return nil
}
