package protocol

import "encoding/json"

// JsonCodec implements the Codec interface using JSON for encoding and decoding.
type JsonCodec struct{}

// NewJsonCodec creates a new instance of JsonCodec.
func NewJsonCodec() *JsonCodec {
	return &JsonCodec{}
}

// Encode marshals the given data structure into a JSON byte slice.
func (c *JsonCodec) Encode(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}

// Decode unmarshals the given JSON byte slice into the provided data structure.
func (c *JsonCodec) Decode(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
