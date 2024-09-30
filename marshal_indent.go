package jessy

import (
	"bytes"
	"sync"

	"github.com/avpetkun/jessy-go/zstr"
)

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

func AppendIndentFlags(dst []byte, value any, flags Flags, prefix, indent string) (data []byte, err error) {
	buf := appendIndentBuf.Get().(*bytes.Buffer)
	data = buf.AvailableBuffer()
	data, err = encodeAny(data, value, flags)
	if err == nil {
		dst = zstr.AppendIndent(dst, data, prefix, indent)
	}
	buf.Grow(len(data))
	buf.Reset()
	appendIndentBuf.Put(buf)
	return dst, err
}

var appendIndentBuf = sync.Pool{New: func() any { return new(bytes.Buffer) }}
