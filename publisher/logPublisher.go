package publisher

import (
	"bytes"
	"fmt"

	"github.com/aleveille/lagrande/metric"
)

type logPublisher struct {
}

func NewLogPublisher(endpoint string) Publisher {
	return &logPublisher{}
}

// PublishMetrics prints all the passed metrics to the logger
func (p *logPublisher) PublishMetrics(metrics *[]*metric.Metric) error {
	for _, m := range *metrics {
		fmt.Printf("%i %s[%s][%s]=%v", m.Timestamp, string(*m.Name), string(*m.Tags), string(*m.Tags), string(*m.Metadata.Tags), m.Value)
	}

	return nil
}

func (p *logPublisher) PublishBytes(byteArrays *[]*[]byte) error {
	//if log.GetLevel() == log.TraceLevel {
	bytesBuff := bytes.NewBuffer(nil)
	for _, bArr := range *byteArrays {
		bytesBuff.Write(*bArr)
	}
	fmt.Printf("%s", bytesBuff.String())
	//}

	return nil
}
