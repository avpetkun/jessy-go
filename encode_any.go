package jessy

import (
	"reflect"
	"sort"
	"strings"
	"sync"
	"unsafe"

	"github.com/avpetkun/jessy-go/zgo"
	"github.com/avpetkun/jessy-go/zstr"
)

func Precache(value any, flags Flags) {
	eface := zgo.UnpackEface(value)
	if eface.Type == nil {
		return
	}
	getTypeEncoder(eface.Type, flags)
}

func PrecacheFor[T any](flags Flags) {
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
	encoder := createTypeEncoder(0, flags, typ.Native(), typ.IfaceIndir(), false)
	encodersTypesCache[flags].Store(typ, encoder)
	return encoder
}

func nopEncoder(dst []byte, v unsafe.Pointer) ([]byte, error) {
	return dst, nil
}

func nullEncoder(dst []byte, v unsafe.Pointer) ([]byte, error) {
	return append(dst, 'n', 'u', 'l', 'l'), nil
}

func createItemTypeEncoder(deep int, flags Flags, t reflect.Type) UnsafeEncoder {
	ifaceIndir := t.Kind() == reflect.Pointer || zgo.RTypeIfaceIndir(t)
	return createTypeEncoder(deep, flags, t, ifaceIndir, false)
}

func createTypeEncoder(deep int, flags Flags, t reflect.Type, ifaceIndir, embedded bool) UnsafeEncoder {
	if deep++; deep >= MarshalMaxDeep {
		return nopEncoder
	}

	if t.Kind() == reflect.Pointer {
		return pointerEncoder(deep, flags, t, ifaceIndir, embedded)
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
		return structEncoder(deep, flags, t, ifaceIndir, embedded)
	case reflect.String:
		return stringEncoder(flags)
	case reflect.Map:
		return mapEncoder(deep, t, flags, !ifaceIndir)
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
	case reflect.Complex64:
		return complex64Encoder(flags)
	case reflect.Complex128:
		return complex128Encoder(flags)

	case reflect.Interface:
		return func(dst []byte, value unsafe.Pointer) ([]byte, error) {
			eface := (*zgo.EmptyInterface)(value)
			if eface.Type == nil {
				return append(dst, 'n', 'u', 'l', 'l'), nil
			}
			return getTypeEncoder(eface.Type, flags)(dst, eface.Data)
		}
	}

	return nopEncoder
}

type StructField struct {
	Key     string
	KeyLen  int
	Offset  uintptr
	Encoder UnsafeEncoder
}

func structEncoder(deep int, flags Flags, t reflect.Type, ifaceIndir, embedded bool) UnsafeEncoder {
	prettySpaces := flags.Has(PrettySpaces)

	fieldsCount := t.NumField()
	if fieldsCount == 0 {
		return func(dst []byte, value unsafe.Pointer) ([]byte, error) {
			return append(dst, '{', '}'), nil
		}
	}

	fields := make([]StructField, 0, fieldsCount)
	for i := range fieldsCount {
		f := t.Field(i)

		anonymousStruct := f.Anonymous && tReallyStruct(f.Type) && !tImplementsAny(f.Type)
		if !f.IsExported() && !anonymousStruct {
			continue
		}

		parts := strings.Split(f.Tag.Get("json"), ",")
		name := parts[0]
		if name == "-" {
			continue
		}
		if name == "" {
			name = f.Name
		}

		fieldFlags := flags
		for _, action := range parts[1:] {
			switch action {
			case "omitempty":
				fieldFlags |= OmitEmpty
			case "string":
				fieldFlags |= NeedQuotes
			}
		}

		fieldEncoder := createTypeEncoder(deep, fieldFlags, f.Type, ifaceIndir, anonymousStruct)

		if anonymousStruct {
			fields = append(fields, StructField{
				KeyLen:  0,
				Offset:  f.Offset,
				Encoder: fieldEncoder,
			})
		} else {
			key := `"` + name + `":`
			if prettySpaces {
				key += " "
			}
			fields = append(fields, StructField{
				Key:     key,
				KeyLen:  len(key),
				Offset:  f.Offset,
				Encoder: fieldEncoder,
			})
		}
	}

	if len(fields) == 0 {
		return nopStructEncoder
	}

	sort.Slice(fields, func(i, j int) bool {
		return fields[i].Key < fields[j].Key
	})

	if prettySpaces {
		return structEncoderPretty(deep, fields, embedded)
	}
	return structEncoderMinimal(fields, embedded)
}

