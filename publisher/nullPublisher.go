package publisher

import "github.com/aleveille/lagrande/metric"

type nullPublisher struct {
}

func NewNullPublisher(endpoint string) Publisher {
	return &nullPublisher{}
}

// PublishMetrics is no-op for nullPublisher
func (p *nullPublisher) PublishMetrics(metrics *[]*metric.Metric) error {
	return nil
}

// PublishBytes is no-op for nullPublisher
func (p *nullPublisher) PublishBytes(byteArrays *[]*[]byte) error {
	return nil
}
