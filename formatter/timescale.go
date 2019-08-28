package formatter

import (
	"fmt"

	"github.com/aleveille/lagrande/metric"
)

type timescale struct {
}

// NewTimescaleFormatter is a no-op formatter since the publisher use the Metric struct directly
func NewTimescaleFormatter() Formatter {
	return &timescale{}
}

// FormatData is no-op for Timescale since the publisher use the Metric struct directly
func (f *timescale) FormatData(metrics *[]*metric.Metric) *[]*[]byte {
	return nil
}

// Format a series of comma-delimited strings of key=value, as expected by the timescalePublisher to re-split them
func (f *timescale) FormatTags(tags *string) *[]byte {
	if *tags != "" {
		byteStr := []byte(fmt.Sprintf(",%s", *tags))
		return &byteStr
	}

	return &byteForEmpty
}
