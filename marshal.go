package jessy

import (
	"reflect"
	"unsafe"
)

var (
	MarshalMaxDeep uint = 10

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

func Marshal(value any) (data []byte, err error) {
	return encodeAny(nil, value, EncodeStandard)
}

func AppendMarshal(dst []byte, value any) (data []byte, err error) {
	return encodeAny(dst, value, EncodeStandard)
}

func MarshalFast(value any) (data []byte, err error) {
	return encodeAny(nil, value, EncodeFastest)
}

func AppendMarshalFast(dst []byte, value any) (data []byte, err error) {
	return encodeAny(dst, value, EncodeFastest)
}

func AppendMarshalFlags(dst []byte, value any, flags Flags) (data []byte, err error) {
	return encodeAny(dst, value, flags)
}
