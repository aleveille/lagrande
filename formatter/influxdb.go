package formatter

// references:
//  - https://graphite.readthedocs.io/en/latest/tags.html
//  - https://graphite.readthedocs.io/en/latest/feeding-Influxdb.html
//    - <metric path> <metric value> <metric timestamp>
//    - <metric path[;tag1=value1;tag2=value2;...]> <metric value> <metric timestamp>

import (
	"fmt"

	"github.com/aleveille/lagrande/metric"
)

type influxdb struct {
}

func NewInfluxdbFormatter() Formatter {
	return &influxdb{}
}

// Format according to InfluxDB Line Protocol:
// <measurement>[,<tag-key>=<tag-value>...] <field-key>=<field-value>[,<field2-key>=<field2-value>...] [unix-nano-timestamp]
func (f *influxdb) FormatData(metrics *[]*metric.Metric) *[]*[]byte {
	r := make([]*[]byte, 8*len(*metrics))

	for i, metric := range *metrics {
		r[(4 * i)] = metric.Name
		r[(4*i)+1] = metric.Tags
		r[(4*i)+2] = metric.Metadata.Tags
		r[(4*i)+3] = &byteForSpace
		r[(4*i)+4] = &bytesForValueEquals
		r[(4*i)+5] = (*metrics)[i].Value
		r[(4*i)+6] = &byteForSpace
		byteTs := []byte(fmt.Sprintf("%d", *((*metrics)[0].Timestamp)))
		r[(4*i)+7] = &byteTs
	}

	return &r
}

// Format a series of comma-delimited strings of key=value into InfluxDB tag format: ,<tag-key>=<tag-value>,...
// https://docs.influxdata.com/influxdb/v1.7/introduction/getting-started/
func (f *influxdb) FormatTags(tags *string) *[]byte {
	if *tags != "" {
		byteStr := []byte(fmt.Sprintf(",%s", *tags))
		return &byteStr
	}

	return &byteForEmpty
}
