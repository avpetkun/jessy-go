package jessy

import (
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"unsafe"

	"github.com/avpetkun/jessy-go/zgo"
	"github.com/avpetkun/jessy-go/zstr"
)

var (
	MarshalMaxDeep = 10

	customEncoders []customEncoder
)

type MarshalFlags byte

const (
	MarshalDefault   MarshalFlags = 0
	MarshalEmbedded  MarshalFlags = 1 << 0
	MarshalOmitEmpty MarshalFlags = 1 << 1
	MarshalQuote     MarshalFlags = 1 << 2
	MarshalQuoted    MarshalFlags = 1 << 3
)

type UnsafeEncoder func(dst []byte, value unsafe.Pointer) ([]byte, error)
type ValueEncoder[T any] func(dst []byte, value T) ([]byte, error)

type customEncoder struct {
	reflect.Type
	Encoder func(flags MarshalFlags) UnsafeEncoder
}

func AddUnsafeEncoder[T any](encoder func(flags MarshalFlags) UnsafeEncoder) {
	customEncoders = append(customEncoders, customEncoder{
		Type:    reflect.TypeFor[T](),
		Encoder: encoder,
	})
}

func AddValueEncoder[T any](encoder func(flags MarshalFlags) ValueEncoder[T]) {
	AddUnsafeEncoder[T](func(flags MarshalFlags) UnsafeEncoder {
		valEnc := encoder(flags)
		return func(dst []byte, value unsafe.Pointer) ([]byte, error) {
			return valEnc(dst, *(*T)(value))
		}
	})
}

func Marshal(value any) (data []byte, err error) {
	return AppendMarshal(nil, value)
}

func AppendMarshal(dst []byte, value any) (data []byte, err error) {
	eface := zgo.UnpackEface(value)
	enc := getValTypeEncoder(eface.Type)
	return enc(dst, eface.Value)
}

var encodersValCache sync.Map

func getValTypeEncoder(typ *zgo.Type) UnsafeEncoder {
	if val, ok := encodersValCache.Load(typ); ok {
		return val.(UnsafeEncoder)
	}
	t := typ.Native()
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	enc := getValEncoder(0, 0, t, MarshalDefault)
	encodersValCache.Store(typ, enc)
	return enc
}

func getValEncoder(deep, offset int, t reflect.Type, flags MarshalFlags) UnsafeEncoder {
	if deep++; deep == MarshalMaxDeep {
		return nopEncoder
	}
	if t.Kind() == reflect.Pointer {
		return pointerEncoder(deep, offset, t, flags)
	}
	for i := range customEncoders {
		if customEncoders[i].Type == t {
			enc := customEncoders[i].Encoder(flags)
			return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
				return enc(dst, unsafe.Add(v, offset))
			}
		}
	}
	{
		tp := reflect.PointerTo(t)
		switch {
		case t.Implements(typeAppendMarshaler):
			return appendMarshalerEncoder(offset, t)
		case tp.Implements(typeAppendMarshaler):
			return appendMarshalerEncoder(offset, tp)
		case t.Implements(typeMarshaler):
			return marshalerEncoder(offset, t)
		case tp.Implements(typeMarshaler):
			return marshalerEncoder(offset, tp)
		case t.Implements(typeTextMarshaler):
			return textMarshalerEncoder(offset, t)
		case tp.Implements(typeTextMarshaler):
			return textMarshalerEncoder(offset, tp)
		}
	}
	switch t.Kind() {
	case reflect.Struct:
		return structEncoder(deep, offset, t, flags)
	case reflect.Map:
		return mapEncoder(deep, offset, t, flags)
	case reflect.Array:
		return arrayEncoder(deep, offset, t, flags)
	case reflect.Slice:
		return sliceEncoder(deep, offset, t, flags)
	case reflect.String:
		return stringEncoder(offset, flags)
	case reflect.Bool:
		return boolEncoder(offset, flags)
	case reflect.Int:
		return intEncoder(offset, flags)
	case reflect.Int8:
		return int8Encoder(offset, flags)
	case reflect.Int16:
		return int16Encoder(offset, flags)
	case reflect.Int32:
		return int32Encoder(offset, flags)
	case reflect.Int64:
		return int64Encoder(offset, flags)
	case reflect.Uint:
		return uintEncoder(offset, flags)
	case reflect.Uint8:
		return uint8Encoder(offset, flags)
	case reflect.Uint16:
		return uint16Encoder(offset, flags)
	case reflect.Uint32:
		return uint32Encoder(offset, flags)
	case reflect.Uint64:
		return uint64Encoder(offset, flags)
	case reflect.Float32:
		return float32Encoder(offset, flags)
	case reflect.Float64:
		return float64Encoder(offset, flags)
	case reflect.Interface:
		return func(dst []byte, value unsafe.Pointer) ([]byte, error) {
			eface := (*zgo.EmptyInterface)(unsafe.Add(value, offset))
			return getValTypeEncoder(eface.Type)(dst, eface.Value)
		}
	default:
		return nopEncoder
	}
}

