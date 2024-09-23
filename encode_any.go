package jessy

import (
	"reflect"
	"sync"
	"unsafe"

	"github.com/avpetkun/jessy-go/zgo"
	"github.com/avpetkun/jessy-go/zstr"
)

func encodeAny(dst []byte, value any, flags Flags) ([]byte, error) {
	eface := zgo.UnpackEface(value)
	if eface.Value == nil {
		return append(dst, 'n', 'u', 'l', 'l'), nil
	}
	return getTypeEncoder(eface.Type, flags)(dst, eface.Value)
}

var encodersTypesCache [encodeFlagsLen]sync.Map

func getTypeEncoder(typ *zgo.Type, flags Flags) UnsafeEncoder {
	if val, ok := encodersTypesCache[flags].Load(typ); ok {
		return val.(UnsafeEncoder)
	}
	encoder := createTypeEncoder(0, flags, typ.Native(), false, false, false)
	encodersTypesCache[flags].Store(typ, encoder)
	return encoder
}

func nopEncoder(dst []byte, v unsafe.Pointer) ([]byte, error) {
	return dst, nil
}

func panicEncoder(dst []byte, v unsafe.Pointer) ([]byte, error) {
	panic("it's panic encoder! smth was wrong")
}

func createTypeEncoderNested(deep uint, flags Flags, t reflect.Type) UnsafeEncoder {
	return createTypeEncoder(deep, flags, t, false, false, t.Kind() == reflect.Pointer)
}

func createTypeEncoder(deep uint, flags Flags, t reflect.Type, wasStruct, byPointer, doUnpack bool) UnsafeEncoder {
	if deep++; deep >= MarshalMaxDeep {
		return nopEncoder
	}

	if doUnpack {
		return pointerEncoder(deep, flags, t, wasStruct, byPointer, doUnpack)
	}

	/*switch {
	case t.Implements(typeAppendMarshaler):
		return appendMarshalerEncoder(t)
	case t.Implements(typeMarshaler):
		return marshalerEncoder(t, unpack && t.Kind() == reflect.Pointer)
	case t.Implements(typeTextMarshaler):
		return textMarshalerEncoder(t, flags)
	}*/

	switch t.Kind() {
	case reflect.Pointer:
		doUnpack = t.Elem().Kind() != reflect.Struct
		return createTypeEncoder(deep, flags, t.Elem(), wasStruct, true, doUnpack)
	case reflect.Struct:
		fieldsCount := t.NumField()
		if fieldsCount == 0 {
			return func(dst []byte, value unsafe.Pointer) ([]byte, error) {
				return append(dst, '{', '}'), nil
			}
		}
		return func(dst []byte, value unsafe.Pointer) ([]byte, error) {
			dst = append(dst, '{')
			for i := range fieldsCount {
				f := t.Field(i)
				ft := f.Type
				doUnpack = false
				if fieldsCount == 1 {
					if ft.Kind() == reflect.Pointer {
						ft = ft.Elem()
						doUnpack = byPointer
					}
					byPointer = false
				} else {
					byPointer = ft.Kind() == reflect.Struct
					if !byPointer && ft.Kind() == reflect.Pointer && ft.Elem().Kind() == reflect.Struct {
						doUnpack = true
					}
				}
				fValue := unsafe.Add(value, f.Offset)

				if i != 0 {
					dst = append(dst, ',')
				}
				dst = append(dst, '"')
				dst = append(dst, f.Name...)
				dst = append(dst, '"', ':')
				dst, _ = createTypeEncoder(deep, flags, ft, true, byPointer, doUnpack)(dst, fValue)
			}
			dst = append(dst, '}')
			return dst, nil
		}

	case reflect.String:
		return stringEncoder(flags)
	case reflect.Map:
		if wasStruct {
			return pointerEncoder(deep, flags, t, false, false, true)
		}
		return mapEncoder(deep, t, flags)
	case reflect.Slice:
		return sliceEncoder(deep, t, flags)
	case reflect.Array:
		return arrayEncoder(deep, t, flags)

	case reflect.Bool:
		return boolEncoder(flags)
	case reflect.Int:
		return intEncoder(flags)
	case reflect.Int8:
		return int8Encoder(flags)
	case reflect.Int16:
		return int16Encoder(flags)
	case reflect.Int32:
		return int32Encoder(flags)
	case reflect.Int64:
		return int64Encoder(flags)
	case reflect.Uint:
		return uintEncoder(flags)
	case reflect.Uint8:
		return uint8Encoder(flags)
	case reflect.Uint16:
		return uint16Encoder(flags)
	case reflect.Uint32:
		return uint32Encoder(flags)
	case reflect.Uint64:
		return uint64Encoder(flags)
	case reflect.Float32:
		return float32Encoder(flags)
	case reflect.Float64:
		return float64Encoder(flags)

	case reflect.Interface:
		return func(dst []byte, value unsafe.Pointer) ([]byte, error) {
			eface := (*zgo.EmptyInterface)(value)
			if eface.Value == nil {
				return append(dst, 'n', 'u', 'l', 'l'), nil
			}
			return getTypeEncoder(eface.Type, flags)(dst, eface.Value)
		}
	}

	return panicEncoder
}

func pointerEncoder(deep uint, flags Flags, t reflect.Type, wasStruct, byPointer, doUnpack bool) UnsafeEncoder {
	omitEmpty := flags.Has(OmitEmpty)
	flags = flags.excludes(OmitEmpty)
	elemEncoder := createTypeEncoder(deep, flags, t, wasStruct, byPointer, false)

	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		if doUnpack {
			v = *(*unsafe.Pointer)(v)
		}
		if v == nil {
			if omitEmpty {
				return dst, nil
			}
			return append(dst, 'n', 'u', 'l', 'l'), nil
		}
		return elemEncoder(dst, v)
	}
}

func marshalerEncoder(t reflect.Type, unpack bool) UnsafeEncoder {
	getInterface := zgo.NewInterfacerFromRType[Marshaler](t)
	if getInterface == nil {
		return nopEncoder
	}
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		if unpack {
			v = *(*unsafe.Pointer)(v)
		}
		if v == nil {
			return append(dst, "null"...), nil
		}
		data, err := getInterface(v).MarshalJSON()
		if err != nil {
			return dst, err
		}
		return append(dst, data...), nil
	}
}

func appendMarshalerEncoder(t reflect.Type) UnsafeEncoder {
	getInterface := zgo.NewInterfacerFromRType[AppendMarshaler](t)
	if getInterface == nil {
		return nopEncoder
	}
	return func(dst []byte, v unsafe.Pointer) (newDst []byte, err error) {
		newDst, err = getInterface(v).AppendMarshalJSON(dst)
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
		return nopEncoder
	}

	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		data, err := getInterface(v).MarshalText()
		if err != nil {
			return dst, nil
		}
		if needValidate {
			return zstr.AppendQuotedString(dst, data, escapeHTML), nil
		}
		dst = append(dst, '"')
		dst = append(dst, data...)
		dst = append(dst, '"')
		return dst, nil
	}
}
