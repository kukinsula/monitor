package mq

type Service struct {
	*client
	codec Codec
}

func NewService(address string) *Service {
	return &Service{
		client: newClient(address),
		codec:  &JsonCodec{},
	}
}
