package polaris

import (
	"bytes"
	"encoding/gob"
	"encoding/json"

	"github.com/pkg/errors"
)

type Encoder[T any] interface {
	Encode(v T) ([]byte, error)
	Decode([]byte) (T, error)
}

var (
	_ Encoder[any] = (*jsonEncoder[any])(nil)
)

type jsonEncoder[T any] struct{}

func (f jsonEncoder[T]) Encode(v T) ([]byte, error) {
	return json.Marshal(v)
}

func (f jsonEncoder[T]) Decode(data []byte) (t T, err error) {
	if err := json.Unmarshal(data, &t); err != nil {
		return t, errors.WithStack(err)
	}
	return t, nil
}

// google.golang.org/genai supports json
func JSONEncoder[T any]() *jsonEncoder[T] {
	return new(jsonEncoder[T])
}

type gobEncoder[T any] struct{}

func (g gobEncoder[T]) Encode(v T) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	if err := gob.NewEncoder(buf).Encode(v); err != nil {
		return nil, errors.WithStack(err)
	}
	return buf.Bytes(), nil
}

func (g gobEncoder[T]) Decode(data []byte) (t T, err error) {
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&t); err != nil {
		return t, errors.WithStack(err)
	}
	return t, nil
}

// cloud.google.com/go/vertexai/genai are not support json
func GobEncoder[T any]() *gobEncoder[T] {
	return new(gobEncoder[T])
}
