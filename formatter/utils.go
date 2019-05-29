package formatter

var byteForSpace = []byte(" ")
var byteForEqual = []byte("=")
var byteForComma = []byte(",")
var byteForQuote = []byte("\"")
var byteForLineReturn = []byte("\n")
var byteForCurlyOpen = []byte("{")
var bytesForAtlasMetricContinue = []byte(",{")
var byteForCurlyClose = []byte("}")
var byteForBraceOpen = []byte("[")
var byteForBraceClose = []byte("]")
var bytesForNamespaceColonQuote = []byte("\"namespace\":\"")
var bytesForQuoteCommaIdColonQuote = []byte("\",\"id\":\"")
var bytesForCommaTimestampColon = []byte(",\"timestamp\":")
var bytesForCommaDatapointColonTimestampColon = []byte(",\"datapoint\":{\"timestamp\":")
var bytesForCommaValueColon = []byte(",\"value\":")
var bytesForCommaMetricsColon = []byte(",\"metrics\":")
