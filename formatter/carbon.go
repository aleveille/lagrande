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
	r := make([]*[]byte, 7*len(*metrics))

	for i, m := range *metrics {
		byteTs := []byte(fmt.Sprintf("%d", *(m.Timestamp)/1000/1000/1000))

		r[(7 * i)] = m.Name
		r[(7*i)+1] = m.Metadata.Tags
		r[(7*i)+2] = &byteForSpace
		r[(7*i)+3] = m.Value
		r[(7*i)+4] = &byteForSpace
		r[(7*i)+5] = &byteTs
		r[(7*i)+6] = &byteForLineReturn
	}

	return &r
}

// Format a series of comma-delimited strings of key=value into Carbon (Graphite) tag format: ;<tag-key>=<tag-value>;...
// https://graphite.readthedocs.io/en/latest/tags.html
func (f *carbon) FormatTags(tags *string) *[]byte {
	if len(*tags) == 0 {
		byteStr := []byte{}
		return &byteStr
	}

	byteStr := []byte(fmt.Sprintf(";%s", strings.ReplaceAll(*tags, ",", ";"))) //TODO inline this?
	return &byteStr
}
