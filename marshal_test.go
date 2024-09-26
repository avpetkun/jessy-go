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
	*Struct

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

	StrSlice    []string  `json:"strSlice"`
	StrSlicePtr *[]string `json:"strSlicePtr"`
	ByteSlice   []byte

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

	NestedJMarshalPtrPtr  NestedJMarshalPtrPtr
	NestedJMarshalPtrPtr2 NestedJMarshalPtrPtr2

	TMarhalVal TMarshalVal

	JMarshalPtrEmpty *JMarshalPtr
	JMarshalPtrOmit  *JMarshalPtr `json:",omitempty"`

	AppendMarshalVal AppendMarshalVal

	NilMap  map[string]int
	OmitMap map[string]int `json:",omitempty"`

	MapValVal map[string]int
	MapEmpty  map[string]int
	MapAnyVal map[any]int
	MapValAny map[int]any
	MapAnyAny map[any]any

	MapValValPtr *map[string]int

	MarshalMapKey    map[TextMapKey]*TMarshalVal
	MarshalMapKeyPtr map[*TextMapKey]*TMarshalVal

	AnyVal1 any
	AnyVal2 any

	Bool1Ptr *bool
	Bool2Ptr *bool

	IntPtr     *int
	Int8Ptr    *int8
	Int16Ptr   *int16
	Int32Ptr   *int32
	Int64Ptr   *int64
	BytePtr    *byte
	Uint8Ptr   *uint8
	Uint16Ptr  *uint16
	Uint32Ptr  *uint32
	Uint64Ptr  *uint64
	Float32Ptr *float32
	Float64Ptr *float64

	StringPtr *string

	IntArr3Ptr *[3]int

	AnyValPtr *any

	DoubleIntPtr      **int
	DoubleStrSlicePtr **[]string

	StructSlice    []SliceStruct
	StructSlicePtr []*SliceStruct

	Complex64     complex64
	ComplexNeg128 complex128

	_ struct{}
}

type SliceStruct struct {
	A int
	B int
}

type AppendMarshalVal struct{ data []byte }

func (v AppendMarshalVal) AppendMarshalJSON(dst []byte) ([]byte, error) {
	return append(dst, v.data...), nil
}

func (v AppendMarshalVal) MarshalJSON() ([]byte, error) { return v.data, nil }

type TMarshalVal struct{ data []byte }

func (v TMarshalVal) MarshalText() ([]byte, error) { return v.data, nil }

type NestedJMarshalPtrPtr struct{ *JMarshalPtr }

type NestedJMarshalPtrPtr2 struct {
	*JMarshalPtr
	X int
}

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
	s := Struct{
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
		ByteSlice: []byte(`"hello!"`),

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

		NestedJMarshalPtrPtr: NestedJMarshalPtrPtr{
			&JMarshalPtr{[]byte(`"NestedJMarshalPtrPtr"`)},
		},
		NestedJMarshalPtrPtr2: NestedJMarshalPtrPtr2{
			&JMarshalPtr{[]byte(`"NestedJMarshalPtrPtr2"`)},
			123,
		},

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

		StructSlice: []SliceStruct{
			{1, 2}, {3, 4},
		},
		StructSlicePtr: []*SliceStruct{
			{1, 2}, {3, 4},
		},

		Complex64:     123 + 456i,
		ComplexNeg128: -123 - 4.56i,
	}

	s.IntPtr = &s.Int
	s.StringPtr = &s.String

	s.StrSlicePtr = &s.StrSlice
	s.MapValValPtr = &s.MapValVal
	s.MarshalMapKeyPtr = map[*TextMapKey]*TMarshalVal{
		{"a"}:   {[]byte("a1")},
		{"b"}:   {[]byte("b1")},
		{"c"}:   {[]byte("c1")},
		{"de"}:  {[]byte("de1")},
		{"fgk"}: {[]byte("fgk1")},
	}

	s.Bool1Ptr = &s.Bool1
	s.Bool2Ptr = &s.Bool2
	s.IntPtr = &s.Int
	s.Int8Ptr = &s.Int8
	s.Int16Ptr = &s.Int16
	s.Int32Ptr = &s.Int32
	s.Int64Ptr = &s.Int64
	s.BytePtr = &s.Byte
	s.Uint8Ptr = &s.Uint8
	s.Uint16Ptr = &s.Uint16
	s.Uint32Ptr = &s.Uint32
	s.Uint64Ptr = &s.Uint64
	s.Float32Ptr = &s.Float32
	s.Float64Ptr = &s.Float64
	s.StringPtr = &s.String
	s.IntArr3Ptr = &s.IntArr3

	s.AnyValPtr = &s.AnyVal1

	s.DoubleIntPtr = &s.IntPtr
	s.DoubleStrSlicePtr = &s.StrSlicePtr

	return s
}

