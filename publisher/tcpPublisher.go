package publisher

import (
	"errors"
	"net"
	"time"

	"github.com/aleveille/lagrande/metric"
	log "github.com/sirupsen/logrus"
)

type tcpPublisher struct {
	dialer   net.Dialer
	conn     net.Conn
	endpoint string
}

func NewTcpPublisher(endpoint string) Publisher {
	p := tcpPublisher{dialer: net.Dialer{Timeout: 200 * time.Millisecond}, endpoint: endpoint}
	p.connect()
	return &p
}

func (p *tcpPublisher) connect() {
	localConn, err := p.dialer.Dial("tcp", p.endpoint)

	if err != nil {
		log.Errorf("Error establishing tcp connection to %s:\n%s", p.endpoint, err)
	} else {
		p.conn = localConn
	}
}

// PublishMetrics is unimplemented for tcpPublisher
func (p *tcpPublisher) PublishMetrics(metrics *[]*metric.Metric) error {
	return errors.New("PublishMetrics is not supported for TCP publisher yet")
}

func (p *tcpPublisher) PublishBytes(byteArrays *[]*[]byte) error {
	if p.conn == nil {
		p.connect()
		if p.conn == nil {
			return errors.New("couldn't successfully open TCP connection")
		}
	}

	p.conn.SetWriteDeadline(time.Now().Add(400 * time.Millisecond))

	for _, bArr := range *byteArrays {
		if bArr != nil {
			_, err := p.conn.Write(*bArr)
			if err != nil {
				closeErr := p.conn.Close()
				if closeErr != nil {
					log.Errorf("Error closing tcp connection to %s:\n%s", p.endpoint, closeErr)
				}
				p.conn = nil

				return err
			}
		}
	}

	return nil
}
