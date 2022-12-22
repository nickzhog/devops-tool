package db

import (
	"encoding/json"
	"os"
	"time"

	"github.com/nickzhog/practicum-metric/internal/server/config"
	"github.com/nickzhog/practicum-metric/internal/server/metric"
	"github.com/nickzhog/practicum-metric/pkg/logging"
)

func Connect(cfg *config.Config, logger *logging.Logger) metric.Storage {
	var storage metric.Storage = metric.NewMemStorage()
	if cfg.Settings.StoreFile == "" {
		return storage
	}

	if cfg.Settings.Restore {
		storage = getFromFile(cfg.Settings.StoreFile, logger)
	}

	file, err := os.OpenFile(cfg.Settings.StoreFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		logger.Fatal(err)
		return storage
	}

	go func() {
		for {
			time.Sleep(cfg.Settings.StoreInterval)
			err = updateFile(storage, file)
			if err != nil {
				logger.Error(err)
			}
		}
	}()

	return storage
}

func getFromFile(file string, logger *logging.Logger) *metric.MemStorage {
	newStorage := metric.NewMemStorage()

	data, err := os.ReadFile(file)
	if err != nil {
		logger.Error(err)
		return newStorage
	}
	var metrics []metric.MetricExport
	err = json.Unmarshal(data, &metrics)
	if err != nil {
		logger.Error(err)
		return newStorage
	}

	for _, v := range metrics {
		switch v.MType {
		case metric.CounterType:
			newStorage.UpdateCounter(v.ID, *v.Delta)
		case metric.GaugeType:
			newStorage.UpdateGauge(v.ID, *v.Value)
		}
	}

	return newStorage
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
