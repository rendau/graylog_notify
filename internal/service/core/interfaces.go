package core

type SourceI interface {
	ParseMessage(data []byte) (*Message, error)
}

type DestinationI interface {
	Send(msg *Message) error
}
