package jessy

import (
	"encoding/json"
	"testing"

	"github.com/avpetkun/jessy-go/require"
	"github.com/avpetkun/jessy-go/zstr"
)

func TestHash(t *testing.T) {
	{
		s := getTestStruct()
		h1, err := Hash(s)
		require.NoError(t, err)

		s = getTestStruct()
		h2, err := Hash(s)
		require.NoError(t, err)

		s = getTestStruct()
		h3, err := Hash(&s)
		require.NoError(t, err)

		require.Equal(t, h1, h2)
		require.Equal(t, h1, h3)
	}
	{
		s := getTestMoreStruct()
		h1, err := Hash(s)
		require.NoError(t, err)

		s = getTestMoreStruct()
		h2, err := Hash(s)
		require.NoError(t, err)

		s = getTestMoreStruct()
		h3, err := Hash(&s)
		require.NoError(t, err)

		require.Equal(t, h1, h2)
		require.Equal(t, h1, h3)
	}
	{
		s1 := getTestStruct()
		h1, err := Hash(s1)
		require.NoError(t, err)

		s2 := getTestMoreStruct()
		h2, err := Hash(s2)
		require.NoError(t, err)

		require.NotEqual(t, h1, h2)
	}
}

func TestHashMany(t *testing.T) {
	Hash(nil)
	Hash(struct{ M *RawMessage }{})
	Hash(&struct{ M *RawMessage }{})
	Hash(struct{ M RawMessage }{})
	Hash(&struct{ M RawMessage }{})
	Hash(&struct {
		M RawMessage
		X int
	}{})
	Hash(struct{ Int int }{})
	Hash(&struct{ Int int }{})
	Hash(struct{ Int *int }{})
	Hash(&struct{ Int *int }{})

	{
		rawNil := json.RawMessage(nil)
		str := &struct{ M *json.RawMessage }{&rawNil}
		Hash(str)
	}

	Hash([]string{"a", "b"})
	Hash(&[]string{"a", "b"})

	{
		rawNil := json.RawMessage(nil)
		Hash(struct{ V *[]any }{&[]any{rawNil}})
	}
	{
		rawNil := json.RawMessage(nil)
		val := []any{rawNil}
		Hash(&val)
	}
	{
		rawText := json.RawMessage([]byte(`"123"`))
		str := struct{ M *json.RawMessage }{&rawText}
		Hash(str)
	}
	{
		rawText := json.RawMessage([]byte(`"123"`))
		str := struct {
			X int
			M *json.RawMessage
		}{123, &rawText}
		Hash(str)
	}
	{
		rawText := json.RawMessage([]byte(`"123"`))
		str := &struct {
			X int
			M *json.RawMessage
		}{123, &rawText}
		Hash(str)
	}
	{
		rawText := json.RawMessage([]byte(`"123"`))
		str := struct {
			M *json.RawMessage
			X int
		}{&rawText, 123}
		Hash(str)
	}

	{
		rawText := json.RawMessage([]byte(`"123"`))
		str := struct {
			M json.RawMessage
			X int
		}{rawText, 123}
		Hash(str)
	}
	{
		rawText := json.RawMessage([]byte(`"123"`))
		str := struct{ M json.RawMessage }{rawText}
		Hash(str)
	}
	{
		rawText := json.RawMessage([]byte(`"123"`))
		str := &struct{ M json.RawMessage }{rawText}
		Hash(str)
	}

	Hash(json.RawMessage([]byte{}))

	Hash(struct{ M *json.RawMessage }{})

	Hash(json.RawMessage("123"))

	jsonMsg := json.RawMessage("123")
	Hash(&jsonMsg)

	Hash(123)

	Hash("123")

	type Str struct {
		I int
		S string
	}
	str := &Str{I: 123, S: "test_str"}

	Hash(str)

	Hash(Str{I: 123, S: "str"})

	Hash(&map[any]string{123: "M"})

	Hash(&map[string]any{"M": json.RawMessage(nil)})

	Hash(map[string]any{"M": json.RawMessage(nil)})

	Hash(map[string]int{
		"x:y": 1,
		"y:x": 2,
		"a:z": 3,
		"z:a": 4,
	})

	sliceNoCycle := []any{nil, nil}
	Hash(sliceNoCycle)

	Hash(func() any {
		type (
			S2 struct{ Field string }
			S  struct{ *S2 }
		)
		return S{}
	}())

	v := getTestStruct()

	AddValueEncoder(func(flags Flags) ValueEncoder[[10]byte] {
		return func(dst []byte, v [10]byte) ([]byte, error) {
			dst = append(dst, `"custom:`...)
			dst = zstr.AppendHex(dst, v[:])
			dst = append(dst, '"')
			return dst, nil
		}
	})

	for range 100 {
		Hash(&v)

		Hash(v)

	}
}
