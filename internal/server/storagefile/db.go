package storagefile

import (
	"os"
	"time"

	"github.com/nickzhog/devops-tool/internal/server/config"
	"github.com/nickzhog/devops-tool/internal/server/metric"
	"github.com/nickzhog/devops-tool/pkg/logging"
)

func StartUpdates(cfg *config.Config, logger *logging.Logger) metric.Storage {
	var storage metric.Storage = metric.NewMemStorage()
	if cfg.Settings.StoreFile == "" {
		return storage
	}

	if cfg.Settings.Restore {
		var err error
		storage, err = getFromFile(cfg.Settings.StoreFile)
		if err != nil {
			logger.Error(err)
		}
	}

	file, err := os.OpenFile(cfg.Settings.StoreFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		logger.Fatal(err)
		return storage
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

	return storage
}

func getFromFile(file string) (metric.Storage, error) {
	var newStorage metric.Storage = metric.NewMemStorage()
	data, err := os.ReadFile(file)
	if err != nil {
		return newStorage, err
	}

	err = newStorage.ImportFromJSON(data)

	return newStorage, err
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
	_, err = f.Write(storage.ExportToJSON())

	return
}
