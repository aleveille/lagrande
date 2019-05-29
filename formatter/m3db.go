package formatter

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/aleveille/lagrande/metric"
)

// References:
//  - https://github.com/m3db/m3

type m3db struct {
}

// TODO Comment exported function NewAtlasFormatter()
func NewM3DBFormatter() Formatter {
	return &m3db{}
}

// Format according to M3DB format (see m3db.example)
func (f *m3db) FormatData(metrics *[]*metric.Metric) *[]*[]byte {
	r := make([]*[]byte, 14)

	//0:   {
	//1:     "namespace":"
	//2:     $namespace
	//3:     ","id":"
	//4:     $id
	//5:     "
	//6:     ,
	//7:     "tags":[ ... ]
	//8:     ,"datapoint":{"timestamp":
	//9:      $timestamp
	//10:      ,
	//11:      "value":
	//12:      $value
	//13:    }
	//14:   }

	r[0] = &byteForCurlyOpen
	r[1] = &bytesForNamespaceColonQuote
	// TODO dynamic?
	bytesNamespaceName := []byte("default")
	r[2] = &bytesNamespaceName
	r[3] = &bytesForQuoteCommaIdColonQuote
	// TODO dynamic?
	bytesIdName := []byte("foo")
	r[4] = &bytesIdName
	r[5] = &byteForQuote
	r[6] = &byteForComma
	r[7] = f.finalizeTags((*metrics)[0].Metadata.Tags, (*metrics)[0].Metadata.Name)
	r[8] = &bytesForCommaDatapointColonTimestampColon
	byteTs := []byte(fmt.Sprintf("%d", *((*metrics)[0].Timestamp)/1000/1000/1000))
	r[9] = &byteTs
	r[10] = &bytesForCommaValueColon
	r[11] = (*metrics)[0].Value
	r[12] = &byteForCurlyClose
	r[13] = &byteForCurlyClose
	return &r
}

// Format a series of comma-delimited strings of key=value into M3DB tag format:
// "tags": [
//   {
//     "name": "randomValue",
//     "atlas.dstype": "gauge"
//   },...
// ]
// https://github.com/m3db/m3#write-a-datapoint
func (f *m3db) FormatTags(tags *string) *[]byte {
	var sb strings.Builder

	tagTokenizerRE := regexp.MustCompile(`[[:word:]]+=[[:word:]]+`)

	for _, m := range tagTokenizerRE.FindAllString(*tags, -1) {
		sb.WriteString(",") // For M3DB, we always preprend the "__name__" tag when finalizing the tags so we don't have to worry about the leading comma
		kv := strings.Split(m, "=")
		sb.WriteString(fmt.Sprintf("{\"name\":\"%s\",\"value\":\"%s\"}", kv[0], kv[1]))
	}

	sbBytes := []byte(sb.String())
	return &sbBytes
}

func (f *m3db) finalizeTags(tags *[]byte, name *[]byte) *[]byte {
	var sb strings.Builder
	sb.WriteString("\"tags\":[")

	sb.WriteString("{\"name\":\"__name__\",\"value\":\"")
	sb.WriteString(string(*name))
	sb.WriteString("\"}")
	sb.WriteString(string(*tags))

	sb.WriteString("]")
	sbBytes := []byte(sb.String())
	return &sbBytes
}
