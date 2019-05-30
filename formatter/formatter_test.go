package formatter

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"gotest.tools/assert"

	"github.com/aleveille/lagrande/metric"
)

func TestAtlasSingleMetricFormat(t *testing.T) {
	// <metric path> <metric value> <metric timestamp>\n

	atlasFormatter := NewAtlasFormatter()

	metricName := "testValue"
	byteName := []byte(metricName)
	metricType := "gauge"
	byteType := []byte(metricType)
	tags := "nf.node=localhost"
	byteTags := atlasFormatter.FormatTags(&tags)

	staticMeta := metric.MetricStaticMetadata{Name: &byteName, Tags: byteTags, MetricType: &byteType}

	byteArrValue := []byte(fmt.Sprintf("%d", 42))
	timestamp := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).UnixNano()

	m := metric.Metric{Metadata: &staticMeta, Name: &byteName, Value: &byteArrValue, Timestamp: &timestamp}
	mArr := []*metric.Metric{&m}

	formattedMetric := atlasFormatter.FormatData(&mArr)

	var sb strings.Builder
	for _, bytePtr := range *formattedMetric {
		sb.WriteString(string(*bytePtr))
	}

	formattedString := sb.String()
	assert.Equal(t, formattedString, "{\"tags\":{\"nf.node\":\"localhost\"},\"metrics\":[{\"tags\":{\"name\":\"testValue\",\"atlas.dstype\":\"gauge\"},\"timestamp\":1257894000000000,\"value\":42}]}")
}

func TestAtlasTwoMetricsFormat(t *testing.T) {
	// <metric path> <metric value> <metric timestamp>\n

	atlasFormatter := NewAtlasFormatter()

	byteName1 := []byte("testValue1")
	byteName2 := []byte("testValue2")
	tags := "nf.node=localhost"
	byteTags := atlasFormatter.FormatTags(&tags)
	byteType := []byte("gauge")

	staticMeta1 := metric.MetricStaticMetadata{Name: &byteName1, Tags: byteTags, MetricType: &byteType}
	staticMeta2 := metric.MetricStaticMetadata{Name: &byteName2, Tags: byteTags, MetricType: &byteType}

	byteArrValue1 := []byte(fmt.Sprintf("%d", 42))
	byteArrValue2 := []byte(fmt.Sprintf("%d", 84))
	timestamp := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).UnixNano()

	m1 := metric.Metric{Metadata: &staticMeta1, Name: &byteName1, Value: &byteArrValue1, Timestamp: &timestamp}
	m2 := metric.Metric{Metadata: &staticMeta2, Name: &byteName2, Value: &byteArrValue2, Timestamp: &timestamp}
	mArr := []*metric.Metric{&m1, &m2}

	formattedMetric := atlasFormatter.FormatData(&mArr)

	var sb strings.Builder
	for _, bytePtr := range *formattedMetric {
		sb.WriteString(string(*bytePtr))
	}

	formattedString := sb.String()
	assert.Equal(t, formattedString, "{\"tags\":{\"nf.node\":\"localhost\"},\"metrics\":[{\"tags\":{\"name\":\"testValue1\",\"atlas.dstype\":\"gauge\"},\"timestamp\":1257894000000000,\"value\":42},{\"tags\":{\"name\":\"testValue2\",\"atlas.dstype\":\"gauge\"},\"timestamp\":1257894000000000,\"value\":84}]}")
}

