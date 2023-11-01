package core

type DestinationI interface {
	Send(msg map[string]any) error
}