var expectedMarshalResult = `{"EmbedVpub":123,"EmbedVpriv":3145,"embed_v_ptr":789,"AnyVal1":123,"AnyVal2":"abc","AnyValPtr":123,"AppendMarshalVal":"AppendMarshalVal","Bool1":true,"Bool1Ptr":true,"Bool2":false,"Bool2Ptr":false,"Byte":12,"ByteArr":"custom:0x01020300000000000000","ByteArr5":"0x0102030405","BytePtr":12,"ByteSlice":"ImhlbGxvISI=","Complex64":"123+456i","ComplexNeg128":"-123-4.56i","DoubleIntPtr":123,"DoubleStrSlicePtr":["a","b","c"],"Float32":16.17,"Float32Ptr":16.17,"Float64":17.18,"Float64Ptr":17.18,"Int":123,"Int16":567,"Int16Ptr":567,"Int32":789,"Int32Ptr":789,"Int64":-91011,"Int64Ptr":-91011,"Int8":35,"Int8Ptr":35,"IntArr2":[1,2],"IntArr3":[1,2,3],"IntArr3Ptr":[1,2,3],"IntPtr":123,"JMarshalPtrEmpty":null,"JMarshalPtrPtr":"JMarshalPtrPtr","JMarshalPtrVal":"JMarshalPtrVal","JMarshalValPtr":"JMarshalValPtr","JMarshalValVal":"JMarshalValVal","MapAnyAny":{"1":"a","b":2},"MapAnyVal":{"1":2,"3":4},"MapEmpty":{},"MapValAny":{"1":2,"2":"b"},"MapValVal":{"a":1,"b":2},"MapValValPtr":{"a":1,"b":2},"MarshalMapKey":{"a":"a1","b":"b1","c":"c1","de":"de1","fgk":"fgk1"},"MarshalMapKeyPtr":{"a":"a1","b":"b1","c":"c1","de":"de1","fgk":"fgk1"},"Nested1":{"nested_u":435345,"nested_v":2},"Nested2":{"nested_u_priv":78634},"NestedJMarshalPtrPtr":"NestedJMarshalPtrPtr","NestedJMarshalPtrPtr2":"NestedJMarshalPtrPtr2","NestedPtr1":{"nested_u":986754,"nested_v":3},"NestedPtr2":{"nested_u":986755,"nested_v":33},"NestedPtrNil":null,"NilMap":null,"String":"test_string","StringPtr":"test_string","StructSlice":[{"A":1,"B":2},{"A":3,"B":4}],"StructSlicePtr":[{"A":1,"B":2},{"A":3,"B":4}],"TMarhalVal":"TMarhalVal","Uint16":1314,"Uint16Ptr":1314,"Uint32":1415,"Uint32Ptr":1415,"Uint64":1516,"Uint64Ptr":1516,"Uint8":13,"Uint8Ptr":13,"strSlice":["a","b","c"],"strSlicePtr":["a","b","c"]}`

