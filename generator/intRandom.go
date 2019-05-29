package generator

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/aleveille/lagrande/formatter"
	"github.com/aleveille/lagrande/metric"
)

type intRandom struct {
	sharedData *intRandomSharedData
}

type intRandomSharedData struct {
	metadata *metric.MetricStaticMetadata
	min      int
	max      int
	reset    bool

	cache     []*[]byte
	formatter *formatter.Formatter
}

const intRandomMaxCacheSize = 500000

// NewIntRandomGenerator returns a struct compliant with the Generator interface
// You want to call this method once per config and then clone the generator using Clone() so that metadata and cache are shared for all workers
func NewIntRandomGenerator(config CLIConfig, tags *[]byte, f *formatter.Formatter) (Generator, error) {
	confName := "answerToEverything"
	confMin := 0
	confMax := math.MaxInt32

	for _, arg := range config.Args {
		kv := strings.Split(arg, ":")
		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		switch key {
		case "name":
			if len(value) == 0 {
				return nil, fmt.Errorf("Error parsing random int name '%s'", value)
			}
			confName = value
		case "min":
			v, err := strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("Error parsing random int min '%s'", value)
			}
			confMin = v
		case "max":
			v, err := strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("Error parsing random int max '%s'", value)
			}
			confMax = v
		}
	}

	if confMax < confMin {
		return nil, fmt.Errorf("Maximum '%d' cannot be inferior to minimum '%d'", confMax, confMin)
	}

	metricName := []byte(confName)
	metricType := []byte("gauge")

	var cache []*[]byte
	cacheSize := confMax - confMin + 1
	if cacheSize <= intRandomMaxCacheSize && cacheSize > 0 {
		cache = make([]*[]byte, cacheSize)
	}

	staticMeta := &metric.MetricStaticMetadata{
		Name:       &metricName,
		Tags:       tags,
		MetricType: &metricType,
	}

	sharedData := &intRandomSharedData{
		metadata:  staticMeta,
		min:       confMin,
		max:       confMax,
		cache:     cache,
		formatter: f,
	}

	return &intRandom{sharedData: sharedData}, nil
}

// Clone the current generator into a new struct with the current value for value and the same pointer for sharedData
func (g intRandom) Clone(newName string) Generator {
	return &intRandom{sharedData: g.sharedData}
}

// Return the name of the generator (as specificed on the command-line)
func (g *intRandom) GetName() string {
	return string(*g.sharedData.metadata.Name)
}

// Return a human-readable description of the generator
func (g *intRandom) ToString() string {
	return fmt.Sprintf("Random int generator (%s) between %d and %d", g.sharedData.metadata.Name, g.sharedData.min, g.sharedData.max)
}

// Generates a metric struct with a value computed from the generator's rules
func (g *intRandom) GenerateMetric() *metric.Metric {
	timestamp := time.Now().UnixNano()
	var retMetric *metric.Metric

	randomInt := rand.Intn(g.sharedData.max-g.sharedData.min) + g.sharedData.min
	if g.sharedData.cache != nil {
		if g.sharedData.cache[randomInt-g.sharedData.min] == nil {
			g.sharedData.cache[randomInt-g.sharedData.min] = intToByteArrPtr(randomInt)
		}

		retMetric = &metric.Metric{
			Metadata:  g.sharedData.metadata,
			Value:     g.sharedData.cache[randomInt-g.sharedData.min],
			Timestamp: &timestamp}

		return retMetric
	}

	byteArr := []byte(fmt.Sprintf("%d", randomInt))
	retMetric = &metric.Metric{
		Metadata:  g.sharedData.metadata,
		Value:     &byteArr,
		Timestamp: &timestamp}

	return retMetric
}
