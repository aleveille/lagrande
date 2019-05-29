package formatter

import (
	"github.com/aleveille/lagrande/metric"
)

// IDEA: Remove the Formatter type and instead have the formatter return a function.
//       Use first-class funcs and a func type instead of interface https://blog.learngoprogramming.com/go-functions-overview-anonymous-closures-higher-order-deferred-concurrent-6799008dde7b
//       eg: type TagFormatter func(tags *string) *[]byte
//       eg: type DataFormatter func(metrics *[]*metric.MetricHP) *[]*[]byte
//       the in (eg) Atlas:
type Formatter interface {
	//GetFormattedMetrics(metrics []generator.Metric) string

	FormatTags(tags *string) *[]byte

	// FormatMetrics returns an array of pointers that the publisher will have to go through in order to read and
	// send all the bytes relevant to the data at hand.
	// The Metric type itself is an aggregate of a few pointers (name, value, etc) so
	// The inner most array of bytes is a *part* of the formatted data itself (the name of the metric, it's value, etc)
	// and we pass along a pointer to that byte array in order to avoid copying the memory.
	// The outer most array is the aggregate of all those parts in order to have the full data formatted for the type
	// *[]byte("metric_path1"),*[]byte(" "),*[]byte("value"),*[]byte(" "),*[]byte("timestamp"),*[]byte("\n")
	FormatData(metrics *[]*metric.Metric) *[]*[]byte
}
