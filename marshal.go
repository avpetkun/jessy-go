package jessy

import (
	"bytes"
	"reflect"
	"slices"
	"sync"
	"unsafe"

	"github.com/avpetkun/jessy-go/zstr"
)

var (
	marshalMaxDeep uint32 = 20

	customEncoders []customEncoder
)

func SetMarshalMaxDeep(deep int) {
	if deep < 1 {
		panic("marshal max deep must be > 0")
	}
	marshalMaxDeep = uint32(deep)
	ResetEncodersCache()
}

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
	return MarshalFlags(value, EncodeStandard)
}

func MarshalFast(value any) ([]byte, error) {
	return MarshalFlags(value, EncodeFastest)
}

func MarshalPretty(value any) ([]byte, error) {
	return MarshalFlags(value, EncodeStandard|PrettySpaces)
}

func MarshalPrettyFast(value any) ([]byte, error) {
	return MarshalFlags(value, EncodeFastest|PrettySpaces)
}

func MarshalFlags(value any, flags Flags) (dst []byte, err error) {
	buf := getMarshalBuf()
	data, err := encodeAny(buf.AvailableBuffer(), value, flags)
	if err == nil {
		buf.Grow(len(data))
		dst = make([]byte, len(data))
		copy(dst, data)
	}
	putMarshalBuf(buf)
	return dst, err
}

func Append(dst []byte, value any) ([]byte, error) {
	return encodeAny(dst, value, EncodeStandard)
}

func AppendFast(dst []byte, value any) ([]byte, error) {
	return encodeAny(dst, value, EncodeFastest)
}

func AppendPretty(dst []byte, value any) ([]byte, error) {
	return encodeAny(dst, value, EncodeStandard|PrettySpaces)
}

func AppendPrettyFast(dst []byte, value any) ([]byte, error) {
	return encodeAny(dst, value, EncodeFastest|PrettySpaces)
}

func AppendFlags(dst []byte, value any, flags Flags) ([]byte, error) {
	return encodeAny(dst, value, flags)
}

func MarshalIndent(value any, prefix, indent string) ([]byte, error) {
	return AppendIndent(nil, value, prefix, indent)
}

func MarshalIndentFast(value any, prefix, indent string) ([]byte, error) {
	return AppendIndentFast(nil, value, prefix, indent)
}

func MarshalIndentFlags(value any, flags Flags, prefix, indent string) ([]byte, error) {
	return AppendIndentFlags(nil, value, flags, prefix, indent)
}

func AppendIndent(dst []byte, value any, prefix, indent string) ([]byte, error) {
	return AppendIndentFlags(dst, value, EncodeStandard, prefix, indent)
}

func AppendIndentFast(dst []byte, value any, prefix, indent string) ([]byte, error) {
	return AppendIndentFlags(dst, value, EncodeFastest, prefix, indent)
}

func AppendIndentFlags(dst []byte, value any, flags Flags, prefix, indent string) (data []byte, err error) {
	buf := getMarshalBuf()
	data = buf.AvailableBuffer()
	data, err = encodeAny(data, value, flags)
	if err == nil {
		dst = slices.Grow(dst, len(data)*2)
		dst = zstr.AppendIndent(dst, data, prefix, indent)
	}
	buf.Grow(len(data))
	putMarshalBuf(buf)
	return dst, err
}

var marshalBufPool = sync.Pool{New: func() any { return new(bytes.Buffer) }}

func getMarshalBuf() *bytes.Buffer {
	return marshalBufPool.Get().(*bytes.Buffer)
}

func putMarshalBuf(buf *bytes.Buffer) {
	buf.Reset()
	marshalBufPool.Put(buf)
}
