package jessy

import (
	"unsafe"

	"github.com/avpetkun/jessy-go/zgo"
	"github.com/avpetkun/jessy-go/zstr"
)

func stringEncoder(flags Flags) UnsafeEncoder {
	omitEmpty := flags.Has(OmitEmpty)
	escapeHTML := flags.Has(EscapeHTML)
	needValidate := flags.Has(ValidateString) || escapeHTML

	if omitEmpty {
		if needValidate {
			return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
				h := (*zgo.String)(v)
				if h.Len == 0 {
					return dst, nil
				}
				data := unsafe.Slice(h.Data, h.Len)
				return zstr.AppendQuotedString(dst, data, escapeHTML), nil
			}
		}
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			h := (*zgo.String)(v)
			if h.Len == 0 {
				return dst, nil
			}
			data := unsafe.Slice(h.Data, h.Len)
			dst = append(dst, '"')
			dst = append(dst, data...)
			dst = append(dst, '"')
			return dst, nil
		}
	}

	if needValidate {
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			h := (*zgo.String)(v)
			if h.Len == 0 {
				return append(dst, '"', '"'), nil
			}
			data := unsafe.Slice(h.Data, h.Len)
			return zstr.AppendQuotedString(dst, data, escapeHTML), nil
		}
	}
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		h := (*zgo.String)(v)
		if h.Len == 0 {
			return append(dst, '"', '"'), nil
		}
		data := unsafe.Slice(h.Data, h.Len)
		dst = append(dst, '"')
		dst = append(dst, data...)
		dst = append(dst, '"')
		return dst, nil
	}
}
