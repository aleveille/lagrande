package generator

import "github.com/aleveille/lagrande/metric"

// CLIConfig represents a set of 'key: value' pairs that are used to configure each generator
type CLIConfig struct {
	Args []string
}

// Generator is the common interface to all generators
type Generator interface {
	GenerateMetric() *metric.Metric
	Clone(newName string, specificTags *[]byte) Generator
	GetName() string
	ToString() string
}
