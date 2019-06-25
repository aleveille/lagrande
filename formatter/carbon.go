package formatter

// references:
//  - https://graphite.readthedocs.io/en/latest/tags.html
//  - https://graphite.readthedocs.io/en/latest/feeding-carbon.html
//    - <metric path> <metric value> <metric timestamp>
//    - <metric path[;tag1=value1;tag2=value2;...]> <metric value> <metric timestamp>

import (
	"fmt"
	"strings"

	"github.com/aleveille/lagrande/metric"
)

type carbon struct {
}

func NewCarbonFormatter() Formatter {
	return &carbon{}
}

// Format according to Carbon protocol:
// metric_path value timestamp\n
func (f *carbon) FormatData(metrics *[]*metric.Metric) *[]*[]byte {
	r := make([]*[]byte, 8*len(*metrics))

	for i, m := range *metrics {
		if len(*metrics) > 1 {
			fmt.Println("More than one metric at a time")
		}
		byteTs := []byte(fmt.Sprintf("%d", *(m.Timestamp)/1000/1000/1000))

		r[(8 * i)] = m.Name
		r[(8*i)+1] = m.Metadata.Tags
		r[(8*i)+2] = m.Tags
		r[(8*i)+3] = &byteForSpace
		r[(8*i)+4] = m.Value
		r[(8*i)+5] = &byteForSpace
		r[(8*i)+6] = &byteTs
		r[(8*i)+7] = &byteForLineReturn
	}

	return &r
}

// Format a series of comma-delimited strings of key=value into Carbon (Graphite) tag format: ;<tag-key>=<tag-value>;...
// https://graphite.readthedocs.io/en/latest/tags.html
func (f *carbon) FormatTags(tags *string) *[]byte {
	if len(*tags) == 0 {
		byteStr := []byte("")
		return &byteStr
	}

	byteStr := []byte(fmt.Sprintf(";%s", strings.ReplaceAll(*tags, ",", ";"))) //TODO inline this?
	return &byteStr
}
