package jessy

import (
	"bytes"
	"reflect"
	"unsafe"
)

var (
	MarshalMaxDeep = 10

	customEncoders []customEncoder
)

type UnsafeEncoder func(dst []byte, value unsafe.Pointer) ([]byte, error)
type ValueEncoder[T any] func(dst []byte, value T) ([]byte, error)

type customEncoder struct {
	reflect.Type
	Encoder func(flags Flags) UnsafeEncoder
}

func AddUnsafeEncoder[T any](encoder func(flags Flags) UnsafeEncoder) {
	customEncoders = append(customEncoders, customEncoder{
		Type:    reflect.TypeFor[T](),
		Encoder: encoder,
	})
}

func AddValueEncoder[T any](encoder func(flags Flags) ValueEncoder[T]) {
	AddUnsafeEncoder[T](func(flags Flags) UnsafeEncoder {
		valEnc := encoder(flags)
		return func(dst []byte, value unsafe.Pointer) ([]byte, error) {
			return valEnc(dst, *(*T)(value))
		}
	})
}

func Marshal(value any) ([]byte, error) {
	return encodeAny(nil, value, EncodeStandard)
}

func MarshalFast(value any) ([]byte, error) {
	return encodeAny(nil, value, EncodeFastest)
}

func MarshalPretty(value any) ([]byte, error) {
	return encodeAny(nil, value, EncodeStandard|PrettySpaces)
}

func MarshalFastPretty(value any) ([]byte, error) {
	return encodeAny(nil, value, EncodeFastest|PrettySpaces)
}

func AppendMarshal(dst []byte, value any) ([]byte, error) {
	return encodeAny(dst, value, EncodeStandard)
}

func AppendMarshalPretty(dst []byte, value any) ([]byte, error) {
	return encodeAny(dst, value, EncodeStandard|PrettySpaces)
}

func AppendMarshalFast(dst []byte, value any) ([]byte, error) {
	return encodeAny(dst, value, EncodeFastest)
}

func AppendMarshalFastPretty(dst []byte, value any) ([]byte, error) {
	return encodeAny(dst, value, EncodeFastest|PrettySpaces)
}

func AppendMarshalFlags(dst []byte, value any, flags Flags) ([]byte, error) {
	return encodeAny(dst, value, flags)
}

func MarshalFlags(value any, flags Flags) ([]byte, error) {
	return encodeAny(nil, value, flags)
}

func MarshalIndent(v any, prefix, indent string) (data []byte, err error) {
	data, err = Marshal(v)
	if err != nil {
		return
	}
	var buf bytes.Buffer
	err = Indent(&buf, data, prefix, indent)
	data = buf.Bytes()
	return
}
