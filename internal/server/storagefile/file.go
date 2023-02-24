package storagefile

import (
	"context"
	"os"
	"time"

	"github.com/nickzhog/devops-tool/internal/server/config"
	"github.com/nickzhog/devops-tool/internal/server/metric"
	"github.com/nickzhog/devops-tool/pkg/logging"
)

type StorageFile interface {
	StartUpdate(ctx context.Context)
}

type storageFile struct {
	file     *os.File
	interval time.Duration
	logger   *logging.Logger
	storage  metric.Storage
}

func NewStorageFile(ctx context.Context, cfg *config.Config, logger *logging.Logger, storage metric.Storage) StorageFile {
	if cfg.Settings.Restore {
		err := importFromFile(ctx, cfg.Settings.StoreFile, storage)
		if err != nil {
			logger.Error(err)
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
	data, err := s.storage.ExportToJSON(ctx)
	if err != nil {
		return
	}
	_, err = s.file.Write(data)

	return
}

func importFromFile(ctx context.Context, file string, storage metric.Storage) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	err = storage.ImportFromJSON(ctx, data)

	return err
}