var encodersKeyCache sync.Map

func getKeyTypeEncoder(typ *zgo.Type) UnsafeEncoder {
	if val, ok := encodersKeyCache.Load(typ); ok {
		return val.(UnsafeEncoder)
	}
	t := typ.Native()
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	enc := getKeyEncoder(t)
	encodersKeyCache.Store(typ, enc)
	return enc
}

func getKeyEncoder(t reflect.Type) UnsafeEncoder {
	if t.Kind() == reflect.Pointer {
		return keyPointerEncoder(t)
	}
	for i := range customEncoders {
		if customEncoders[i].Type == t {
			return customEncoders[i].Encoder(MarshalQuote)
		}
	}
	if t.Implements(typeTextMarshaler) {
		return textMarshalerEncoder(0, t)
	} else if tp := reflect.PointerTo(t); tp.Implements(typeTextMarshaler) {
		return textMarshalerEncoder(0, tp)
	}
	switch t.Kind() {
	case reflect.Array:
		if t.Elem().Kind() == reflect.Uint8 {
			return arrayByteHexEncoder(0, uintptr(t.Len()), MarshalDefault)
		}
	case reflect.String:
		return stringEncoder(0, MarshalDefault)
	case reflect.Bool:
		return boolEncoder(0, MarshalDefault)
	case reflect.Int:
		return intEncoder(0, MarshalDefault)
	case reflect.Int8:
		return int8Encoder(0, MarshalDefault)
	case reflect.Int16:
		return int16Encoder(0, MarshalDefault)
	case reflect.Int32:
		return int32Encoder(0, MarshalDefault)
	case reflect.Int64:
		return int64Encoder(0, MarshalDefault)
	case reflect.Uint:
		return uintEncoder(0, MarshalDefault)
	case reflect.Uint8:
		return uint8Encoder(0, MarshalDefault)
	case reflect.Uint16:
		return uint16Encoder(0, MarshalDefault)
	case reflect.Uint32:
		return uint32Encoder(0, MarshalDefault)
	case reflect.Uint64:
		return uint64Encoder(0, MarshalDefault)
	case reflect.Float32:
		return float32Encoder(0, MarshalDefault)
	case reflect.Float64:
		return float64Encoder(0, MarshalDefault)
	case reflect.Interface:
		return func(dst []byte, value unsafe.Pointer) ([]byte, error) {
			eface := (*zgo.EmptyInterface)(value)
			return getKeyTypeEncoder(eface.Type)(dst, eface.Value)
		}
	}
	return nopEncoder
}

func nopEncoder(dst []byte, v unsafe.Pointer) ([]byte, error) {
	return dst, nil
}

func structEncoder(deep, offset int, t reflect.Type, flags MarshalFlags) UnsafeEncoder {
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
		var fieldFlags MarshalFlags
		if action == "omitempty" {
			fieldFlags |= MarshalOmitEmpty
		}
		if f.Anonymous {
			fields = append(fields, Field{
				KeyLen:  1,
				Encoder: getValEncoder(deep, int(f.Offset), f.Type, (fieldFlags | MarshalEmbedded)),
			})
		} else if f.IsExported() {
			key := `"` + name + `":`
			fields = append(fields, Field{
				Key:     key,
				KeyLen:  len(key),
				Encoder: getValEncoder(deep, int(f.Offset), f.Type, fieldFlags),
			})
		}
	}
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].Key < fields[j].Key
	})
	return func(dst []byte, v unsafe.Pointer) (_ []byte, err error) {
		v = unsafe.Add(v, offset)
		if flags&MarshalEmbedded == 0 {
			dst = append(dst, '{')
		}
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
		if flags&MarshalEmbedded == 0 {
			dst = append(dst, '}')
		}
		return dst, nil
	}
}

