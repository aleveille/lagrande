package generator

import (
	"testing"

	"gotest.tools/assert"

	"github.com/aleveille/lagrande/metric"
)

func TestCounterStaticInt(t *testing.T) {
	config := CLIConfig{Args: []string{"value: 99", "increment: 0"}}
	gen, err := NewIntCounterGenerator(config, nil, nil)

	assert.NilError(t, err)

	metric := gen.GenerateMetric()
	assert.Equal(t, "99", string(*metric.Value))

	metric = gen.GenerateMetric()
	assert.Equal(t, "99", string(*metric.Value))

	metric = gen.GenerateMetric()
	assert.Equal(t, "99", string(*metric.Value))
}

func TestCounterIncrementInt(t *testing.T) {
	config := CLIConfig{Args: []string{"value: 99", "increment: 1"}}
	gen, err := NewIntCounterGenerator(config, nil, nil)

	assert.NilError(t, err)

	metric := gen.GenerateMetric()
	assert.Equal(t, "99", string(*metric.Value))

	metric = gen.GenerateMetric()
	assert.Equal(t, "100", string(*metric.Value))

	metric = gen.GenerateMetric()
	assert.Equal(t, "101", string(*metric.Value))
}

func TestCounterIncrementNoCache(t *testing.T) {
	config := CLIConfig{Args: []string{"value: 0", "min: 0", "max: 100000000", "increment: 50000000"}}
	gen, err := NewIntCounterGenerator(config, nil, nil)

	assert.NilError(t, err)

	metric := gen.GenerateMetric()
	assert.Equal(t, "0", string(*metric.Value))

	metric = gen.GenerateMetric()
	assert.Equal(t, "50000000", string(*metric.Value))

	metric = gen.GenerateMetric()
	assert.Equal(t, "100000000", string(*metric.Value))
}

func TestCounterDecrementInt(t *testing.T) {
	config := CLIConfig{Args: []string{"value: 99", "increment: -1"}}
	gen, err := NewIntCounterGenerator(config, nil, nil)

	assert.NilError(t, err)

	metric := gen.GenerateMetric()
	assert.Equal(t, "99", string(*metric.Value))

	metric = gen.GenerateMetric()
	assert.Equal(t, "98", string(*metric.Value))

	metric = gen.GenerateMetric()
	assert.Equal(t, "97", string(*metric.Value))
}

func TestCounterIncrementIntOverflow(t *testing.T) {
	config := CLIConfig{Args: []string{"value: 2147483647", "increment: 1"}}
	gen, err := NewIntCounterGenerator(config, nil, nil)

	assert.NilError(t, err)

	metric := gen.GenerateMetric()
	assert.Equal(t, "2147483647", string(*metric.Value))

	metric = gen.GenerateMetric()
	assert.Equal(t, "0", string(*metric.Value))

	metric = gen.GenerateMetric()
	assert.Equal(t, "1", string(*metric.Value))
}

func TestCounterDecrementIntOverflow(t *testing.T) {
	config := CLIConfig{Args: []string{"value: -2147483648", "increment: -1", "min: -2147483648", "max: 0"}}
	gen, err := NewIntCounterGenerator(config, nil, nil)

	assert.NilError(t, err)

	metric := gen.GenerateMetric()
	assert.Equal(t, "-2147483648", string(*metric.Value))

	metric = gen.GenerateMetric()
	assert.Equal(t, "0", string(*metric.Value))

	metric = gen.GenerateMetric()
	assert.Equal(t, "-1", string(*metric.Value))
}

func TestCounterIncrementIntReset(t *testing.T) {
	config := CLIConfig{Args: []string{"value: 1", "increment: 2", "min: 0", "max: 5"}}
	gen, err := NewIntCounterGenerator(config, nil, nil)

	assert.NilError(t, err)

	metric := gen.GenerateMetric()
	assert.Equal(t, "1", string(*metric.Value))

	metric = gen.GenerateMetric()
	assert.Equal(t, "3", string(*metric.Value))

	metric = gen.GenerateMetric()
	assert.Equal(t, "5", string(*metric.Value))

	metric = gen.GenerateMetric()
	assert.Equal(t, "0", string(*metric.Value))

	metric = gen.GenerateMetric()
	assert.Equal(t, "2", string(*metric.Value))
}

func TestCounterDecrementIntReset(t *testing.T) {
	config := CLIConfig{Args: []string{"value: 6", "increment: -3", "min: 0", "max: 10"}}
	gen, err := NewIntCounterGenerator(config, nil, nil)

	assert.NilError(t, err)

	metric := gen.GenerateMetric()
	assert.Equal(t, "6", string(*metric.Value))

	metric = gen.GenerateMetric()
	assert.Equal(t, "3", string(*metric.Value))

	metric = gen.GenerateMetric()
	assert.Equal(t, "0", string(*metric.Value))

	metric = gen.GenerateMetric()
	assert.Equal(t, "10", string(*metric.Value))

	metric = gen.GenerateMetric()
	assert.Equal(t, "7", string(*metric.Value))
}

