package jessy

import (
	"errors"
	"fmt"
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
	omitEmpty := flags.Has(OmitEmpty)
	escapeHTML := flags.Has(EscapeHTML)

	if flags.Has(CompactMarshaler) {
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			i := getInterface(v)
			if i == nil {
				if omitEmpty {
					return dst, nil
				}
				return append(dst, 'n', 'u', 'l', 'l'), nil
			}
			data, err := i.MarshalJSON()
			if err != nil {
				return dst, errors.Join(fmt.Errorf("failed to call MarshalJSON of type <%s>", t), err)
			}
			return zstr.AppendCompactJSON(dst, data, escapeHTML), nil
		}
	}
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		i := getInterface(v)
		if i == nil {
			if omitEmpty {
				return dst, nil
			}
			return append(dst, 'n', 'u', 'l', 'l'), nil
		}
		data, err := i.MarshalJSON()
		if err != nil {
			return dst, errors.Join(fmt.Errorf("failed to call MarshalJSON of type <%s>", t), err)
		}
		return append(dst, data...), nil
	}
}

func appendMarshalerEncoder(t reflect.Type, flags Flags) UnsafeEncoder {
	omitEmpty := flags.Has(OmitEmpty)
	getInterface := zgo.NewInterfacerFromRType[AppendMarshaler](t)
	if getInterface == nil {
		return nullEncoder
	}
	return func(dst []byte, v unsafe.Pointer) (newDst []byte, err error) {
		i := getInterface(v)
		if i == nil {
			if omitEmpty {
				return dst, nil
			}
			return append(dst, 'n', 'u', 'l', 'l'), nil
		}
		newDst, err = i.AppendMarshalJSON(dst)
		if err != nil {
			return dst, errors.Join(fmt.Errorf("failed to call AppendMarshalJSON of type <%s>", t), err)
		}
		return newDst, nil
	}
}

func textMarshalerEncoder(t reflect.Type, flags Flags) UnsafeEncoder {
	omitEmpty := flags.Has(OmitEmpty)
	escapeHTML := flags.Has(EscapeHTML)
	needValidate := flags.Has(ValidateTextMarshaler) || escapeHTML

	getInterface := zgo.NewInterfacerFromRType[TextMarshaler](t)
	if getInterface == nil {
		return nullEncoder
	}

	if needValidate {
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			i := getInterface(v)
			if i == nil {
				if omitEmpty {
					return dst, nil
				}
				return append(dst, 'n', 'u', 'l', 'l'), nil
			}
			data, err := i.MarshalText()
			if err != nil {
				return dst, errors.Join(fmt.Errorf("failed to call MarshalText of type <%s>", t), err)
			}
			return zstr.AppendQuotedString(dst, data, escapeHTML), nil
		}
	}
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		i := getInterface(v)
		if i == nil {
			if omitEmpty {
				return dst, nil
			}
			return append(dst, 'n', 'u', 'l', 'l'), nil
		}
		data, err := i.MarshalText()
		if err != nil {
			return dst, errors.Join(fmt.Errorf("failed to call MarshalText of type <%s>", t), err)
		}
		dst = append(dst, '"')
		dst = append(dst, data...)
		dst = append(dst, '"')
		return dst, nil
	}
}