func TestMarshalAll(t *testing.T) {
	{
		Marshal(jsonbyte(7))
		Marshal(textbyte(4))
		Marshal((*unmarshalerText)(nil))
		println("--------")
		Marshal(map[*unmarshalerText]int{
			(*unmarshalerText)(nil): 1,
		})
		println("--------")
		Marshal(map[*unmarshalerText]int{
			(*unmarshalerText)(nil): 1,
			{"A", "B"}:              2,
		})
		rawText := json.RawMessage([]byte(`"foo"`))
		println("--------")
		Marshal(rawText)
		println("--------")
		Marshal(&rawText)
	}

	{
		data, err := Marshal(nil)
		require.NoError(t, err)
		require.Equal(t, `null`, string(data))
	}

	{
		data, err := Marshal(struct{ M *RawMessage }{})
		require.NoError(t, err)
		require.Equal(t, `{"M":null}`, string(data))
	}
	{
		data, err := Marshal(&struct{ M *RawMessage }{})
		require.NoError(t, err)
		require.Equal(t, `{"M":null}`, string(data))
	}
	{
		data, err := Marshal(struct{ M RawMessage }{})
		require.NoError(t, err)
		require.Equal(t, `{"M":null}`, string(data))
	}
	{
		data, err := Marshal(&struct{ M RawMessage }{})
		require.NoError(t, err)
		require.Equal(t, `{"M":null}`, string(data))
	}

	{
		data, err := Marshal(&struct {
			M RawMessage
			X int
		}{})
		require.NoError(t, err)
		require.Equal(t, `{"M":null,"X":0}`, string(data))
	}

	{
		data, err := Marshal(struct{ Int int }{})
		require.NoError(t, err)
		require.Equal(t, `{"Int":0}`, string(data))
	}
	{
		data, err := Marshal(&struct{ Int int }{})
		require.NoError(t, err)
		require.Equal(t, `{"Int":0}`, string(data))
	}
	{
		data, err := Marshal(struct{ Int *int }{})
		require.NoError(t, err)
		require.Equal(t, `{"Int":null}`, string(data))
	}
	{
		data, err := Marshal(&struct{ Int *int }{})
		require.NoError(t, err)
		require.Equal(t, `{"Int":null}`, string(data))
	}

	{
		rawNil := json.RawMessage(nil)
		str := &struct{ M *json.RawMessage }{&rawNil}
		println("str", str, "M", str.M)
		data, err := Marshal(str)
		require.NoError(t, err)
		require.Equal(t, `{"M":null}`, string(data))
		println("------")
	}
	{
		data, err := Marshal([]string{"a", "b"})
		require.NoError(t, err)
		require.Equal(t, `["a","b"]`, string(data))
		println("------")
	}
	{
		data, err := Marshal(&[]string{"a", "b"})
		require.NoError(t, err)
		require.Equal(t, `["a","b"]`, string(data))
		println("------")
	}
	{
		rawNil := json.RawMessage(nil)
		data, err := Marshal(struct{ V *[]any }{&[]any{rawNil}})
		require.NoError(t, err)
		require.Equal(t, `{"V":[null]}`, string(data))
		println("------")
	}
	{
		rawNil := json.RawMessage(nil)
		val := []any{rawNil}
		data, err := Marshal(&val)
		require.NoError(t, err)
		require.Equal(t, `[null]`, string(data))
		println("------")
	}
	{
		rawText := json.RawMessage([]byte(`"123"`))
		str := struct{ M *json.RawMessage }{&rawText}
		println("str", "M", str.M, rawText)
		data, err := Marshal(str)
		require.NoError(t, err)
		require.Equal(t, `{"M":"123"}`, string(data))
		println("------")
	}
	{
		rawText := json.RawMessage([]byte(`"123"`))
		str := struct {
			X int
			M *json.RawMessage
		}{123, &rawText}
		println("str", "M", str.M)
		data, err := Marshal(str)
		require.NoError(t, err)
		require.Equal(t, `{"M":"123","X":123}`, string(data))
		println("------")
	}
	{
		rawText := json.RawMessage([]byte(`"123"`))
		str := &struct {
			X int
			M *json.RawMessage
		}{123, &rawText}
		println("str", "M", str.M)
		data, err := Marshal(str)
		require.NoError(t, err)
		require.Equal(t, `{"M":"123","X":123}`, string(data))
		println("------")
	}
	{
		rawText := json.RawMessage([]byte(`"123"`))
		str := struct {
			M *json.RawMessage
			X int
		}{&rawText, 123}
		println("str", "M", str.M)
		data, err := Marshal(str)
		require.NoError(t, err)
		expectedData, _ := json.Marshal(str)
		require.Equal(t, string(expectedData), string(data))
		println("------")
	}

	{
		rawText := json.RawMessage([]byte(`"123"`))
		str := struct {
			M json.RawMessage
			X int
		}{rawText, 123}
		println("str", "M", str.M)
		data, err := Marshal(str)
		require.NoError(t, err)
		expectedData, _ := json.Marshal(str)
		require.Equal(t, string(expectedData), string(data))
		println("------")
	}
	{
		rawText := json.RawMessage([]byte(`"123"`))
		str := struct{ M json.RawMessage }{rawText}
		println("str", "M", str.M)
		data, err := Marshal(str)
		require.NoError(t, err)
		expectedData, _ := json.Marshal(str)
		require.Equal(t, string(expectedData), string(data))
		println("------")
	}
	{
		rawText := json.RawMessage([]byte(`"123"`))
		str := &struct{ M json.RawMessage }{rawText}
		println("str", "M", str.M)
		data, err := Marshal(str)
		require.NoError(t, err)
		expectedData, _ := json.Marshal(str)
		require.Equal(t, string(expectedData), string(data))
		println("------")
	}

	{
		data, err := Marshal(json.RawMessage([]byte{}))
		require.NoError(t, err)
		require.Equal(t, ``, string(data))
	}

	data, err := Marshal(struct{ M *json.RawMessage }{})
	require.NoError(t, err)
	require.Equal(t, `{"M":null}`, string(data))
	println("------")

	data, err = Marshal(json.RawMessage("123"))
	require.NoError(t, err)
	require.Equal(t, `123`, string(data))
	println("------")

	jsonMsg := json.RawMessage("123")
	data, err = Marshal(&jsonMsg)
	require.NoError(t, err)
	require.Equal(t, `123`, string(data))
	println("------")

	data, err = Marshal(123)
	require.NoError(t, err)
	require.Equal(t, `123`, string(data))
	println("------")

	data, err = Marshal("123")
	require.NoError(t, err)
	require.Equal(t, `"123"`, string(data))
	println("------")

	type Str struct {
		I int
		S string
	}
	str := &Str{I: 123, S: "test_str"}

	println("str", str, "i", &str.I, "s", &str.S)

	data, err = Marshal(str)
	require.NoError(t, err)
	require.Equal(t, `{"I":123,"S":"test_str"}`, string(data))
	println("------")

	data, err = Marshal(Str{I: 123, S: "str"})
	require.NoError(t, err)
	require.Equal(t, `{"I":123,"S":"str"}`, string(data))
	println("------")

	data, err = Marshal(&map[any]string{123: "M"})
	require.NoError(t, err)
	require.Equal(t, `{"123":"M"}`, string(data))

	data, err = Marshal(&map[string]any{"M": json.RawMessage(nil)})
	require.Equal(t, `{"M":null}`, string(data))
	require.NoError(t, err)

	data, err = Marshal(map[string]any{"M": json.RawMessage(nil)})
	require.Equal(t, `{"M":null}`, string(data))
	require.NoError(t, err)

	data, err = Marshal(map[string]int{
		"x:y": 1,
		"y:x": 2,
		"a:z": 3,
		"z:a": 4,
	})
	require.NoError(t, err)
	require.Equal(t, `{"a:z":3,"x:y":1,"y:x":2,"z:a":4}`, string(data))

	sliceNoCycle := []any{nil, nil}
	data, err = Marshal(sliceNoCycle)
	require.Equal(t, "[null,null]", string(data))
	require.NoError(t, err)

	data, err = Marshal(func() any {
		type (
			S2 struct{ Field string }
			S  struct{ *S2 }
		)
		return S{}
	}())
	require.NoError(t, err)
	require.Equal(t, "{}", string(data))

	v := getTestStruct()

	AddValueEncoder(func(flags Flags) ValueEncoder[[10]byte] {
		return func(dst []byte, v [10]byte) ([]byte, error) {
			dst = append(dst, `"custom:`...)
			dst = zstr.AppendHex(dst, v[:])
			dst = append(dst, '"')
			return dst, nil
		}
	})

	//data, _ = MarshalFastPretty(v)
	//os.WriteFile("min.json", data, os.ModePerm)

	return
	for range 100 {
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
		data, _ = AppendMarshalFast(data[:0], vp)
	}
}