func TestCounterIncrementIntMax(t *testing.T) {
	config := CLIConfig{Args: []string{"value: 1", "increment: 2", "max: 5", "reset: false"}}
	gen, err := NewIntCounterGenerator(config, nil, nil)

	assert.NilError(t, err)

	metric := gen.GenerateMetric()
	assert.Equal(t, "1", string(*metric.Value))

	metric = gen.GenerateMetric()
	assert.Equal(t, "3", string(*metric.Value))

	metric = gen.GenerateMetric()
	assert.Equal(t, "5", string(*metric.Value))

	metric = gen.GenerateMetric()
	assert.Equal(t, "5", string(*metric.Value))
}

func TestCounterDecrementIntMin(t *testing.T) {
	config := CLIConfig{Args: []string{"value: 6", "increment: -3", "min: 0", "reset: false"}}
	gen, err := NewIntCounterGenerator(config, nil, nil)

	assert.NilError(t, err)

	metric := gen.GenerateMetric()
	assert.Equal(t, "6", string(*metric.Value))

	metric = gen.GenerateMetric()
	assert.Equal(t, "3", string(*metric.Value))

	metric = gen.GenerateMetric()
	assert.Equal(t, "0", string(*metric.Value))

	metric = gen.GenerateMetric()
	assert.Equal(t, "0", string(*metric.Value))
}

func TestCounterStaticFloat(t *testing.T) {
	config := CLIConfig{Args: []string{"value: 1.1", "increment: 0"}}
	gen, err := NewFloatCounterGenerator(config, nil, nil)

	assert.NilError(t, err)

	metric := gen.GenerateMetric()
	assert.Equal(t, "1.1000", string(*metric.Value))

	metric = gen.GenerateMetric()
	assert.Equal(t, "1.1000", string(*metric.Value))

	metric = gen.GenerateMetric()
	assert.Equal(t, "1.1000", string(*metric.Value))
}

func TestCounterIncrementFloat(t *testing.T) {
	config := CLIConfig{Args: []string{"value: 50.0", "increment: 2.5"}}
	gen, err := NewFloatCounterGenerator(config, nil, nil)

	assert.NilError(t, err)

	metric := gen.GenerateMetric()
	assert.Equal(t, "50.0000", string(*metric.Value))

	metric = gen.GenerateMetric()
	assert.Equal(t, "52.5000", string(*metric.Value))

	metric = gen.GenerateMetric()
	assert.Equal(t, "55.0000", string(*metric.Value))
}

func TestCounterIncrementVerySmallFloat(t *testing.T) {
	config := CLIConfig{Args: []string{"value: -0.00000015", "min: -1000", "increment: -2.5"}}
	gen, err := NewFloatCounterGenerator(config, nil, nil)

	assert.NilError(t, err)

	metric := gen.GenerateMetric()
	assert.Equal(t, "-0.0000", string(*metric.Value))

	metric = gen.GenerateMetric()
	assert.Equal(t, "-2.5000", string(*metric.Value))

	metric = gen.GenerateMetric()
	assert.Equal(t, "-5.0000", string(*metric.Value))
}

func TestCounterDecrementFloat(t *testing.T) {
	config := CLIConfig{Args: []string{"value: 25", "increment: -2.5"}}
	gen, err := NewFloatCounterGenerator(config, nil, nil)

	assert.NilError(t, err)

	metric := gen.GenerateMetric()
	assert.Equal(t, "25.0000", string(*metric.Value))

	metric = gen.GenerateMetric()
	assert.Equal(t, "22.5000", string(*metric.Value))

	metric = gen.GenerateMetric()
	assert.Equal(t, "20.0000", string(*metric.Value))
}

func TestCounterIncrementFloatReset(t *testing.T) {
	config := CLIConfig{Args: []string{"value: 1.0", "increment: 2.5", "min: 0", "max: 6"}}
	gen, err := NewFloatCounterGenerator(config, nil, nil)

	assert.NilError(t, err)

	metric := gen.GenerateMetric()
	assert.Equal(t, "1.0000", string(*metric.Value))

	metric = gen.GenerateMetric()
	assert.Equal(t, "3.5000", string(*metric.Value))

	metric = gen.GenerateMetric()
	assert.Equal(t, "6.0000", string(*metric.Value))

	metric = gen.GenerateMetric()
	assert.Equal(t, "0.0000", string(*metric.Value))

	metric = gen.GenerateMetric()
	assert.Equal(t, "2.5000", string(*metric.Value))
}

