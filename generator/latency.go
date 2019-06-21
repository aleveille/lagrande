package generator

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aleveille/lagrande/formatter"
	"github.com/aleveille/lagrande/metric"

	rand "golang.org/x/exp/rand"
	distuv "gonum.org/v1/gonum/stat/distuv"
)

type latencyDistribution struct {
	name       *[]byte
	tags       *[]byte
	sharedData *latencyDistributionSharedData
}

type latencyDistributionSharedData struct {
	metadata *metric.MetricStaticMetadata
	min      float64
	max      float64
	distrib  distuv.Gamma

	formatter *formatter.Formatter
}

// Latency (aka: web requests response time) distribution is generally shaped like a "skewed" continuous
// distribution where most of the data is gathered around the mean but there's a tail on the right (higher latency, eg: p95, p99...)
// This can represented with a Gamma distribution with Alpha greater than 1 and typically less than ~4, and Beta lower than ~5
//
// If you want to customize the distribution, you can very very roughly think of:
//  - Alpha as the skewness. Data is gathered more around the left with lower values of Alpha and more to the right with higher values of Alpha.
//  - Beta as the control to how much the data is grouped or scattered. Data is more concentrated around the "peak" with lower values of Beta and more scattered (longer and bigger tail) with higher values of Beta.
//
//
//  For what it's worth, the default Gamma distribution created by this generator looks like this:
//      ___
//     .    .
//    .      .
//    .       .
//    .        .
//   .          .
//   .           .
//   .             .
//   .               .
//  .                   .
//  .                          .
//  .                                  .          .
//
// Where the X axis is the response time and the Y axis is the likelyhood of generating this response time

// If you want to vizualize Gamma distribution interactively: https://www.medcalc.org/manual/gamma_distribution_functions.php

// NewLatencyDistributionGenerator returns a struct compliant with the Generator interface
// You want to call this method once per config and then clone the generator using Clone() so that metadata and cache are shared for all workers
func NewLatencyDistributionGenerator(config CLIConfig, tags *[]byte, f *formatter.Formatter) (Generator, error) {
	confName := "random"
	confMin := 100.0
	confMax := 10000.0
	confAlpha := 1.5
	confBeta := 10.0

	for _, arg := range config.Args {
		kv := strings.Split(arg, ":")
		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		switch key {
		case "name":
			if len(value) == 0 {
				return nil, fmt.Errorf("Error parsing latency distribution name '%s'", value)
			}
			confName = value
		case "min":
			v, err := strconv.ParseFloat(value, 10)
			if err != nil {
				return nil, fmt.Errorf("Error parsing latency distribution min '%s'", value)
			}
			confMin = v
		case "max":
			v, err := strconv.ParseFloat(value, 10)
			if err != nil {
				return nil, fmt.Errorf("Error parsing latency distribution max '%s'", value)
			}
			confMax = v
		case "alpha":
			v, err := strconv.ParseFloat(value, 10)
			if err != nil {
				return nil, fmt.Errorf("Error parsing latency distribution alpha '%s'", value)
			}
			confAlpha = v
		case "beta":
			v, err := strconv.ParseFloat(value, 10)
			if err != nil {
				return nil, fmt.Errorf("Error parsing latency distribution beta '%s'", value)
			}
			confBeta = v
		}
	}

	if confMax < confMin {
		return nil, fmt.Errorf("Maximum '%f' cannot be inferior to minimum '%f'", confMax, confMin)
	}
	if confAlpha <= 0.0 {
		return nil, fmt.Errorf("Alpha must be a positive number")
	}
	if confBeta <= 0.0 {
		return nil, fmt.Errorf("Beta must be a positive number")
	}

	metricName := []byte(confName)
	metricType := []byte("gauge")

	staticMeta := &metric.MetricStaticMetadata{
		Name:       &metricName,
		Tags:       tags,
		MetricType: &metricType,
	}

	randSource := rand.NewSource(uint64(time.Now().UnixNano()))
	betaDistrib := distuv.Gamma{Alpha: confAlpha, Beta: confBeta, Src: randSource}

	sharedData := &latencyDistributionSharedData{
		metadata:  staticMeta,
		min:       confMin,
		max:       confMax,
		distrib:   betaDistrib,
		formatter: f,
	}

	return &latencyDistribution{sharedData: sharedData}, nil
}

// Clone the current generator into a new struct with the current value for value and the same pointer for sharedData
func (g latencyDistribution) Clone(newName string, specificTags *[]byte) Generator {
	newg := latencyDistribution{sharedData: g.sharedData}
	newNameBytes := []byte(newName)
	newg.name = &newNameBytes
	newg.tags = specificTags
	return &newg
}

// Return the name of the generator (as specificed on the command-line)
func (g *latencyDistribution) GetName() string {
	if g.name != nil {
		return string(*g.name)
	}
	return string(*g.sharedData.metadata.Name)
}

func (g *latencyDistribution) scale(num float64) float64 {
	return num*(g.sharedData.max-g.sharedData.min) + g.sharedData.min
}

// Return a human-readable description of the generators
func (g *latencyDistribution) ToString() string {
	return fmt.Sprintf("Latency distribution generator (%s) between %.4f and %.4f with a mean of %.4f, P50=%.4f, P95=%.4f and P99=%.4f", *g.sharedData.metadata.Name, g.sharedData.min, g.sharedData.max, g.scale(g.sharedData.distrib.Mean()), g.scale(g.sharedData.distrib.Quantile(0.5)), g.scale(g.sharedData.distrib.Quantile(0.95)), g.scale(g.sharedData.distrib.Quantile(0.99)))
}

// Generates a metric struct with a value computed from the generator's rules
func (g *latencyDistribution) GenerateMetric() *metric.Metric {
	value := g.scale(g.sharedData.distrib.Rand())

	timestamp := time.Now().UnixNano() / int64(time.Nanosecond)

	retMetric := &metric.Metric{
		Metadata:  g.sharedData.metadata,
		Name:      g.name,
		Value:     float64ToByteArrPtr(value),
		Tags:      g.tags,
		Timestamp: &timestamp,
	}

	return retMetric
}
