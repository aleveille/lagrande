package metric

type MetricStaticMetadata struct {
	Name       *[]byte
	Tags       *[]byte
	MetricType *[]byte
}

type Metric struct {
	Metadata  *MetricStaticMetadata
	Tags      *[]byte
	Value     *[]byte
	Timestamp *int64
}
