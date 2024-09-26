package jessy

import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/avpetkun/jessy-go/zgo"
	"github.com/avpetkun/jessy-go/zstr"
)

var typeJsonNumber = reflect.TypeFor[Number]()

//go:linkname isValidJsonNumber encoding/json.isValidNumber
func isValidJsonNumber(s string) bool

func stringEncoder(t reflect.Type, flags Flags) UnsafeEncoder {
	omitEmpty := flags.Has(OmitEmpty)
	escapeHTML := flags.Has(EscapeHTML)
	needQuotes := flags.Has(NeedQuotes)
	needValidate := flags.Has(ValidateString) || escapeHTML

	if t == typeJsonNumber {
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			h := (*zgo.String)(v)
			if h.Len == 0 {
				if omitEmpty {
					return dst, nil
				}
				return append(dst, '0'), nil
			}
			numStr := unsafe.String(h.Data, h.Len)
			if !isValidJsonNumber(numStr) {
				return dst, fmt.Errorf("json: invalid number literal %q", numStr)
			}
			if needQuotes {
				dst = append(dst, '"')
				dst = append(dst, numStr...)
				dst = append(dst, '"')
			} else {
				dst = append(dst, numStr...)
			}
			return dst, nil
		}
	}

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
