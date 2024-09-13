package jessy

import (
	"encoding/json"
	"fmt"
	"io"
	"testing"
)

type Struct struct {
	Bool1 bool
	Bool2 bool

	Int     int
	Int8    int8
	Int16   int16
	Int32   int32
	Int64   int64
	Byte    byte
	Uint8   uint8
	Uint16  uint16
	Uint32  uint32
	Uint64  uint64
	Float32 float32
	Float64 float64

	String string

	IntArr3 [3]int
	IntArr2 [2]int

	StrSlice []string `json:"strSlice"`

	intHidden int
	IntOmit   int `json:",omitempty"`

	Embedded
	embedded
	*EmbeddedPtr

	Nested1 Nested
	Nested2 nested

	NestedPtr1   *Nested
	NestedPtr2   *Nested
	NestedPtrNil *Nested

	NestedPtrOmitEmpty *Nested `json:",omitempty"`

	nestedHidden Nested

	JMarshalValVal JMarshalVal
	JMarshalValPtr JMarshalPtr
	JMarshalPtrVal *JMarshalVal
	JMarshalPtrPtr *JMarshalPtr
}

type JMarshalPtr struct{}

var jMarshalPtrData = []byte(`"JMarshalPtr"`)

func (s *JMarshalPtr) MarshalJSON() ([]byte, error) {
	return jMarshalPtrData, nil
}

type JMarshalVal struct{}

var jMarshalValData = []byte(`"JMarshalVal"`)

func (s JMarshalVal) MarshalJSON() ([]byte, error) {
	return jMarshalValData, nil
}

type Embedded struct{ EmbedVpub int }

type EmbeddedPtr struct {
	EmbedVPtr int `json:"embed_v_ptr"`
}

type embedded struct{ EmbedVpriv int }

type Nested struct {
	U int `json:"nested_u"`
	V int `json:"nested_v"`
}

type nested struct {
	U int `json:"nested_u_priv"`
}

func getTestStruct() Struct {
	return Struct{
		Bool1:   true,
		Bool2:   false,
		Int:     123,
		Int8:    35,
		Int16:   567,
		Int32:   789,
		Int64:   -91011,
		Byte:    12,
		Uint8:   13,
		Uint16:  1314,
		Uint32:  1415,
		Uint64:  1516,
		Float32: 16.17,
		Float64: 17.18,

		String: "test_string",

		IntArr3: [3]int{1, 2, 3},
		IntArr2: [2]int{1, 2},

		StrSlice: []string{"a", "b", "c"},

		intHidden: 123,

		Embedded: Embedded{123},
		embedded: embedded{3145},

		EmbeddedPtr: &EmbeddedPtr{789},

		Nested1: Nested{435345, 2},
		Nested2: nested{78634},

		NestedPtr1: &Nested{986754, 3},
		NestedPtr2: &Nested{986755, 33},

		nestedHidden: Nested{56432, 4},

		JMarshalValVal: JMarshalVal{},
		JMarshalValPtr: JMarshalPtr{},
		JMarshalPtrVal: &JMarshalVal{},
		JMarshalPtrPtr: &JMarshalPtr{},
	}
}

func TestMarshal(t *testing.T) {
	v := getTestStruct()

	data, err := Marshal(&v)
	fmt.Println("TestMarshal data", string(data))
	fmt.Println("TestMarshal err", err)
}

func BenchmarkMarshal(b *testing.B) {
	vv := getTestStruct()
	value := &vv

	buf := make([]byte, 1024)

	b.Run("jessy", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			buf, _ = AppendMarshal(buf[:0], value)
		}
	})
	b.Run("json", func(b *testing.B) {
		enc := json.NewEncoder(io.Discard)
		b.ResetTimer()
		for range b.N {
			enc.Encode(value)
		}
	})
}
