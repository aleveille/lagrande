package generator

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/aleveille/lagrande/formatter"
	"github.com/aleveille/lagrande/metric"
)

type intCounter struct {
	name       *[]byte
	value      int
	tags       *[]byte
	sharedData *intCounterSharedData
}

type intCounterSharedData struct {
	metadata  *metric.MetricStaticMetadata
	increment int
	min       int
	max       int
	reset     bool

	cache     []*[]byte
	formatter *formatter.Formatter
}

// Improvement, right now the cache assume an increment of +/-1. If the increment is different, the cache can be very sparse.
const intCounterMaxCacheSize = 500000

// NewIntCounterGenerator returns a struct compliant with the Generator interface
// You want to call this method once per config and then clone the generator using Clone() so that metadata and cache are shared for all workers
func NewIntCounterGenerator(config CLIConfig, tags *[]byte, f *formatter.Formatter) (Generator, error) {
	confName := "answerToEverything"
	confValue := 42
	confIncrement := 1
	confMin := 0
	confMax := math.MaxInt32
	confReset := true

	for _, arg := range config.Args {
		kv := strings.Split(arg, ":")
		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		switch key {
		case "name":
			if len(value) == 0 {
				return nil, fmt.Errorf("Error parsing int counter name '%s'", value)
			}
			confName = value
		case "value":
			v, err := strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("Error parsing int counter value '%s'", value)
			}
			confValue = v
		case "increment":
			v, err := strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("Error parsing int counter increment '%s'", value)
			}
			confIncrement = v
		case "min":
			v, err := strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("Error parsing int counter min '%s'", value)
			}
			confMin = v
		case "max":
			v, err := strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("Error parsing int counter max '%s'", value)
			}
			confMax = v
		case "reset":
			v, err := strconv.ParseBool(value)
			if err != nil {
				return nil, fmt.Errorf("Error parsing int counter reset '%s'", value)
			}
			confReset = v
		}
	}

	if confMax < confMin {
		return nil, fmt.Errorf("Maximum '%d' cannot be inferior to minimum '%d'", confMax, confMin)
	}

	if confIncrement == 0 {
		confMin = confValue
		confMax = confValue
	}

	metricName := []byte(confName)
	metricType := []byte("counter")

	var cache []*[]byte
	cacheSize := confMax - confMin + 1
	if cacheSize <= intCounterMaxCacheSize && cacheSize > 0 {
		cache = make([]*[]byte, cacheSize)
		if confIncrement == 0 {
			cache[0] = intToByteArrPtr(confValue)
		}
	}

	staticMeta := &metric.MetricStaticMetadata{
		Name:       &metricName,
		Tags:       tags,
		MetricType: &metricType,
	}

	sharedData := &intCounterSharedData{
		metadata:  staticMeta,
		increment: confIncrement,
		min:       confMin,
		max:       confMax,
		reset:     confReset,
		cache:     cache,
		formatter: f,
	}

	return &intCounter{value: confValue, sharedData: sharedData}, nil
}

// Clone the current generator into a new struct with the current value for value and the same pointer for sharedData
func (g intCounter) Clone(newName string, specificTags *[]byte) Generator {
	newg := intCounter{value: g.value, sharedData: g.sharedData}
	newNameBytes := []byte(newName)
	newg.name = &newNameBytes
	newg.tags = specificTags
	return &newg
}

// Return the name of the generator (as specificed on the command-line)
func (g *intCounter) GetName() string {
	if g.name != nil {
		return string(*g.name)
	}
	return string(*g.sharedData.metadata.Name)
}

// Return a human-readable description of the generator
func (g *intCounter) ToString() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Int counter generator (%s)", *g.sharedData.metadata.Name))
	if g.sharedData.increment != 0 {
		sb.WriteString(fmt.Sprintf(" of initial value %d with increments of %d", g.value, g.sharedData.increment))
		if g.sharedData.increment > 0 {
			sb.WriteString(fmt.Sprintf(" up to a maximum of %d", g.sharedData.max))
			if g.sharedData.reset {
				sb.WriteString(fmt.Sprintf(", it will then reset to %d.", g.sharedData.min))
			}
		} else {
			sb.WriteString(fmt.Sprintf(" up to a minimum of %d", g.sharedData.min))
			if g.sharedData.reset {
				sb.WriteString(fmt.Sprintf(", it will then reset to %d.", g.sharedData.max))
			}
		}
	} else {
		sb.WriteString(fmt.Sprintf(" with a value of %d", g.value))
	}

	return sb.String()
}

// Generates a metric struct with a value computed from the generator's rules
func (g *intCounter) GenerateMetric() *metric.Metric {
	timestamp := time.Now().UnixNano()
	var retMetric *metric.Metric

	if g.sharedData.increment == 0 { // Static value, use cache[0] right away
		retMetric = &metric.Metric{
			Metadata:  g.sharedData.metadata,
			Name:      g.name,
			Value:     g.sharedData.cache[0],
			Tags:      g.tags,
			Timestamp: &timestamp,
		}
	} else { // If counter value, first check if the cache was initialized
		if g.sharedData.cache != nil {
			// It the cache is initialized, check if the current value is in cache. If not, add it
			if g.sharedData.cache[(g.value-g.sharedData.min)] == nil {
				g.sharedData.cache[(g.value - g.sharedData.min)] = intToByteArrPtr(g.value)
			}

			retMetric = &metric.Metric{
				Metadata:  g.sharedData.metadata,
				Name:      g.name,
				Value:     g.sharedData.cache[(g.value - g.sharedData.min)],
				Tags:      g.tags,
				Timestamp: &timestamp,
			}
		} else { // If the cache is not initialize, create a metric without cache
			byteArr := intToByteArrPtr(g.value)
			retMetric = &metric.Metric{
				Metadata:  g.sharedData.metadata,
				Name:      g.name,
				Value:     byteArr,
				Tags:      g.tags,
				Timestamp: &timestamp}
		}

		g.value += g.sharedData.increment
		if g.value > g.sharedData.max || g.value < g.sharedData.min {
			if g.sharedData.increment > 0 {
				if g.sharedData.reset {
					g.value = g.sharedData.min
				} else {
					// bug: or "limitation" if multiple workers are using the same generator, the first worker to attain min/max will "set" the single-value cache
					// for all workers so all of them will return the max value as soon as the first worker does
					g.value = g.sharedData.max // If we hit the max and we'll stay there, change to 'increment == 0' mode
					g.sharedData.increment = 0
					g.sharedData.cache = make([]*[]byte, 1)
					g.sharedData.cache[0] = intToByteArrPtr(g.value)
				}
			} else {
				if g.sharedData.reset {
					g.value = g.sharedData.max
				} else {
					g.value = g.sharedData.min // If we hit the min and we'll stay there, change to 'increment == 0' mode
					g.sharedData.increment = 0
					g.sharedData.cache = make([]*[]byte, 1)
					g.sharedData.cache[0] = intToByteArrPtr(g.value)
				}
			}
		}
	}

	return retMetric
}
