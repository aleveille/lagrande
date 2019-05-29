package generator

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/aleveille/lagrande/formatter"
	"github.com/aleveille/lagrande/metric"
)

type floatRandom struct {
	sharedData *floatRandomSharedData
}

type floatRandomSharedData struct {
	metadata *metric.MetricStaticMetadata
	min      float64
	max      float64
	reset    bool

	cache     [][]byte
	formatter *formatter.Formatter
}

// Improvement, right now the cache assume an increment of +/-1. If the increment is different, the cache can be very sparse.
const floatRandomMaxCacheSize = 500000

// NewFloatRandomGenerator returns a struct compliant with the Generator interface
// You want to call this method once per config and then clone the generator using Clone() so that metadata and cache are shared for all workers
func NewFloatRandomGenerator(config CLIConfig, tags *[]byte, f *formatter.Formatter) (Generator, error) {
	confName := "random"
	confMin := 0.0
	confMax := 100.0

	for _, arg := range config.Args {
		kv := strings.Split(arg, ":")
		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		switch key {
		case "name":
			if len(value) == 0 {
				return nil, fmt.Errorf("Error parsing random float name '%s'", value)
			}
			confName = value
		case "min":
			v, err := strconv.ParseFloat(value, 10)
			if err != nil {
				return nil, fmt.Errorf("Error parsing random float min '%s'", value)
			}
			confMin = v
		case "max":
			v, err := strconv.ParseFloat(value, 10)
			if err != nil {
				return nil, fmt.Errorf("Error parsing random float max '%s'", value)
			}
			confMax = v
		}
	}

	if confMax < confMin {
		return nil, fmt.Errorf("Maximum '%f' cannot be inferior to minimum '%f'", confMax, confMin)
	}

	metricName := []byte(confName)
	metricType := []byte("gauge")

	staticMeta := &metric.MetricStaticMetadata{
		Name:       &metricName,
		Tags:       tags,
		MetricType: &metricType,
	}

	sharedData := &floatRandomSharedData{
		metadata:  staticMeta,
		min:       confMin,
		max:       confMax,
		formatter: f,
	}

	return &floatRandom{sharedData: sharedData}, nil
}

// Clone the current generator into a new struct with the current value for value and the same pointer for sharedData
func (g floatRandom) Clone(newName string) Generator {
	return &floatRandom{sharedData: g.sharedData}
}

// Return the name of the generator (as specificed on the command-line)
func (g *floatRandom) GetName() string {
	return string(*g.sharedData.metadata.Name)
}

// Return a human-readable description of the generator
func (g *floatRandom) ToString() string {
	return fmt.Sprintf("Random float generator (%s) between %f and %f", *g.sharedData.metadata.Name, g.sharedData.min, g.sharedData.max)
}

// Generates a metric struct with a value computed from the generator's rules
func (g *floatRandom) GenerateMetric() *metric.Metric {
	timestamp := time.Now().UnixNano()
	var retMetric *metric.Metric

	retMetric = &metric.Metric{
		Metadata:  g.sharedData.metadata,
		Value:     float64ToByteArrPtr(rand.Float64()*(g.sharedData.max-g.sharedData.min) + g.sharedData.min),
		Timestamp: &timestamp}

	return retMetric
}