func mapEncoder(deep, offset int, t reflect.Type, flags MarshalFlags) UnsafeEncoder {
	encodeKey := getKeyEncoder(t.Key())
	encodeVal := getValEncoder(deep, 0, t.Elem(), MarshalDefault)
	getIterator := zgo.NewPointerMapIteratorForType(t)

	return func(dst []byte, value unsafe.Pointer) ([]byte, error) {
		it, count := getIterator(unsafe.Add(value, offset))
		if it == nil {
			if flags&MarshalOmitEmpty != 0 {
				return dst, nil
			}
			return append(dst, 'n', 'u', 'l', 'l'), nil
		}
		if count == 0 {
			it.Release()
			return append(dst, '{', '}'), nil
		}
		var err error
		dst = append(dst, '{')
		was := 0
		for range count {
			if was != 0 {
				dst = append(dst, ',')
			}
			dstLen0 := len(dst)
			dst, err = encodeKey(dst, it.Key)
			if err != nil {
				it.Release()
				return dst, err
			}
			if len(dst) == dstLen0 {
				dst = dst[:dstLen0-was]
				continue
			}
			dst = append(dst, ':')
			dstLen1 := len(dst)
			dst, err = encodeVal(dst, it.Elem)
			if err != nil {
				it.Release()
				return dst, err
			}
			if len(dst) == dstLen1 {
				dst = dst[:dstLen0-was]
				continue
			}

			was = 1
			it.Next()
		}
		it.Release()
		dst = append(dst, '}')
		return dst, nil
	}
}

func keyPointerEncoder(t reflect.Type) UnsafeEncoder {
	elemEncoder := getKeyEncoder(t.Elem())
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		vp := *(*uintptr)(v)
		if vp == 0 {
			return append(dst, 'n', 'u', 'l', 'l'), nil
		}
		return elemEncoder(dst, unsafe.Pointer(vp))
	}
}

func pointerEncoder(deep, offset int, t reflect.Type, flags MarshalFlags) UnsafeEncoder {
	elemEncoder := getValEncoder(deep, 0, t.Elem(), flags)
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		v = unsafe.Add(v, offset)
		vp := *(*uintptr)(v)
		if vp == 0 {
			if flags&MarshalOmitEmpty != 0 {
				return dst, nil
			}
			return append(dst, 'n', 'u', 'l', 'l'), nil
		}
		return elemEncoder(dst, unsafe.Pointer(vp))
	}
}

func stringEncoder(offset int, flags MarshalFlags) UnsafeEncoder {
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		h := (*zgo.String)(unsafe.Add(v, offset))
		if h.Len == 0 {
			if flags&MarshalOmitEmpty != 0 {
				return dst, nil
			}
			return append(dst, '"', '"'), nil
		}
		data := unsafe.Slice(h.Data, h.Len)
		dst = zstr.AppendString(dst, data)
		return dst, nil
	}
}

func sliceEncoder(deep, offset int, t reflect.Type, flags MarshalFlags) UnsafeEncoder {
	elem := t.Elem()
	if elem.Kind() == reflect.Uint8 {
		return sliceBase64Encoder(offset, flags)
	}
	elemSize := elem.Size()
	elemEncoder := getValEncoder(deep, 0, elem, MarshalDefault)
	return func(dst []byte, v unsafe.Pointer) (_ []byte, err error) {
		h := (*zgo.Slice)(unsafe.Add(v, offset))
		if h.Len == 0 {
			if flags&MarshalOmitEmpty != 0 {
				return dst, nil
			}
			return append(dst, '[', ']'), nil
		}
		dst = append(dst, '[')
		for i := range h.Len {
			if i > 0 {
				dst = append(dst, ',')
			}
			vp := unsafe.Pointer(h.Data + elemSize*i)
			dst, err = elemEncoder(dst, vp)
			if err != nil {
				return dst, err
			}
		}
		dst = append(dst, ']')
		return dst, nil
	}
}

func sliceBase64Encoder(offset int, flags MarshalFlags) UnsafeEncoder {
	return func(dst []byte, v unsafe.Pointer) (_ []byte, err error) {
		data := *(*[]byte)(unsafe.Add(v, offset))
		if len(data) == 0 {
			if flags&MarshalOmitEmpty != 0 {
				return dst, nil
			}
			return append(dst, '[', ']'), nil
		}
		return zstr.AppendBase64String(dst, data), nil
	}
}

