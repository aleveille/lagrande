package publisher

import "github.com/aleveille/lagrande/metric"

type Publisher interface {
	PublishBytes(bytes *[]*[]byte) error
	PublishMetrics(metrics *[]*metric.Metric) error
}