func TestInfluxDBSingleMetricFormat(t *testing.T) {
	// <measurement>[,<tag-key>=<tag-value>...] <field-key>=<field-value>[,<field2-key>=<field2-value>...] [unix-nano-timestamp]

	influxdbFormatter := NewInfluxdbFormatter()

	byteName := []byte("testValue")
	tags := "tag1=value1,tag2=value2"
	byteTags := influxdbFormatter.FormatTags(&tags)
	byteType := []byte("gauge")

	staticMeta := metric.MetricStaticMetadata{Name: &byteName, Tags: byteTags, MetricType: &byteType}

	byteArrValue := []byte(fmt.Sprintf("%d", 42))
	timestamp := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).UnixNano()

	m := metric.Metric{Metadata: &staticMeta, Name: &byteName, Value: &byteArrValue, Timestamp: &timestamp}
	mArr := []*metric.Metric{&m}

	formattedMetric := influxdbFormatter.FormatData(&mArr)

	var sb strings.Builder
	for _, bytePtr := range *formattedMetric {
		sb.WriteString(string(*bytePtr))
	}

	formattedString := sb.String()
	assert.Equal(t, formattedString, "lagrande,tag1=value1,tag2=value2 testValue=42 1257894000000000000")
}

func TestInfluxDBTwoMetricsFormat(t *testing.T) {
	// <measurement>[,<tag-key>=<tag-value>...] <field-key>=<field-value>[,<field2-key>=<field2-value>...] [unix-nano-timestamp]

	influxdbFormatter := NewInfluxdbFormatter()

	byteName1 := []byte("testValue1")
	byteName2 := []byte("testValue2")
	tags := "tag1=value1,tag2=value2"
	byteTags := influxdbFormatter.FormatTags(&tags)
	byteType := []byte("gauge")

	staticMeta1 := metric.MetricStaticMetadata{Name: &byteName1, Tags: byteTags, MetricType: &byteType}
	staticMeta2 := metric.MetricStaticMetadata{Name: &byteName2, Tags: byteTags, MetricType: &byteType}

	byteArrValue1 := []byte(fmt.Sprintf("%d", 42))
	byteArrValue2 := []byte(fmt.Sprintf("%d", 84))
	timestamp := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).UnixNano()

	m1 := metric.Metric{Metadata: &staticMeta1, Name: &byteName1, Value: &byteArrValue1, Timestamp: &timestamp}
	m2 := metric.Metric{Metadata: &staticMeta2, Name: &byteName2, Value: &byteArrValue2, Timestamp: &timestamp}
	mArr := []*metric.Metric{&m1, &m2}

	formattedMetric := influxdbFormatter.FormatData(&mArr)

	var sb strings.Builder
	for _, bytePtr := range *formattedMetric {
		sb.WriteString(string(*bytePtr))
	}

	formattedString := sb.String()
	assert.Equal(t, formattedString, "lagrande,tag1=value1,tag2=value2 testValue1=42,testValue2=84 1257894000000000000")
}

func TestCarbonSingleMetricNoTagsFormat(t *testing.T) {
	// <metric path> <metric value> <metric timestamp>\n

	carbonFormatter := NewCarbonFormatter()

	byteName := []byte("testValue")
	tags := ""
	byteTags := carbonFormatter.FormatTags(&tags)
	byteType := []byte("gauge")

	staticMeta := metric.MetricStaticMetadata{Name: &byteName, Tags: byteTags, MetricType: &byteType}

	byteArrValue := []byte(fmt.Sprintf("%d", 42))
	timestamp := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).UnixNano()

	m := metric.Metric{Metadata: &staticMeta, Name: &byteName, Value: &byteArrValue, Timestamp: &timestamp}
	mArr := []*metric.Metric{&m}

	formattedMetric := carbonFormatter.FormatData(&mArr)

	var sb strings.Builder
	for _, bytePtr := range *formattedMetric {
		sb.WriteString(string(*bytePtr))
	}

	formattedString := sb.String()
	assert.Equal(t, formattedString, "testValue 42 1257894000\n")
}