func nopStructEncoder(dst []byte, value unsafe.Pointer) ([]byte, error) {
	if value == nil {
		return append(dst, 'n', 'u', 'l', 'l'), nil
	}
	return append(dst, '{', '}'), nil
}

func structEncoderPretty(deep int, fields []StructField, embedded bool) UnsafeEncoder {
	if embedded && deep > 0 {
		deep--
	}
	deepSpace0 := strings.Repeat("\t", deep)
	deepSpace1 := strings.Repeat("\t", deep+1)
	if embedded {
		return func(dst []byte, value unsafe.Pointer) ([]byte, error) {
			var err error
			for i := range fields {
				dst = append(dst, deepSpace0...)
				dst = append(dst, fields[i].Key...)
				dstLen := len(dst)
				dst, err = fields[i].Encoder(dst, unsafe.Add(value, fields[i].Offset))
				if err != nil {
					return dst, err
				}
				if len(dst) == dstLen {
					dst = dst[:dstLen-fields[i].KeyLen-deep]
				} else {
					dst = append(dst, ',', '\n')
				}
			}
			if i := len(dst); i != 0 && dst[i-1] == '\n' {
				dst = dst[:i-2]
			}
			return dst, nil
		}
	}
	return func(dst []byte, value unsafe.Pointer) ([]byte, error) {
		dst = append(dst, '{', '\n')
		var err error
		for i := range fields {
			dst = append(dst, deepSpace1...)
			dst = append(dst, fields[i].Key...)
			dstLen := len(dst)
			dst, err = fields[i].Encoder(dst, unsafe.Add(value, fields[i].Offset))
			if err != nil {
				return dst, err
			}
			if len(dst) == dstLen {
				dst = dst[:dstLen-fields[i].KeyLen-deep-1]
			} else {
				dst = append(dst, ',', '\n')
			}
		}
		if i := len(dst) - 2; dst[i] == ',' {
			dst = dst[:i]
		}
		dst = append(dst, '\n')
		dst = append(dst, deepSpace0...)
		dst = append(dst, '}')
		return dst, nil
	}
}

func structEncoderMinimal(fields []StructField, embedded bool) UnsafeEncoder {
	if embedded {
		return func(dst []byte, value unsafe.Pointer) ([]byte, error) {
			var err error
			for i := range fields {
				dst = append(dst, fields[i].Key...)
				dstLen := len(dst)
				dst, err = fields[i].Encoder(dst, unsafe.Add(value, fields[i].Offset))
				if err != nil {
					return dst, err
				}
				if len(dst) == dstLen {
					dst = dst[:dstLen-fields[i].KeyLen]
				} else {
					dst = append(dst, ',')
				}
			}
			if i := len(dst); i != 0 && dst[i-1] == ',' {
				dst = dst[:i-1]
			}
			return dst, nil
		}
	}
	return func(dst []byte, value unsafe.Pointer) ([]byte, error) {
		dst = append(dst, '{')
		var err error
		for i := range fields {
			dst = append(dst, fields[i].Key...)
			dstLen := len(dst)
			dst, err = fields[i].Encoder(dst, unsafe.Add(value, fields[i].Offset))
			if err != nil {
				return dst, err
			}
			if len(dst) == dstLen {
				dst = dst[:dstLen-fields[i].KeyLen]
			} else {
				dst = append(dst, ',')
			}
		}
		if i := len(dst) - 1; dst[i] == ',' {
			dst[i] = '}'
		} else {
			dst = append(dst, '}')
		}
		return dst, nil
	}
}

func pointerEncoder(deep int, flags Flags, t reflect.Type, ifaceIndir, embedded bool) UnsafeEncoder {
	omitEmpty := flags.Has(OmitEmpty)
	needQuotes := flags.Has(NeedQuotes)
	flags = flags.excludes(OmitEmpty)
	elemEncoder := createTypeEncoder(deep, flags, t.Elem(), true, embedded)

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
