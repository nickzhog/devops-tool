package storagedb

import (
	"context"
	"encoding/json"
	"time"

	"github.com/nickzhog/devops-tool/internal/server/config"
	"github.com/nickzhog/devops-tool/internal/server/metric"
	"github.com/nickzhog/devops-tool/pkg/logging"
)

func StartUpdates(rep metric.Repository, cfg *config.Config, logger *logging.Logger) metric.Storage {
	var storage metric.Storage = metric.NewMemStorage()
	rep.CreateTable(context.TODO())

	if cfg.Settings.Restore {
		var err error
		storage, err = getFromDB(rep)
		if err != nil {
			logger.Error(err)
		}
	}
	go func() {
		for {
			err := updateDB(rep, storage)
			if err != nil {
				logger.Error(err)
			}
			time.Sleep(cfg.Settings.StoreInterval)
		}
	}()

	return storage
}

func getFromDB(rep metric.Repository) (metric.Storage, error) {
	var newStorage metric.Storage = metric.NewMemStorage()
	metrics, err := rep.FindAll(context.Background())
	if err != nil {
		return newStorage, err
	}
	data, err := json.Marshal(metrics)
	if err != nil {
		return newStorage, err
	}
	err = newStorage.ImportFromJSON(data)

	return newStorage, err
}

func updateDB(rep metric.Repository, storage metric.Storage) error {
	data := storage.ExportToJSON()
	var metrics []metric.MetricExport
	err := json.Unmarshal(data, &metrics)
	if err != nil {
		return err
	}

	for _, v := range metrics {
		err = rep.Upsert(context.Background(), v)
		if err != nil {
			return err
		}
	}

	return nil
}
