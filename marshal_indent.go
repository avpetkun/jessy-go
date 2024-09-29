package jessy

import "github.com/avpetkun/jessy-go/zstr"

func MarshalIndent(value any, prefix, indent string) ([]byte, error) {
	return AppendIndent(nil, value, prefix, indent)
}

func MarshalIndentFast(value any, prefix, indent string) ([]byte, error) {
	return AppendIndentFast(nil, value, prefix, indent)
}

func AppendIndent(dst []byte, value any, prefix, indent string) ([]byte, error) {
	return AppendIndentFlags(dst, value, EncodeStandard, prefix, indent)
}

func AppendIndentFast(dst []byte, value any, prefix, indent string) ([]byte, error) {
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