func BenchmarkMarshal(b *testing.B) {
	vv := getTestStruct()
	value := &vv

	b.Run("jessy-hash", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			_, _ = Hash(value)
		}
	})
	b.Run("jessy-fast", func(b *testing.B) {
		buf := make([]byte, 1024)
		b.ResetTimer()
		for range b.N {
			buf, _ = AppendMarshalFast(buf[:0], value)
		}
	})
	b.Run("jessy-fast-pretty", func(b *testing.B) {
		buf := make([]byte, 1024)
		b.ResetTimer()
		for range b.N {
			buf, _ = AppendMarshalFlags(buf[:0], value, EncodeFastest|PrettySpaces)
		}
	})
	b.Run("jessy-standard", func(b *testing.B) {
		buf := make([]byte, 1024)
		b.ResetTimer()
		for range b.N {
			buf, _ = AppendMarshal(buf[:0], value)
		}
	})
	b.Run("jessy-standard-pretty", func(b *testing.B) {
		buf := make([]byte, 1024)
		b.ResetTimer()
		for range b.N {
			buf, _ = AppendMarshalFlags(buf[:0], value, EncodeStandard|PrettySpaces)
		}
	})
	b.Run("json", func(b *testing.B) {
		enc := json.NewEncoder(io.Discard)
		b.ResetTimer()
		for range b.N {
			enc.Encode(value)
		}
	})
	b.Run("json-indent", func(b *testing.B) {
		enc := json.NewEncoder(io.Discard)
		enc.SetIndent("", "\t")
		b.ResetTimer()
		for range b.N {
			enc.Encode(value)
		}
	})
}
