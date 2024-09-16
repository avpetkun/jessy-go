package jessy

import (
	"encoding/json"
	"io"
	"net/http"
	_ "net/http/pprof"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/avpetkun/jessy-go/zgo"
	"github.com/avpetkun/jessy-go/zstr"
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
	ByteArr [10]byte

	ByteArr5    [5]byte `json:",omitempty"`
	ByteArrOmit [5]byte `json:",omitempty"`

	StrSlice  []string `json:"strSlice"`
	ByteSlice []byte

	intHidden int
	IntOmit   int `json:",omitempty"`

	TEmbedded
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

	JMarshalPtrEmpty *JMarshalPtr
	JMarshalPtrOmit  *JMarshalPtr `json:",omitempty"`

	AppendMarshalVal AppendMarshalVal

	TMarhalVal TMarshalVal

	NilMap  map[string]int
	OmitMap map[string]int `json:",omitempty"`

	MapValVal map[string]int
	MapEmpty  map[string]int
	MapAnyVal map[any]int
	MapValAny map[int]any
	MapAnyAny map[any]any

	MarshalMapKey map[TextMapKey]*TMarshalVal

	AnyVal1 any
	AnyVal2 any

	_ struct{}
}

type AppendMarshalVal struct{ data []byte }

func (v AppendMarshalVal) AppendMarshalJSON(dst []byte) ([]byte, error) {
	return append(dst, v.data...), nil
}

func (v AppendMarshalVal) MarshalJSON() ([]byte, error) { return v.data, nil }

type TMarshalVal struct{ data []byte }

func (v TMarshalVal) MarshalText() ([]byte, error) { return v.data, nil }

type JMarshalPtr struct{ Data []byte }

func (s *JMarshalPtr) MarshalJSON() ([]byte, error) { return s.Data, nil }

type JMarshalVal struct{ Data []byte }

func (s JMarshalVal) MarshalJSON() ([]byte, error) { return s.Data, nil }

type TextMapKey struct{ string }

func (v TextMapKey) MarshalText() ([]byte, error) { return zgo.S2B(v.string), nil }

type TEmbedded struct{ EmbedVpub int }

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

		IntArr3:  [3]int{1, 2, 3},
		IntArr2:  [2]int{1, 2},
		ByteArr:  [10]byte{1, 2, 3},
		ByteArr5: [5]byte{1, 2, 3, 4, 5},

		StrSlice:  []string{"a", "b", "c"},
		ByteSlice: []byte("hello!"),

		intHidden: 123,

		TEmbedded: TEmbedded{123},
		embedded:  embedded{3145},

		EmbeddedPtr: &EmbeddedPtr{789},

		Nested1: Nested{435345, 2},
		Nested2: nested{78634},

		NestedPtr1: &Nested{986754, 3},
		NestedPtr2: &Nested{986755, 33},

		nestedHidden: Nested{56432, 4},

		JMarshalValVal: JMarshalVal{[]byte(`"JMarshalValVal"`)},
		JMarshalValPtr: JMarshalPtr{[]byte(`"JMarshalValPtr"`)},
		JMarshalPtrVal: &JMarshalVal{[]byte(`"JMarshalPtrVal"`)},
		JMarshalPtrPtr: &JMarshalPtr{[]byte(`"JMarshalPtrPtr"`)},

		TMarhalVal: TMarshalVal{[]byte("TMarhalVal")},

		AppendMarshalVal: AppendMarshalVal{[]byte(`"AppendMarshalVal"`)},

		NilMap:  nil,
		OmitMap: nil,

		MapValVal: map[string]int{"a": 1, "b": 2},
		MapEmpty:  map[string]int{},
		MapAnyVal: map[any]int{1: 2, 3: 4},
		MapValAny: map[int]any{1: 2, 2: "b"},
		MapAnyAny: map[any]any{1: "a", "b": 2},

		MarshalMapKey: map[TextMapKey]*TMarshalVal{
			{"a"}:   {[]byte("a1")},
			{"b"}:   {[]byte("b1")},
			{"c"}:   {[]byte("c1")},
			{"de"}:  {[]byte("de1")},
			{"fgk"}: {[]byte("fgk1")},
		},

		AnyVal1: 123,
		AnyVal2: "abc",
	}
}

var expectedMarshalResult = `{"embed_v_ptr":789,"EmbedVpub":123,"EmbedVpriv":3145,"AnyVal1":123,"AnyVal2":"abc","AppendMarshalVal":"AppendMarshalVal","Bool1":true,"Bool2":false,"Byte":12,"ByteArr":"custom:0x01020300000000000000","ByteArr5":"0x0102030405","ByteSlice":"aGVsbG8h","Float32":16.17,"Float64":17.18,"Int":123,"Int16":567,"Int32":789,"Int64":-91011,"Int8":35,"IntArr2":[1,2],"IntArr3":[1,2,3],"JMarshalPtrEmpty":null,"JMarshalPtrPtr":"JMarshalPtrPtr","JMarshalPtrVal":"JMarshalPtrVal","JMarshalValPtr":"JMarshalValPtr","JMarshalValVal":"JMarshalValVal","MapAnyAny":{"1":"a","b":2},"MapAnyVal":{"1":2,"3":4},"MapEmpty":{},"MapValAny":{"1":2,"2":"b"},"MapValVal":{"a":1,"b":2},"MarshalMapKey":{"a":"a1","b":"b1","c":"c1","de":"de1","fgk":"fgk1"},"Nested1":{"nested_u":435345,"nested_v":2},"Nested2":{"nested_u_priv":78634},"NestedPtr1":{"nested_u":986754,"nested_v":3},"NestedPtr2":{"nested_u":986755,"nested_v":33},"NestedPtrNil":null,"NilMap":null,"String":"test_string","TMarhalVal":"TMarhalVal","Uint16":1314,"Uint32":1415,"Uint64":1516,"Uint8":13,"strSlice":["a","b","c"]}`

func TestMarshal(t *testing.T) {
	v := getTestStruct()

	AddValueEncoder(func(flags Flags) ValueEncoder[[10]byte] {
		return func(dst []byte, v [10]byte) ([]byte, error) {
			dst = append(dst, `"custom:`...)
			dst = zstr.AppendHex(dst, v[:])
			dst = append(dst, '"')
			return dst, nil
		}
	})

	for range 1000 {
		data, err := Marshal(&v)
		require.NoError(t, err)
		require.Equal(t, expectedMarshalResult, string(data))

		data, err = Marshal(v)
		require.NoError(t, err)
		require.Equal(t, expectedMarshalResult, string(data))
	}
}

func TestMarshalLoop(t *testing.T) {
	t.SkipNow()

	go http.ListenAndServe(":4114", nil)

	v := getTestStruct()
	vp := &v

	AddValueEncoder(func(flags Flags) ValueEncoder[[10]byte] {
		return func(dst []byte, v [10]byte) ([]byte, error) {
			dst = append(dst, `"custom:`...)
			dst = zstr.AppendHex(dst, v[:])
			dst = append(dst, '"')
			return dst, nil
		}
	})

	var data []byte
	for {
		data, _ = AppendMarshal(data[:0], vp)
	}
}

func BenchmarkMarshal(b *testing.B) {
	vv := getTestStruct()
	value := &vv

	b.Run("jessy-fast", func(b *testing.B) {
		buf := make([]byte, 1024)
		b.ResetTimer()
		for range b.N {
			buf, _ = AppendMarshalFast(buf[:0], value)
		}
	})
	b.Run("jessy-standard", func(b *testing.B) {
		buf := make([]byte, 1024)
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
