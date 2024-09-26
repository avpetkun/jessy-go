package jessy

import (
	"reflect"
	"sync"
	"unsafe"

	"github.com/avpetkun/jessy-go/zgo"
)

func MarshalPrecache(value any, flags Flags) {
	eface := zgo.UnpackEface(value)
	if eface.Type == nil {
		return
	}
	getTypeEncoder(eface.Type, flags)
}

func MarshalPrecacheFor[T any](flags Flags) {
	typ := zgo.TypeFor[T]()
	if typ == nil {
		return
	}
	getTypeEncoder(typ, flags)
}

func encodeAny(dst []byte, value any, flags Flags) ([]byte, error) {
	eface := zgo.UnpackEface(value)
	if eface.Type == nil {
		return append(dst, 'n', 'u', 'l', 'l'), nil
	}
	return getTypeEncoder(eface.Type, flags)(dst, eface.Data)
}

var encodersTypesCache [encodeFlagsLen]sync.Map

func getTypeEncoder(typ *zgo.Type, flags Flags) UnsafeEncoder {
	if val, ok := encodersTypesCache[flags].Load(typ); ok {
		return val.(UnsafeEncoder)
	}
	encoder := createTypeEncoder(0, 0, flags, typ.Native(), typ.IfaceIndir(), false)
	encodersTypesCache[flags].Store(typ, encoder)
	return encoder
}

func nopEncoder(dst []byte, v unsafe.Pointer) ([]byte, error) {
	return dst, nil
}

func nullEncoder(dst []byte, v unsafe.Pointer) ([]byte, error) {
	return append(dst, 'n', 'u', 'l', 'l'), nil
}

func createItemTypeEncoder(deep, indent uint32, flags Flags, t reflect.Type) UnsafeEncoder {
	ifaceIndir := t.Kind() == reflect.Pointer || zgo.RTypeIfaceIndir(t)
	return createTypeEncoder(deep, indent, flags, t, ifaceIndir, false)
}

func createTypeEncoder(deep, indent uint32, flags Flags, t reflect.Type, ifaceIndir, embedded bool) UnsafeEncoder {
	if t.Kind() == reflect.Pointer {
		return pointerEncoder(deep, indent, flags, t, ifaceIndir, embedded)
	}

	for i := range customEncoders {
		if customEncoders[i].Type == t {
			return customEncoders[i].Encoder(flags)
		}
	}

	if t == timeType {
		return timeEncoder(flags)
	}

	tp := reflect.PointerTo(t)
	switch {
	case tReallyImplements(t, typeAppendMarshaler):
		return appendMarshalerEncoder(t, flags)
	case tReallyImplements(tp, typeAppendMarshaler):
		return appendMarshalerEncoder(tp, flags)
	case tReallyImplements(t, typeMarshaler):
		return marshalerEncoder(t, flags)
	case tReallyImplements(tp, typeMarshaler):
		return marshalerEncoder(tp, flags)
	case tReallyImplements(t, typeTextMarshaler):
		return textMarshalerEncoder(t, flags)
	case tReallyImplements(tp, typeTextMarshaler):
		return textMarshalerEncoder(tp, flags)
	}

	switch t.Kind() {
	case reflect.Struct:
		return structEncoder(deep, indent, flags, t, ifaceIndir, embedded)
	case reflect.String:
		return stringEncoder(t, flags)
	case reflect.Map:
		return mapEncoder(deep, indent, t, flags, !ifaceIndir)
	case reflect.Slice:
		return sliceEncoder(deep, indent, t, flags)
	case reflect.Array:
		return arrayEncoder(deep, indent, t, flags)
	case reflect.Interface:
		return interfaceEncoder(flags)

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
	case reflect.Complex64:
		return complex64Encoder(flags)
	case reflect.Complex128:
		return complex128Encoder(flags)
	}

	return nopEncoder
}

func pointerEncoder(deep, indent uint32, flags Flags, t reflect.Type, ifaceIndir, embedded bool) UnsafeEncoder {
	omitEmpty := flags.Has(OmitEmpty)
	needQuotes := flags.Has(NeedQuotes)
	flags = flags.excludes(OmitEmpty)
	elemEncoder := createTypeEncoder(deep, indent, flags, t.Elem(), true, embedded)

	if ifaceIndir {
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			v = *(*unsafe.Pointer)(v)
			if v == nil {
				if needQuotes {
					return append(dst, '"', '"'), nil
				}
				if omitEmpty || embedded {
					return dst, nil
				}
				return append(dst, 'n', 'u', 'l', 'l'), nil
			}
			return elemEncoder(dst, v)
		}
	}
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		if v == nil {
			if needQuotes {
				return append(dst, '"', '"'), nil
			}
			if omitEmpty || embedded {
				return dst, nil
			}
			return append(dst, 'n', 'u', 'l', 'l'), nil
		}
		return elemEncoder(dst, v)
	}
}

func interfaceEncoder(flags Flags) UnsafeEncoder {
	return func(dst []byte, value unsafe.Pointer) ([]byte, error) {
		eface := (*zgo.EmptyInterface)(value)
		if eface.Type == nil {
			return append(dst, 'n', 'u', 'l', 'l'), nil
		}
		return getTypeEncoder(eface.Type, flags)(dst, eface.Data)
	}
}
