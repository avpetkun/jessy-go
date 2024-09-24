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
	typ := zgo.NewTypeFor[T]()
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
	return getTypeEncoder(eface.Type, flags)(dst, eface.Value)
}

var encodersTypesCache [encodeFlagsLen]sync.Map

func getTypeEncoder(typ *zgo.Type, flags Flags) UnsafeEncoder {
	if val, ok := encodersTypesCache[flags].Load(typ); ok {
		return val.(UnsafeEncoder)
	}
	encoder := createDirectTypeEncoder(flags, typ.Native())
	encodersTypesCache[flags].Store(typ, encoder)
	return encoder
}

func nopEncoder(dst []byte, v unsafe.Pointer) ([]byte, error) {
	return dst, nil
}

func createDirectTypeEncoder(flags Flags, t reflect.Type) UnsafeEncoder {
	tp := reflect.PointerTo(t)
	switch {
	case t.Implements(typeAppendMarshaler):
		return appendMarshalerEncoder(t)
	case tp.Implements(typeAppendMarshaler):
		return appendMarshalerEncoder(tp)
	case t.Implements(typeMarshaler):
		return marshalerEncoder(t)
	case tp.Implements(typeMarshaler):
		return marshalerEncoder(tp)
	case t.Implements(typeTextMarshaler):
		return textMarshalerEncoder(t, flags)
	case tp.Implements(typeTextMarshaler):
		return textMarshalerEncoder(tp, flags)
	}
	return createTypeEncoder(-1, flags, t, false, false, false)
}

func createItemTypeEncoder(deep int, flags Flags, t reflect.Type) UnsafeEncoder {
	if t.Kind() == reflect.Pointer {
		return pointerEncoder(deep, flags, t.Elem(), false, false, false)
	}
	return createTypeEncoder(deep, flags, t, false, false, false)
}

func createTypeEncoder(deep int, flags Flags, t reflect.Type, wasStruct, byPointer, embedded bool) UnsafeEncoder {
	if t.Kind() == reflect.Pointer {
		if t.Elem().Kind() == reflect.Struct {
			return createTypeEncoder(deep, flags, t.Elem(), wasStruct, true, embedded)
		}
		return pointerEncoder(deep, flags, t.Elem(), wasStruct, true, embedded)
	}

	if deep++; deep >= MarshalMaxDeep {
		return nopEncoder
	}

	for i := range customEncoders {
		if customEncoders[i].Type == t {
			return customEncoders[i].Encoder(flags)
		}
	}

	tp := reflect.PointerTo(t)
	switch {
	case tReallyImplements(t, typeAppendMarshaler):
		return appendMarshalerEncoder(t)
	case tReallyImplements(tp, typeAppendMarshaler):
		return appendMarshalerEncoder(tp)
	case tReallyImplements(t, typeMarshaler):
		return marshalerEncoder(t)
	case tReallyImplements(tp, typeMarshaler):
		return marshalerEncoder(tp)
	case tReallyImplements(t, typeTextMarshaler):
		return textMarshalerEncoder(t, flags)
	case tReallyImplements(tp, typeTextMarshaler):
		return textMarshalerEncoder(tp, flags)
	}

	switch t.Kind() {
	case reflect.Struct:
		return structEncoder(deep, flags, t, byPointer, embedded)
	case reflect.String:
		return stringEncoder(flags)
	case reflect.Map:
		if wasStruct {
			return pointerEncoder(deep, flags, t, false, false, false)
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
	case reflect.Complex64:
		return complex64Encoder(flags)
	case reflect.Complex128:
		return complex128Encoder(flags)

	case reflect.Interface:
		return func(dst []byte, value unsafe.Pointer) ([]byte, error) {
			eface := (*zgo.EmptyInterface)(value)
			if eface.Value == nil {
				return append(dst, 'n', 'u', 'l', 'l'), nil
			}
			return getTypeEncoder(eface.Type, flags)(dst, eface.Value)
		}
	}

	return nopEncoder
}

type StructField struct {
	Key     string
	KeyLen  int
	Null    string
	Offset  uintptr
	Encoder UnsafeEncoder
}

func structEncoder(deep int, flags Flags, t reflect.Type, byPointer, embedded bool) UnsafeEncoder {
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

		embedded := f.Anonymous && tReallyStruct(f.Type)
		if !embedded && !f.IsExported() {
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

		ft := f.Type

		makeUnpack := false
		if fieldsCount == 1 {
			if ft.Kind() == reflect.Pointer {
				ft = ft.Elem()
				makeUnpack = byPointer
			}
			byPointer = false
		} else {
			byPointer = ft.Kind() == reflect.Struct
			if !byPointer && ft.Kind() == reflect.Pointer && ft.Elem().Kind() == reflect.Struct {
				makeUnpack = true
			}
		}

		const wasStruct = true
		var fieldEncoder UnsafeEncoder
		if makeUnpack {
			fieldEncoder = pointerEncoder(deep, fieldFlags, ft, wasStruct, byPointer, embedded)
		} else {
			fieldEncoder = createTypeEncoder(deep, fieldFlags, ft, wasStruct, byPointer, embedded)
		}

		if embedded {
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

	if fields[0].KeyLen == 0 {
		fields[0].Null = "null"
	} else {
		fields[0].Null = "{" + fields[0].Key + "null}"
	}

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
			if value == nil {
				return append(dst, fields[0].Null...), nil
			}
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
		if value == nil {
			return append(dst, fields[0].Null...), nil
		}
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
			if value == nil {
				return append(dst, fields[0].Null...), nil
			}
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
		if value == nil {
			return append(dst, fields[0].Null...), nil
		}
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

func pointerEncoder(deep int, flags Flags, t reflect.Type, wasStruct, byPointer, embedded bool) UnsafeEncoder {
	omitEmpty := flags.Has(OmitEmpty)
	needQuotes := flags.Has(NeedQuotes)
	flags = flags.excludes(OmitEmpty)
	elemEncoder := createTypeEncoder(deep, flags, t, wasStruct, byPointer, embedded)

	if !wasStruct {
		switch t.Kind() {
		case reflect.Slice, reflect.String,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64,
			reflect.Complex64, reflect.Complex128,
			reflect.Bool:
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
	}
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

func marshalerEncoder(t reflect.Type) UnsafeEncoder {
	getInterface := zgo.NewInterfacerFromRType[Marshaler](t)
	if getInterface == nil {
		return nopEncoder
	}
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
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

	if needValidate {
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			data, err := getInterface(v).MarshalText()
			if err != nil {
				return dst, nil
			}
			return zstr.AppendQuotedString(dst, data, escapeHTML), nil
		}
	}
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		data, err := getInterface(v).MarshalText()
		if err != nil {
			return dst, nil
		}
		dst = append(dst, '"')
		dst = append(dst, data...)
		dst = append(dst, '"')
		return dst, nil
	}
}