func TestCarbonSingleMetricWithTagsFormat(t *testing.T) {
	// <metric path> <metric value> <metric timestamp>\n

	carbonFormatter := NewCarbonFormatter()

	byteName := []byte("testValue")
	tags := "tag1=value1,tag2=value2"
	byteTags := carbonFormatter.FormatTags(&tags)
	byteType := []byte("gauge")

	staticMeta := metric.MetricStaticMetadata{Name: &byteName, Tags: byteTags, MetricType: &byteType}

	byteArrValue := []byte(fmt.Sprintf("%d", 42))
	timestamp := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).UnixNano()

	m := metric.Metric{Metadata: &staticMeta, Name: &byteName, Value: &byteArrValue, Timestamp: &timestamp}
	mArr := []*metric.Metric{&m}

	formattedMetric := carbonFormatter.FormatData(&mArr)

	var sb strings.Builder
	for _, bytePtr := range *formattedMetric {
		sb.WriteString(string(*bytePtr))
	}

	formattedString := sb.String()
	assert.Equal(t, formattedString, "testValue;tag1=value1;tag2=value2 42 1257894000\n")
}

func TestCarbonSingleTwoMetricsWithTagsFormat(t *testing.T) {
	// <metric path> <metric value> <metric timestamp>\n

	carbonFormatter := NewCarbonFormatter()

	byteName1 := []byte("testValue1")
	byteName2 := []byte("testValue2")
	tags := "tag1=value1,tag2=value2"
	byteTags := carbonFormatter.FormatTags(&tags)
	byteType := []byte("gauge")

	staticMeta1 := metric.MetricStaticMetadata{Name: &byteName1, Tags: byteTags, MetricType: &byteType}
	staticMeta2 := metric.MetricStaticMetadata{Name: &byteName2, Tags: byteTags, MetricType: &byteType}

	byteArrValue1 := []byte(fmt.Sprintf("%d", 42))
	byteArrValue2 := []byte(fmt.Sprintf("%d", 84))
	timestamp := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).UnixNano()

	m1 := metric.Metric{Metadata: &staticMeta1, Name: &byteName1, Value: &byteArrValue1, Timestamp: &timestamp}
	m2 := metric.Metric{Metadata: &staticMeta2, Name: &byteName2, Value: &byteArrValue2, Timestamp: &timestamp}
	mArr := []*metric.Metric{&m1, &m2}

	formattedMetric := carbonFormatter.FormatData(&mArr)

	var sb strings.Builder
	for _, bytePtr := range *formattedMetric {
		sb.WriteString(string(*bytePtr))
	}

	formattedString := sb.String()
	assert.Equal(t, formattedString, "testValue1;tag1=value1;tag2=value2 42 1257894000\ntestValue2;tag1=value1;tag2=value2 84 1257894000\n")
}

func TestM3DBSingleMetricWithTagsFormat(t *testing.T) {
	// <metric path> <metric value> <metric timestamp>\n

	m3dbFormatter := NewM3DBFormatter()

	metricName := "testValue"
	byteName := []byte(metricName)
	metricType := "gauge"
	byteType := []byte(metricType)
	tags := "tag1=value1,tag2=value2"
	byteTags := m3dbFormatter.FormatTags(&tags)

	staticMeta := metric.MetricStaticMetadata{Name: &byteName, Tags: byteTags, MetricType: &byteType}

	byteArrValue := []byte(fmt.Sprintf("%d", 42))
	timestamp := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).UnixNano()

	m := metric.Metric{Metadata: &staticMeta, Name: &byteName, Value: &byteArrValue, Timestamp: &timestamp}
	mArr := []*metric.Metric{&m}

	formattedMetric := m3dbFormatter.FormatData(&mArr)

	var sb strings.Builder
	for _, bytePtr := range *formattedMetric {
		sb.WriteString(string(*bytePtr))
	}

	formattedString := sb.String()
	assert.Equal(t, formattedString, "{\"namespace\":\"default\",\"id\":\"foo\",\"tags\":[{\"name\":\"__name__\",\"value\":\"testValue\"},{\"name\":\"tag1\",\"value\":\"value1\"},{\"name\":\"tag2\",\"value\":\"value2\"}],\"datapoint\":{\"timestamp\":1257894000,\"value\":42}}")
}
