package publisher

type Publisher interface {
	PublishBytes(bytes *[]*[]byte) error
}
