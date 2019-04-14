package mq

type Service struct {
	*client
	codec Codec
}

func NewService(address string) (*Service, error) {
	client := newClient(address)

	_, err := client.Ping()
	if err != nil {
		return nil, err
	}

	return &Service{
		client: client,
		codec:  &JsonCodec{},
	}, nil
}