func arrayEncoder(deep, offset int, t reflect.Type, flags MarshalFlags) UnsafeEncoder {
	arrayLen := uintptr(t.Len())
	elem := t.Elem()
	if elem.Kind() == reflect.Uint8 {
		return arrayByteHexEncoder(offset, arrayLen, flags)
	}
	elemSize := elem.Size()
	elemEncoder := getValEncoder(deep, 0, elem, MarshalDefault)
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		v = unsafe.Add(v, offset)
		var err error
		dst = append(dst, '[')
		for i := range arrayLen {
			if i > 0 {
				dst = append(dst, ',')
			}
			vp := unsafe.Add(v, elemSize*i)
			dst, err = elemEncoder(dst, vp)
			if err != nil {
				return dst, err
			}
		}
		dst = append(dst, ']')
		return dst, nil
	}
}

func arrayByteHexEncoder(offset int, arrayLen uintptr, flags MarshalFlags) UnsafeEncoder {
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		data := zgo.MakeSliceBytes(unsafe.Add(v, offset), arrayLen, arrayLen)
		if flags&MarshalOmitEmpty != 0 {
			var mask byte
			for i := range data {
				mask |= data[i]
			}
			if mask == 0 {
				return dst, nil
			}
		}
		return zstr.AppendHexString(dst, data), nil
	}
}

func uintEncoder(offset int, flags MarshalFlags) UnsafeEncoder {
	if math.MaxInt == math.MaxInt64 {
		return uint64Encoder(offset, flags)
	}
	return uint32Encoder(offset, flags)
}

func intEncoder(offset int, flags MarshalFlags) UnsafeEncoder {
	if math.MaxInt == math.MaxInt64 {
		return int64Encoder(offset, flags)
	}
	return int32Encoder(offset, flags)
}

func uint64Encoder(offset int, flags MarshalFlags) UnsafeEncoder {
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		n := *(*uint64)(unsafe.Add(v, offset))
		if n == 0 && flags&MarshalOmitEmpty != 0 {
			return dst, nil
		}
		if flags&MarshalQuote != 0 {
			dst = append(dst, '"')
			dst = zstr.AppendUint64(dst, n)
			dst = append(dst, '"')
			return dst, nil
		}
		return zstr.AppendUint64(dst, n), nil
	}
}

func int64Encoder(offset int, flags MarshalFlags) UnsafeEncoder {
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		n := *(*int64)(unsafe.Add(v, offset))
		if n == 0 && flags&MarshalOmitEmpty != 0 {
			return dst, nil
		}
		if flags&MarshalQuote != 0 {
			dst = append(dst, '"')
			dst = zstr.AppendInt64(dst, n)
			dst = append(dst, '"')
			return dst, nil
		}
		return zstr.AppendInt64(dst, n), nil
	}
}

func uint32Encoder(offset int, flags MarshalFlags) UnsafeEncoder {
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		n := *(*uint32)(unsafe.Add(v, offset))
		if n == 0 && flags&MarshalOmitEmpty != 0 {
			return dst, nil
		}
		if flags&MarshalQuote != 0 {
			dst = append(dst, '"')
			dst = zstr.AppendUint64(dst, uint64(n))
			dst = append(dst, '"')
			return dst, nil
		}
		return zstr.AppendUint64(dst, uint64(n)), nil
	}
}

func int32Encoder(offset int, flags MarshalFlags) UnsafeEncoder {
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		n := *(*int32)(unsafe.Add(v, offset))
		if n == 0 && flags&MarshalOmitEmpty != 0 {
			return dst, nil
		}
		if flags&MarshalQuote != 0 {
			dst = append(dst, '"')
			dst = zstr.AppendInt64(dst, int64(n))
			dst = append(dst, '"')
			return dst, nil
		}
		return zstr.AppendInt64(dst, int64(n)), nil
	}
}

func uint16Encoder(offset int, flags MarshalFlags) UnsafeEncoder {
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		n := *(*uint16)(unsafe.Add(v, offset))
		if n == 0 && flags&MarshalOmitEmpty != 0 {
			return dst, nil
		}
		if flags&MarshalQuote != 0 {
			dst = append(dst, '"')
			dst = zstr.AppendUint64(dst, uint64(n))
			dst = append(dst, '"')
			return dst, nil
		}
		return zstr.AppendUint64(dst, uint64(n)), nil
	}
}

