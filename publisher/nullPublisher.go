package publisher

type nullPublisher struct {
}

func NewNullPublisher(endpoint string) Publisher {
	return &nullPublisher{}
}

func (p *nullPublisher) PublishBytes(byteArrays *[]*[]byte) error {
	return nil
}
