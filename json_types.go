package jessy

import (
	"encoding"
	"encoding/json"
	"reflect"
)

var (
	typeIsZeroer = reflect.TypeFor[IsZeroer]()

	typeJsonAppender = reflect.TypeFor[JsonAppender]()
	typeTextAppender = reflect.TypeFor[TextAppender]()

	typeMarshaler     = reflect.TypeFor[Marshaler]()
	typeTextMarshaler = reflect.TypeFor[TextMarshaler]()

	// TODO:
	// typeTextUnmarshaler = reflect.TypeFor[TextUnmarshaler]()
	// typeUnmarshaler     = reflect.TypeFor[Unmarshaler]()
)

type (
	// A Number represents a JSON number literal.
	Number = json.Number

	// RawMessage is a raw encoded JSON value.
	// It implements [Marshaler] and [Unmarshaler] and can
	// be used to delay JSON decoding or precompute a JSON encoding.
	RawMessage = json.RawMessage

	// type TextMarshaler interface {
	//	 MarshalText() (text []byte, err error)
	// }
	TextMarshaler = encoding.TextMarshaler
	// type TextUnmarshaler interface {
	//	 UnmarshalText(text []byte) error
	// }
	TextUnmarshaler = encoding.TextUnmarshaler

	// type Marshaler interface {
	//	 MarshalJSON() ([]byte, error)
	// }
	Marshaler = json.Marshaler
	// type Unmarshaler interface {
	//	 UnmarshalJSON([]byte) error
	// }
	Unmarshaler = json.Unmarshaler

	JsonAppender interface {
		AppendJSON([]byte) ([]byte, error)
	}
	// type TextAppender interface {
	// 	 AppendText([]byte) ([]byte, error)
	// }
	TextAppender = encoding.TextAppender

	IsZeroer interface {
		IsZero() bool
	}
)
