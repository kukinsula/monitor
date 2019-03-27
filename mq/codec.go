package mq

import "encoding/json"

type Codec interface {
	Marshal(data interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
}

type JsonCodec struct{}

func (codec JsonCodec) Marshal(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}

func (codec JsonCodec) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
