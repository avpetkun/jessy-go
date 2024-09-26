package jessy

import (
	"reflect"
	"unsafe"

	"github.com/avpetkun/jessy-go/zgo"
	"github.com/avpetkun/jessy-go/zstr"
)

func marshalerEncoder(t reflect.Type, flags Flags) UnsafeEncoder {
	getInterface := zgo.NewInterfacerFromRType[Marshaler](t)
	if getInterface == nil {
		return nullEncoder
	}
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		i := getInterface(v)
		if i == nil {
			if flags.Has(OmitEmpty) {
				return dst, nil
			}
			return append(dst, 'n', 'u', 'l', 'l'), nil
		}
		data, err := i.MarshalJSON()
		if err != nil {
			return dst, err
		}
		return append(dst, data...), nil
	}
}

func appendMarshalerEncoder(t reflect.Type, flags Flags) UnsafeEncoder {
	getInterface := zgo.NewInterfacerFromRType[AppendMarshaler](t)
	if getInterface == nil {
		return nullEncoder
	}
	return func(dst []byte, v unsafe.Pointer) (newDst []byte, err error) {
		i := getInterface(v)
		if i == nil {
			if flags.Has(OmitEmpty) {
				return dst, nil
			}
			return append(dst, 'n', 'u', 'l', 'l'), nil
		}
		newDst, err = i.AppendMarshalJSON(dst)
		if err != nil {
			return dst, nil
		}
		return newDst, nil
	}
}

func textMarshalerEncoder(t reflect.Type, flags Flags) UnsafeEncoder {
	escapeHTML := flags.Has(EscapeHTML)
	needValidate := flags.Has(ValidateTextMarshaller) || escapeHTML

	getInterface := zgo.NewInterfacerFromRType[TextMarshaler](t)
	if getInterface == nil {
		return nullEncoder
	}

	if needValidate {
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			i := getInterface(v)
			if i == nil {
				if flags.Has(OmitEmpty) {
					return dst, nil
				}
				return append(dst, 'n', 'u', 'l', 'l'), nil
			}
			data, err := i.MarshalText()
			if err != nil {
				return dst, nil
			}
			return zstr.AppendQuotedString(dst, data, escapeHTML), nil
		}
	}
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		i := getInterface(v)
		if i == nil {
			if flags.Has(OmitEmpty) {
				return dst, nil
			}
			return append(dst, 'n', 'u', 'l', 'l'), nil
		}
		data, err := i.MarshalText()
		if err != nil {
			return dst, nil
		}
		dst = append(dst, '"')
		dst = append(dst, data...)
		dst = append(dst, '"')
		return dst, nil
	}
}
