package generator

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aleveille/lagrande/formatter"
	"github.com/aleveille/lagrande/metric"
)

type floatCounter struct {
	name       *[]byte
	value      float64
	sharedData *floatCounterSharedData
}

type floatCounterSharedData struct {
	metadata  *metric.MetricStaticMetadata
	increment float64
	min       float64
	max       float64
	reset     bool

	cache     [][]byte
	formatter *formatter.Formatter
}

// NewFloatCounterGenerator returns a struct compliant with the Generator interface
// You want to call this method once per config and then clone the generator using Clone() so that metadata and cache are shared for all workers
func NewFloatCounterGenerator(config CLIConfig, tags *[]byte, f *formatter.Formatter) (Generator, error) {
	confName := "pi"
	confValue := 3.14
	confIncrement := float64(1.0)
	confMin := float64(0)
	confMax := float64(100000000000000.0)
	confReset := true

	for _, arg := range config.Args {
		kv := strings.Split(arg, ":")
		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		switch key {
		case "name":
			if len(value) == 0 {
				return nil, fmt.Errorf("Error parsing float counter name '%s'", value)
			}
			confName = value
		case "value":
			v, err := strconv.ParseFloat(value, 10)
			if err != nil {
				return nil, fmt.Errorf("Error parsing float counter value '%s'", value)
			}
			confValue = v
		case "increment":
			v, err := strconv.ParseFloat(value, 10)
			if err != nil {
				return nil, fmt.Errorf("Error parsing float counter increment '%s'", value)
			}
			confIncrement = v
		case "min":
			v, err := strconv.ParseFloat(value, 10)
			if err != nil {
				return nil, fmt.Errorf("Error parsing float counter min '%s'", value)
			}
			confMin = v
		case "max":
			v, err := strconv.ParseFloat(value, 10)
			if err != nil {
				return nil, fmt.Errorf("Error parsing float counter max '%s'", value)
			}
			confMax = v
		case "reset":
			v, err := strconv.ParseBool(value)
			if err != nil {
				return nil, fmt.Errorf("Error parsing float counter reset '%s'", value)
			}
			confReset = v
		}
	}

	if confMax < confMin {
		return nil, fmt.Errorf("Maximum '%f' cannot be inferior to minimum '%f'", confMax, confMin)
	}

	if confIncrement == 0 {
		confMin = confValue
		confMax = confValue
	}

	metricName := []byte(confName)
	metricType := []byte("counter")

	staticMeta := &metric.MetricStaticMetadata{
		Name:       &metricName,
		Tags:       tags,
		MetricType: &metricType,
	}

	sharedData := &floatCounterSharedData{
		metadata:  staticMeta,
		increment: confIncrement,
		min:       confMin,
		max:       confMax,
		reset:     confReset,
		formatter: f,
	}

	return &floatCounter{value: confValue, sharedData: sharedData}, nil
}

// Clone the current generator into a new struct with the current value for value and the same pointer for sharedData
func (g floatCounter) Clone(newName string) Generator {
	newg := floatCounter{value: g.value, sharedData: g.sharedData}
	newNameBytes := []byte(newName)
	newg.name = &newNameBytes
	return &newg
}

// Return the name of the generator (as specificed on the command-line)
func (g *floatCounter) GetName() string {
	if g.name != nil {
		return string(*g.name)
	}
	return string(*g.sharedData.metadata.Name)
}

// Return a human-readable description of the generators
func (g *floatCounter) ToString() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Float counter generator (%s)", *g.sharedData.metadata.Name))
	if g.sharedData.increment != 0 {
		sb.WriteString(fmt.Sprintf(" of initial value %.4f with increments of %.4f", g.value, g.sharedData.increment))
		if g.sharedData.increment > 0 {
			sb.WriteString(fmt.Sprintf(" up to a maximum of %.4f", g.sharedData.max))
			if g.sharedData.reset {
				sb.WriteString(fmt.Sprintf(", it will then reset to %.4f.", g.sharedData.min))
			}
		} else {
			sb.WriteString(fmt.Sprintf(" up to a minimum of %.4f", g.sharedData.min))
			if g.sharedData.reset {
				sb.WriteString(fmt.Sprintf(", it will then reset to %.4f.", g.sharedData.max))
			}
		}
	} else {
		sb.WriteString(fmt.Sprintf(" with a value of %.4f", g.value))
	}

	return sb.String()
}

// Generates a metric struct with a value computed from the generator's rules
func (g *floatCounter) GenerateMetric() *metric.Metric {
	timestamp := time.Now().UnixNano()
	var retMetric *metric.Metric

	retMetric = &metric.Metric{
		Metadata:  g.sharedData.metadata,
		Name:      g.name,
		Value:     float64ToByteArrPtr(g.value),
		Timestamp: &timestamp,
	}

	if g.sharedData.increment != 0.0 {
		g.value += g.sharedData.increment
		if g.value > g.sharedData.max || g.value < g.sharedData.min {
			if g.sharedData.increment > 0 {
				if g.sharedData.reset {
					g.value = g.sharedData.min
				} else {
					g.value = g.sharedData.max
				}
			} else {
				if g.sharedData.reset {
					g.value = g.sharedData.max
				} else {
					g.value = g.sharedData.min
				}
			}
		}
	}

	return retMetric
}
