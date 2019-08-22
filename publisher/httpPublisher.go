package publisher

import (
	"bytes"
	"errors"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/aleveille/lagrande/metric"
)

type httpPublisher struct {
	httpClient *http.Client
	endpoint   string
}

func NewHttpPublisher(endpoint string) Publisher {
	httpTransport := &http.Transport{
		DisableCompression: true,
		MaxConnsPerHost:    0,
		Dial: (&net.Dialer{
			Timeout: 200 * time.Millisecond,
		}).Dial,
		IdleConnTimeout:       200 * time.Millisecond,
		ResponseHeaderTimeout: 500 * time.Millisecond,
		TLSHandshakeTimeout:   500 * time.Millisecond,
	}

	httpClient := &http.Client{
		Timeout:   1 * time.Second,
		Transport: httpTransport,
	}

	return &httpPublisher{httpClient: httpClient, endpoint: endpoint}
}

// PublishMetrics is unimplemented for httpPublisher
func (p *httpPublisher) PublishMetrics(metrics *[]*metric.Metric) error {
	return errors.New("PublishMetrics is not supported for HTTP publisher yet")
}

func (p *httpPublisher) PublishBytes(byteArrays *[]*[]byte) error {
	readers := make([]io.Reader, len(*byteArrays))

	//var sb strings.Builder
	for i, bArr := range *byteArrays {
		//sb.WriteString(string(*bArr))
		// TODO: those readers force the copy of bytes into the reader (I think?). We should create our type which would use the pointer
		//       and read from the underlying array
		readers[i] = bytes.NewReader(*bArr)
	}
	//fmt.Println("DEBUG:")
	//fmt.Println(sb.String())

	multiReader := io.MultiReader(readers...)
	_, err := p.httpClient.Post(p.endpoint, "application/json", multiReader)
	return err
}

type ByteArrayReader struct {
	arrayOffset int
	byteOffset  int
	arrays      *[]*[]byte
}

// https://golang.org/src/bytes/reader.go?s=3877:3909
func NewByteArrayReader(arrays *[]*[]byte) *ByteArrayReader {
	bar := ByteArrayReader{arrayOffset: 0, byteOffset: 0, arrays: arrays}
	return &bar
}

func (r *ByteArrayReader) Read(p []byte) (int, error) {
	//canRead := len(p)
	//readByte := 0
	//var err error

	//return readByte, err
	return 0, nil
}

type Reader interface {
	Read(p []byte) (n int, err error)
}
