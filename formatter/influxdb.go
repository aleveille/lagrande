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
	r := make([]*[]byte, 4*len(*metrics)+4)

	fixedName := []byte("lagrande") //TODO The first name is the measurment name (Worker name or something?) actually
	r[0] = &fixedName
	r[1] = (*metrics)[0].Metadata.Tags //TODO Influx tags are per measurement, not per "metric"
	r[2] = &byteForSpace

	i := 0
	for ; i < len(*metrics); i++ {
		r[(4*i)+3] = (*metrics)[i].Metadata.Name
		r[(4*i)+3+1] = &byteForEqual
		r[(4*i)+3+2] = (*metrics)[i].Value
		r[(4*i)+3+3] = &byteForComma // The last comma will be overriden by a space
	}

	r[(4*i)+3-1] = &byteForSpace                                    // Overwrite the last comma
	byteTs := []byte(fmt.Sprintf("%d", *((*metrics)[0].Timestamp))) //TODO Assuming the timestamp of the first metric (InfluxDB timestamp are per measurement, not per "metric")
	r[(4*i)+3] = &byteTs

	return &r
}

// Format a series of comma-delimited strings of key=value into InfluxDB tag format: ,<tag-key>=<tag-value>,...
// https://docs.influxdata.com/influxdb/v1.7/introduction/getting-started/
func (f *influxdb) FormatTags(tags *string) *[]byte {
	byteStr := []byte(fmt.Sprintf(",%s", *tags))
	return &byteStr
}
