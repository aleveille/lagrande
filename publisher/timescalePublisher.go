// Source: Telegraf PostgreSQL output plugin
// https://github.com/influxdata/telegraf/pull/3428/files#diff-4a1e24356fcc6b51245dd5a1d1cc6779
// Local source checked out: https://github.com/svenklemm/telegraf/tree/postgres branch postgres
//
// Authors:
//   * Sven Klemm https://github.com/svenklemm
//   * Blagoj Atanasovski https://github.com/blagojts
//   * Oskari Saarenmaa https://github.com/saaros
//   * Rauli Ikonen https://github.com/rikonen
//
// License : https://github.com/influxdata/telegraf/blob/a8ff6488fb30f3801b7ada4316a119debd0f58f8/LICENSE

package publisher

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aleveille/lagrande/metric"
	tfmetric "github.com/aleveille/telegraf/metric"
	"github.com/influxdata/telegraf"

	"github.com/aleveille/telegraf/plugins/outputs/postgresql"
)

type timescalePublisher struct {
	postgresqlClient *postgresql.Postgresql
	endpoint         string
}

// The next functions (NewTimescalePublisher, PublishBytes, PublishMetrics and mapponizeTags) are my own & serve to wrap the functionnality of the PostgreSQL Telegraf plugin
func NewTimescalePublisher(endpoint string) Publisher {
	p := postgresql.NewPostgresql(endpoint)
	p.Connect()
	return &timescalePublisher{postgresqlClient: p, endpoint: endpoint}
}

// PublishBytes is unimplemented for timescalePublisher
func (p *timescalePublisher) PublishBytes(byteArrays *[]*[]byte) error {
	return errors.New("PublishMetrics is not supported for HTTP publisher yet")
}

// PublishMetrics for Timescale re-wraps our Metric into telegraf/metric and then use the Write of the telegraf plugin
func (p *timescalePublisher) PublishMetrics(metrics *[]*metric.Metric) error {
	tfMetrics := make([]telegraf.Metric, len(*metrics))

	for i, m := range *metrics {
		fields := make(map[string]interface{})

		stringVal := string(*m.Value)
		intVal, err := strconv.Atoi(stringVal)
		if err == nil {
			fields["value"] = intVal
		} else {
			floatVal, err := strconv.ParseFloat(stringVal, 10)
			if err == nil {
				fields["value"] = floatVal
			} else {
				fmt.Println("Warning: couldn't parse metric value for Timescale")
				fields["value"] = stringVal
			}
		}

		var metricType telegraf.ValueType
		switch string(*m.Metadata.MetricType) {
		case "counter":
			metricType = telegraf.Counter
		case "histogram":
			metricType = telegraf.Histogram
		case "summary":
			metricType = telegraf.Summary
		default:
			metricType = telegraf.Gauge
		}

		msTimestamp := *(m.Timestamp) / 1000 / 1000 / 1000

		tfMetric, err := tfmetric.New(
			string(*m.Metadata.Name),
			mapponizeTags(m.Tags, m.Metadata.Tags),
			fields,
			time.Unix(msTimestamp, 0),
			metricType,
		)

		if err != nil {
			return err
		}
		tfMetrics[i] = tfMetric
	}

	return p.postgresqlClient.Write(tfMetrics)
}

func mapponizeTags(sharedTags *[]byte, tags *[]byte) map[string]string {
	tagsMap := make(map[string]string)

	for _, tagsSlice := range [2]*[]byte{sharedTags, tags} {
		if tagsSlice != nil {
			stringTags := string(*tagsSlice)
			tagsSlice := strings.Split(stringTags, ",")
			for _, tag := range tagsSlice {
				tagKeyVal := strings.Split(tag, "=")
				if len(tagKeyVal) >= 1 && tagKeyVal[0] != "" {
					tagVal := ""
					if len(tagKeyVal) >= 2 {
						tagVal = tagKeyVal[1]
					}
					tagsMap[tagKeyVal[0]] = tagVal
				}
			}
		}
	}

	return tagsMap
}
