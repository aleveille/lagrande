package metric

type MetricStaticMetadata struct {
	Name       *[]byte
	Tags       *[]byte
	MetricType *[]byte
}

type Metric struct {
	Metadata  *MetricStaticMetadata
	Name      *[]byte
	Tags      *[]byte
	Value     *[]byte
	Timestamp *int64
}
