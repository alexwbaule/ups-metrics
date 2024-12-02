package writer

import (
	"context"
	"github.com/alexwbaule/ups-metrics/internal/domain/entity/device"
)

type WriteMetric interface {
	Write(ctx context.Context, metric device.Metric) error
}
