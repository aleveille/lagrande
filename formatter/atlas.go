package formatter

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/aleveille/lagrande/metric"
)

// References:
//  - https://raw.githubusercontent.com/Netflix/atlas/v1.5.x/scripts/publish-test.sh

type atlas struct {
}

/*
var jsonStart = []byte("{")
var metricsArrStart = []byte(",\"metrics\": [")
var metricStart = []byte("{")
var metricContinue = []byte(",{")
var metricTimestamp = []byte(",\"timestamp\":")
var metricValue = []byte(",\"value\":")
var metricEnd = []byte("}")
var jsonEnd = []byte("]}")
*/

// TODO Comment exported function NewAtlasFormatter()
func NewAtlasFormatter() Formatter {
	return &atlas{}
}

// Format according to Atlas format (see atlas.example)
func (f *atlas) FormatData(metrics *[]*metric.Metric) *[]*[]byte {
	r := make([]*[]byte, (7*len(*metrics))+6)

	//0:   {
	//1:     "tags": {
	//	       "nf.node": "$node"
	//	     }
	//2:     ,"metrics":
	//3:     [
	//N.0      ,{ or { for the first metric   We also add 3 to N.0 to offset for the previous fields
	//N.1	     "tags": {
	//		       "name": "anwserToEverything",
	//		       "atlas.dstype": "gauge"
	//		     }
	//N.2 & 3    ,"timestamp": $timestamp
	//N.4 & 5    ,"value": 42
	//N.6	   }
	//5:     ]
	//6:   }

	r[0] = &byteForCurlyOpen
	r[1] = (*metrics)[0].Metadata.Tags
	r[2] = &bytesForCommaMetricsColon
	r[3] = &byteForBraceOpen
	i := 0
	firstMetric := true
	for ; i < len(*metrics); i++ {
		if firstMetric {
			r[(7*i)+4] = &byteForCurlyOpen
			firstMetric = false
		} else {
			r[(7*i)+4] = &bytesForAtlasMetricContinue
		}
		metricTags := fmt.Sprintf("name=%s, atlas.dstype=%s", *(*metrics)[i].Name, *(*metrics)[i].Metadata.MetricType)
		r[(7*i)+1+4] = f.FormatTags(&metricTags)
		r[(7*i)+2+4] = &bytesForCommaTimestampColon
		byteTs := []byte(fmt.Sprintf("%d", *((*metrics)[i].Timestamp)/1000))
		r[(7*i)+3+4] = &byteTs
		r[(7*i)+4+4] = &bytesForCommaValueColon
		r[(7*i)+5+4] = (*metrics)[i].Value
		r[(7*i)+6+4] = &byteForCurlyClose
	}

	r[4+len(*metrics)*7] = &byteForBraceClose
	r[5+len(*metrics)*7] = &byteForCurlyClose
	return &r
}

// Format a series of comma-delimited strings of key=value into Atlas tag format:
// "tags": {
//   "name": "randomValue",
//   "atlas.dstype": "gauge"
// },
// https://raw.githubusercontent.com/Netflix/atlas/v1.5.x/scripts/publish-test.sh
func (f *atlas) FormatTags(tags *string) *[]byte {
	var sb strings.Builder
	sb.WriteString("\"tags\":{")

	tagTokenizerRE := regexp.MustCompile(`[0-9A-Za-z_\.]+=[0-9A-Za-z_\.]+`)

	firstTag := true
	for _, m := range tagTokenizerRE.FindAllString(*tags, -1) {
		kv := strings.Split(m, "=")
		if !firstTag {
			sb.WriteString(",")
		} else {
			firstTag = false
		}
		sb.WriteString(fmt.Sprintf("\"%s\":\"%s\"", kv[0], kv[1]))
	}

	sb.WriteString("}")
	sbBytes := []byte(sb.String())
	return &sbBytes
}