func TestCounterDecrementFloatReset(t *testing.T) {
	config := CLIConfig{Args: []string{"value: 6.0", "increment: -2.5", "min: 1", "max: 10"}}
	gen, err := NewFloatCounterGenerator(config, nil, nil)

	assert.NilError(t, err)

	metric := gen.GenerateMetric()
	assert.Equal(t, "6.0000", string(*metric.Value))

	metric = gen.GenerateMetric()
	assert.Equal(t, "3.5000", string(*metric.Value))

	metric = gen.GenerateMetric()
	assert.Equal(t, "1.0000", string(*metric.Value))

	metric = gen.GenerateMetric()
	assert.Equal(t, "10.0000", string(*metric.Value))

	metric = gen.GenerateMetric()
	assert.Equal(t, "7.5000", string(*metric.Value))
}

func TestCounterIncrementFloatMax(t *testing.T) {
	config := CLIConfig{Args: []string{"value: 1.0", "increment: 2.5", "max: 6", "reset: false"}}
	gen, err := NewFloatCounterGenerator(config, nil, nil)

	assert.NilError(t, err)

	metric := gen.GenerateMetric()
	assert.Equal(t, "1.0000", string(*metric.Value))

	metric = gen.GenerateMetric()
	assert.Equal(t, "3.5000", string(*metric.Value))

	metric = gen.GenerateMetric()
	assert.Equal(t, "6.0000", string(*metric.Value))

	metric = gen.GenerateMetric()
	assert.Equal(t, "6.0000", string(*metric.Value))
}

func TestCounterDecrementFloatMin(t *testing.T) {
	config := CLIConfig{Args: []string{"value: 6.0", "increment: -2.5", "min: 1", "reset: false"}}
	gen, err := NewFloatCounterGenerator(config, nil, nil)

	assert.NilError(t, err)

	metric := gen.GenerateMetric()
	assert.Equal(t, "6.0000", string(*metric.Value))

	metric = gen.GenerateMetric()
	assert.Equal(t, "3.5000", string(*metric.Value))

	metric = gen.GenerateMetric()
	assert.Equal(t, "1.0000", string(*metric.Value))

	metric = gen.GenerateMetric()
	assert.Equal(t, "1.0000", string(*metric.Value))
}

var result *metric.Metric // https://dave.cheney.net/2013/06/30/how-to-write-benchmarks-in-go

func BenchmarkCounterStaticInt(b *testing.B) {
	var r *metric.Metric
	config := CLIConfig{Args: []string{"value: 42", "increment: 0"}}
	gen, _ := NewIntCounterGenerator(config, nil, nil)

	for n := 0; n < b.N; n++ {
		r = gen.GenerateMetric()
	}
	result = r
}

func BenchmarkCounterIncrementInt(b *testing.B) {
	config := CLIConfig{Args: []string{"value: 100", "increment: 5"}}
	gen, _ := NewIntCounterGenerator(config, nil, nil)

	for n := 0; n < b.N; n++ {
		gen.GenerateMetric()
	}
}

func BenchmarkCounterRandomInt(b *testing.B) {
	config := CLIConfig{Args: []string{"min: 0", "max: 100"}}
	gen, _ := NewIntRandomGenerator(config, nil, nil)

	for n := 0; n < b.N; n++ {
		gen.GenerateMetric()
	}
}

func BenchmarkCounterStaticFloat(b *testing.B) {
	config := CLIConfig{Args: []string{"value: 42", "increment: 0"}}
	gen, _ := NewFloatCounterGenerator(config, nil, nil)

	for n := 0; n < b.N; n++ {
		gen.GenerateMetric()
	}
}

func BenchmarkCounterIncrementFloat(b *testing.B) {
	config := CLIConfig{Args: []string{"value: 100", "increment: 5"}}
	gen, _ := NewFloatCounterGenerator(config, nil, nil)

	for n := 0; n < b.N; n++ {
		gen.GenerateMetric()
	}
}

func BenchmarkCounterRandomFloat(b *testing.B) {
	config := CLIConfig{Args: []string{"min: 0", "max: 100"}}
	gen, _ := NewFloatRandomGenerator(config, nil, nil)

	for n := 0; n < b.N; n++ {
		gen.GenerateMetric()
	}
}

func BenchmarkLatency(b *testing.B) {
	config := CLIConfig{Args: []string{"min: 0", "max: 100"}}
	gen, _ := NewLatencyDistributionGenerator(config, nil, nil)

	for n := 0; n < b.N; n++ {
		gen.GenerateMetric()
	}
}
