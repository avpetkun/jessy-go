package jessy

import (
	"reflect"
	"unsafe"

	"github.com/avpetkun/jessy-go/zstr"
)

var (
	MarshalMaxDeep uint32 = 20

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

func MarshalFlags(value any, flags Flags) ([]byte, error) {
	return encodeAny(nil, value, flags)
}

func MarshalIndent(value any, prefix, indent string) ([]byte, error) {
	return AppendIndent(nil, value, prefix, indent)
}

func MarshalFastIndent(value any, prefix, indent string) ([]byte, error) {
	return AppendFastIndent(nil, value, prefix, indent)
}

func Append(dst []byte, value any) ([]byte, error) {
	return encodeAny(dst, value, EncodeStandard)
}

func AppendPretty(dst []byte, value any) ([]byte, error) {
	return encodeAny(dst, value, EncodeStandard|PrettySpaces)
}

func AppendFast(dst []byte, value any) ([]byte, error) {
	return encodeAny(dst, value, EncodeFastest)
}

func AppendFastPretty(dst []byte, value any) ([]byte, error) {
	return encodeAny(dst, value, EncodeFastest|PrettySpaces)
}

func AppendFlags(dst []byte, value any, flags Flags) ([]byte, error) {
	return encodeAny(dst, value, flags)
}

func AppendIndent(dst []byte, value any, prefix, indent string) ([]byte, error) {
	return AppendIndentFlags(dst, value, EncodeStandard, prefix, indent)
}

func AppendFastIndent(dst []byte, value any, prefix, indent string) ([]byte, error) {
	return AppendIndentFlags(dst, value, EncodeFastest, prefix, indent)
}

func AppendIndentFlags(dst []byte, value any, flags Flags, prefix, indent string) ([]byte, error) {
	buf := encodeBufferPool.Get().(*encodeBuffer)

	var err error
	buf.marshalBuf, err = encodeAny(buf.marshalBuf, value, flags)
	if err == nil {
		dst = zstr.AppendIndent(dst, buf.marshalBuf, prefix, indent)
	}

	buf.marshalBuf = buf.marshalBuf[:0]
	encodeBufferPool.Put(buf)

	return dst, err
}
