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
	encoder := createDirectTypeEncoder(flags, typ.Native())
	encodersTypesCache[flags].Store(typ, encoder)
	return encoder
}

func nopEncoder(dst []byte, v unsafe.Pointer) ([]byte, error) {
	return dst, nil
}

func createItemTypeEncoder(deep uint, flags Flags, t reflect.Type) UnsafeEncoder {
	return createTypeEncoder(deep, flags, t, false, false, t.Kind() == reflect.Pointer)
}

func tReallyImplements(t, inter reflect.Type) bool {
	if t.Implements(inter) {
		if t.Kind() == reflect.Struct {
			for i := range t.NumField() {
				f := t.Field(i)
				if f.Anonymous && f.Type.Implements(inter) {
					return false
				}
			}
		}
		return true
	}
	return false
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
	return createTypeEncoder(0, flags, t, false, false, false)
}

func createTypeEncoder(deep uint, flags Flags, t reflect.Type, wasStruct, byPointer, doUnpack bool) UnsafeEncoder {
	if deep++; deep >= MarshalMaxDeep {
		return nopEncoder
	}

	if doUnpack {
		return pointerEncoder(deep, flags, t, wasStruct, byPointer, doUnpack)
	}

	if t.Kind() == reflect.Pointer {
		byPointer = true
		doUnpack = t.Elem().Kind() != reflect.Struct
		return createTypeEncoder(deep, flags, t.Elem(), wasStruct, byPointer, doUnpack)
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
		return structEncoder(deep, flags, t, byPointer)
	case reflect.String:
		return stringEncoder(flags)
	case reflect.Map:
		if wasStruct {
			wasStruct, byPointer, doUnpack = false, false, true
			return pointerEncoder(deep, flags, t, wasStruct, byPointer, doUnpack)
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

	return nopEncoder
}

func structEncoder(deep uint, flags Flags, t reflect.Type, byPointer bool) UnsafeEncoder {
	wasStruct := true
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
			doUnpack := false
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
			dst, _ = createTypeEncoder(deep, flags, ft, wasStruct, byPointer, doUnpack)(dst, fValue)
		}
		dst = append(dst, '}')
		return dst, nil
	}
}

/*func structEncoderOld(deep, offset uint, t reflect.Type, flags Flags, inEmbedded bool) UnsafeEncoder {
	if deep++; deep >= MarshalMaxDeep {
		return nopEncoder
	}
	type Field struct {
		Key     string
		KeyLen  int
		Encoder UnsafeEncoder
	}
	fields := []Field{}
	for i := range t.NumField() {
		f := t.Field(i)

		name := f.Tag.Get("json")
		action := ""
		if j := strings.IndexByte(name, ','); j != -1 {
			action = name[j+1:]
			name = name[:j]
		}
		if name == "-" {
			continue
		} else if name == "" {
			name = f.Name
		}
		fieldFlags := flags
		if action == "omitempty" {
			fieldFlags |= OmitEmpty
		}
		if f.Anonymous {
			fields = append(fields, Field{
				KeyLen:  0,
				Encoder: getEmbeddedStructEncoder(deep, uint(f.Offset), f.Type, fieldFlags),
			})
		} else if f.IsExported() {
			key := `"` + name + `":`
			fields = append(fields, Field{
				Key:     key,
				KeyLen:  len(key),
				Encoder: getValueEncoder(deep, uint(f.Offset), f.Type, fieldFlags),
			})
		}
	}
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].Key < fields[j].Key
	})
	if inEmbedded {
		return func(dst []byte, v unsafe.Pointer) (_ []byte, err error) {
			v = unsafe.Add(v, offset)
			was := 0
			for i := range fields {
				if was != 0 {
					dst = append(dst, ',')
				}
				dst = append(dst, fields[i].Key...)
				dstLen := len(dst)
				dst, err = fields[i].Encoder(dst, v)
				if err != nil {
					return dst, err
				}
				if len(dst) == dstLen {
					dst = dst[:dstLen-fields[i].KeyLen-was]
				} else {
					was = 1
				}
			}
			return dst, nil
		}
	}
	return func(dst []byte, v unsafe.Pointer) (_ []byte, err error) {
		v = unsafe.Add(v, offset)
		dst = append(dst, '{')
		was := 0
		for i := range fields {
			if was != 0 {
				dst = append(dst, ',')
			}
			dst = append(dst, fields[i].Key...)
			dstLen := len(dst)
			dst, err = fields[i].Encoder(dst, v)
			if err != nil {
				return dst, err
			}
			if len(dst) == dstLen {
				dst = dst[:dstLen-fields[i].KeyLen-was]
			} else {
				was = 1
			}
		}
		dst = append(dst, '}')
		return dst, nil
	}
}*/

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
