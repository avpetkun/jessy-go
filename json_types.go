package jessy

import (
	"encoding"
	"encoding/json"
	"reflect"
)

var (
	typeTextMarshaler   = reflect.TypeFor[TextMarshaler]()
	typeTextUnmarshaler = reflect.TypeFor[TextUnmarshaler]()

	typeMarshaler   = reflect.TypeFor[Marshaler]()
	typeUnmarshaler = reflect.TypeFor[Unmarshaler]()

	typeAppendMarshaler = reflect.TypeFor[AppendMarshaler]()
)

type (
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

	AppendMarshaler interface {
		AppendMarshalJSON(dst []byte) (newDst []byte, err error)
	}
)
