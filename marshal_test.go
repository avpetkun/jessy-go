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

	NestedPtr    *Nested
	NestedPtrNil *Nested

	NestedPtrOmitEmpty *Nested `json:",omitempty"`

	nestedHidden Nested
}

type Embedded struct{ EmbedVpub int }

type EmbeddedPtr struct {
	EmbedVPtr int `json:"embed_v_ptr"`
}

type embedded struct{ EmbedVpriv int }

type Nested struct {
	U int `json:"nested_u"`
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
		Int64:   91011,
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

		Nested1: Nested{435345},
		Nested2: nested{78634},

		NestedPtr: &Nested{986754},

		nestedHidden: Nested{56432},
	}
}

func TestMarshal(t *testing.T) {
	vv := getTestStruct()
	//vp := getTestStruct()

	data, err := Marshal(&vv)
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