func int16Encoder(offset int, flags MarshalFlags) UnsafeEncoder {
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		n := *(*int16)(unsafe.Add(v, offset))
		if n == 0 && flags&MarshalOmitEmpty != 0 {
			return dst, nil
		}
		if flags&MarshalQuote != 0 {
			dst = append(dst, '"')
			dst = zstr.AppendInt64(dst, int64(n))
			dst = append(dst, '"')
			return dst, nil
		}
		return zstr.AppendInt64(dst, int64(n)), nil
	}
}

func uint8Encoder(offset int, flags MarshalFlags) UnsafeEncoder {
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		n := *(*uint8)(unsafe.Add(v, offset))
		if n == 0 && flags&MarshalOmitEmpty != 0 {
			return dst, nil
		}
		if flags&MarshalQuote != 0 {
			dst = append(dst, '"')
			dst = zstr.AppendUint8(dst, n)
			dst = append(dst, '"')
			return dst, nil
		}
		return zstr.AppendUint8(dst, n), nil
	}
}

func int8Encoder(offset int, flags MarshalFlags) UnsafeEncoder {
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		n := *(*int8)(unsafe.Add(v, offset))
		if n == 0 && flags&MarshalOmitEmpty != 0 {
			return dst, nil
		}
		if flags&MarshalQuote != 0 {
			dst = append(dst, '"')
			dst = zstr.AppendInt8(dst, n)
			dst = append(dst, '"')
			return dst, nil
		}
		return zstr.AppendInt8(dst, n), nil
	}
}

func float32Encoder(offset int, flags MarshalFlags) UnsafeEncoder {
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		n := *(*float32)(unsafe.Add(v, offset))
		if n == 0 && flags&MarshalOmitEmpty != 0 {
			return dst, nil
		}
		if flags&MarshalQuote != 0 {
			dst = append(dst, '"')
			dst = strconv.AppendFloat(dst, float64(n), 'f', -1, 32)
			dst = append(dst, '"')
			return dst, nil
		}
		return strconv.AppendFloat(dst, float64(n), 'f', -1, 32), nil
	}
}

func float64Encoder(offset int, flags MarshalFlags) UnsafeEncoder {
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		n := *(*float64)(unsafe.Add(v, offset))
		if n == 0 && flags&MarshalOmitEmpty != 0 {
			return dst, nil
		}
		if flags&MarshalQuote != 0 {
			dst = append(dst, '"')
			dst = strconv.AppendFloat(dst, n, 'f', -1, 64)
			dst = append(dst, '"')
			return dst, nil
		}
		return strconv.AppendFloat(dst, n, 'f', -1, 64), nil
	}
}

func boolEncoder(offset int, flags MarshalFlags) UnsafeEncoder {
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		if flags&MarshalQuote != 0 {
			if *(*bool)(unsafe.Add(v, offset)) {
				return append(dst, '"', 't', 'r', 'u', 'e', '"'), nil
			}
			return append(dst, '"', 'f', 'a', 'l', 's', 'e', '"'), nil
		}
		if *(*bool)(unsafe.Add(v, offset)) {
			return append(dst, 't', 'r', 'u', 'e'), nil
		}
		if flags&MarshalOmitEmpty != 0 {
			return dst, nil
		}
		return append(dst, 'f', 'a', 'l', 's', 'e'), nil
	}
}

func appendMarshalerEncoder(offset int, t reflect.Type) UnsafeEncoder {
	getInterface := zgo.NewInterfacerFromRType[AppendMarshaler](t)
	return func(dst []byte, v unsafe.Pointer) (newDst []byte, err error) {
		v = unsafe.Add(v, offset)
		newDst, err = getInterface(v).AppendMarshalJSON(dst)
		if err != nil {
			return dst, nil
		}
		return newDst, nil
	}
}

func marshalerEncoder(offset int, t reflect.Type) UnsafeEncoder {
	getInterface := zgo.NewInterfacerFromRType[Marshaler](t)
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		v = unsafe.Add(v, offset)
		data, err := getInterface(v).MarshalJSON()
		if err != nil {
			return dst, nil
		}
		return append(dst, data...), nil
	}
}

func textMarshalerEncoder(offset int, t reflect.Type) UnsafeEncoder {
	getInterface := zgo.NewInterfacerFromRType[TextMarshaler](t)
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		v = unsafe.Add(v, offset)
		data, err := getInterface(v).MarshalText()
		if err != nil {
			return dst, nil
		}
		return zstr.AppendString(dst, data), nil
	}
}
