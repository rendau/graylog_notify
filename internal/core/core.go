package core

type Core struct {
	destination DestinationI
}

func NewCore(destination DestinationI) (*Core, error) {
	return &Core{
		destination: destination,
	}, nil
}
