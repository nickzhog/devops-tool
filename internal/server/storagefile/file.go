package storagefile

import (
	"context"
	"os"
	"time"

	"github.com/nickzhog/devops-tool/internal/server/config"
	"github.com/nickzhog/devops-tool/internal/server/metric"
	"github.com/nickzhog/devops-tool/pkg/logging"
)

func NewStorageFile(cfg *config.Config, logger *logging.Logger, storage metric.Storage) {
	if cfg.Settings.Restore {
		err := importFromFile(cfg.Settings.StoreFile, storage)
		if err != nil {
			logger.Error(err)
		}
	}

	file, err := os.OpenFile(cfg.Settings.StoreFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		logger.Fatal(err)
	}

	go func() {
		for {
			err = updateFile(storage, file)
			if err != nil {
				logger.Error(err)
			}
			time.Sleep(cfg.Settings.StoreInterval)
		}
	}()
}

func importFromFile(file string, storage metric.Storage) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	err = storage.ImportFromJSON(context.TODO(), data)

	return err
}

func updateFile(storage metric.Storage, f *os.File) (err error) {
	err = f.Truncate(0)
	if err != nil {
		return
	}
	_, err = f.Seek(0, 0)
	if err != nil {
		return
	}
	data, err := storage.ExportToJSON(context.TODO())
	if err != nil {
		return
	}
	_, err = f.Write(data)

	return
}
