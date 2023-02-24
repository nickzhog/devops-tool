package redis

import (
	"fmt"

	"github.com/nickzhog/devops-tool/internal/server/metric"
)

func prepareKeyForMetric(m metric.Metric) string {
	return fmt.Sprintf("%s_%s", m.ID, m.MType)
}

func prepareKey(id, mtype string) string {
	return fmt.Sprintf("%s_%s", id, mtype)
}
