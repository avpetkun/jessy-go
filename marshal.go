package jessy

import (
	"math"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"unsafe"

	"github.com/avpetkun/jessy-go/dec"
)

var (
	MarshalMaxDeep = 10

	customEncoders []customEncoder
	encodersCache  sync.Map
)

type UnsafeEncoder func(dst []byte, valuePtr unsafe.Pointer) ([]byte, error)
type ValueEncoder[T any] func(dst []byte, value T) ([]byte, error)

type customEncoder struct {
	reflect.Type
	UnsafeEncoder
}

func AddUnsafeEncoder[T any](encoder UnsafeEncoder) {
	customEncoders = append(customEncoders, customEncoder{
		Type:          reflect.TypeFor[T](),
		UnsafeEncoder: encoder,
	})
}

func AddValueEncoder[T any](encoder ValueEncoder[T]) {
	AddUnsafeEncoder[T](func(dst []byte, valuePtr unsafe.Pointer) ([]byte, error) {
		return encoder(dst, *(*T)(valuePtr))
	})
}

func Marshal(value any) (data []byte, err error) {
	return AppendMarshal(nil, value)
}

func AppendMarshal(dst []byte, value any) (data []byte, err error) {
	eface := *(*goEmptyInterface)(unsafe.Pointer(&value))
	var enc UnsafeEncoder
	if val, ok := encodersCache.Load(eface.Type); ok {
		enc = val.(UnsafeEncoder)
	} else {
		t := eface.Type.Native()
		if t.Kind() == reflect.Pointer {
			t = t.Elem()
		}
		enc = getFieldEncoder(0, 0, t, false, false)
		encodersCache.Store(eface.Type, enc)
	}
	return enc(dst, eface.Value)
}

func getFieldEncoder(deep, offset int, t reflect.Type, isEmbedded, isOmitempty bool) UnsafeEncoder {
	if deep++; deep == MarshalMaxDeep {
		return nopEncoder
	}
	if t.Kind() == reflect.Pointer {
		return pointerEncoder(deep, offset, t, isEmbedded, isOmitempty)
	}
	for i := range customEncoders {
		if customEncoders[i].Type == t {
			enc := customEncoders[i].UnsafeEncoder
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
		return structEncoder(deep, offset, t, isEmbedded)
	case reflect.Map:
		return mapEncoder(deep, offset, t, isOmitempty)
	case reflect.Array:
		return arrayEncoder(deep, offset, t, isOmitempty)
	case reflect.Slice:
		return sliceEncoder(deep, offset, t, isOmitempty)
	case reflect.String:
		return stringEncoder(offset, isOmitempty)
	case reflect.Bool:
		return boolEncoder(offset, isOmitempty)
	case reflect.Int:
		return intEncoder(offset, isOmitempty)
	case reflect.Int8:
		return int8Encoder(offset, isOmitempty)
	case reflect.Int16:
		return int16Encoder(offset, isOmitempty)
	case reflect.Int32:
		return int32Encoder(offset, isOmitempty)
	case reflect.Int64:
		return int64Encoder(offset, isOmitempty)
	case reflect.Uint:
		return uintEncoder(offset, isOmitempty)
	case reflect.Uint8:
		return uint8Encoder(offset, isOmitempty)
	case reflect.Uint16:
		return uint16Encoder(offset, isOmitempty)
	case reflect.Uint32:
		return uint32Encoder(offset, isOmitempty)
	case reflect.Uint64:
		return uint64Encoder(offset, isOmitempty)
	case reflect.Float32:
		return float32Encoder(offset, isOmitempty)
	case reflect.Float64:
		return float64Encoder(offset, isOmitempty)
	default:
		return nopEncoder
	}
}

func nopEncoder(dst []byte, v unsafe.Pointer) ([]byte, error) {
	return dst, nil
}

func structEncoder(deep, offset int, t reflect.Type, isEmbedded bool) UnsafeEncoder {
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
		omitempty := action == "omitempty"

		if f.Anonymous {
			fields = append(fields, Field{
				Key:     ",",
				KeyLen:  1,
				Encoder: getFieldEncoder(deep, int(f.Offset), f.Type, true, omitempty),
			})
		} else if f.IsExported() {
			key := `"` + name + `":`
			if i > 0 {
				key = "," + key
			}
			fields = append(fields, Field{
				Key:     key,
				KeyLen:  len(key),
				Encoder: getFieldEncoder(deep, int(f.Offset), f.Type, false, omitempty),
			})
		}
	}
	if isEmbedded {
		return func(dst []byte, v unsafe.Pointer) (_ []byte, err error) {
			v = unsafe.Add(v, offset)
			for i := range fields {
				dst = append(dst, fields[i].Key...)
				dstLen := len(dst)
				dst, err = fields[i].Encoder(dst, v)
				if err != nil {
					return dst, err
				}
				if len(dst) == dstLen {
					dst = dst[:dstLen-fields[i].KeyLen]
				}
			}
			return dst, nil
		}
	}
	return func(dst []byte, v unsafe.Pointer) (_ []byte, err error) {
		v = unsafe.Add(v, offset)
		dst = append(dst, '{')
		for i := range fields {
			dst = append(dst, fields[i].Key...)
			dstLen := len(dst)
			dst, err = fields[i].Encoder(dst, v)
			if err != nil {
				return dst, err
			}
			if len(dst) == dstLen {
				dst = dst[:dstLen-fields[i].KeyLen]
			}
		}
		dst = append(dst, '}')
		return dst, nil
	}
}

func mapEncoder(deep, offset int, t reflect.Type, isOmitempty bool) UnsafeEncoder {
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		return append(dst, 'n', 'u', 'l', 'l'), nil
	}
}

func pointerEncoder(deep, offset int, t reflect.Type, isEmbedded, isOmitempty bool) UnsafeEncoder {
	elemEncoder := getFieldEncoder(deep, 0, t.Elem(), isEmbedded, isOmitempty)
	if isOmitempty {
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			v = unsafe.Add(v, offset)
			vp := *(*uintptr)(v)
			if vp == 0 {
				return dst, nil
			}
			return elemEncoder(dst, unsafe.Pointer(vp))
		}
	}
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		v = unsafe.Add(v, offset)
		vp := *(*uintptr)(v)
		if vp == 0 {
			return append(dst, 'n', 'u', 'l', 'l'), nil
		}
		return elemEncoder(dst, unsafe.Pointer(vp))
	}
}

func stringEncoder(offset int, isOmitempty bool) UnsafeEncoder {
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		h := (*goStringHeader)(unsafe.Add(v, offset))
		if h.Len == 0 {
			if isOmitempty {
				return dst, nil
			}
			return append(dst, '"', '"'), nil
		}
		data := unsafe.Slice(h.Data, h.Len)
		dst = appendString(dst, data)
		return dst, nil
	}
}

func sliceEncoder(deep, offset int, t reflect.Type, isOmitempty bool) UnsafeEncoder {
	elem := t.Elem()
	elemSize := elem.Size()
	elemEncoder := getFieldEncoder(deep, 0, elem, false, false)
	return func(dst []byte, v unsafe.Pointer) (_ []byte, err error) {
		h := (*goSliceHeader)(unsafe.Add(v, offset))
		if h.Len == 0 {
			if isOmitempty {
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

func arrayEncoder(deep, offset int, t reflect.Type, isOmitempty bool) UnsafeEncoder {
	arrayLen := uintptr(t.Len())
	elem := t.Elem()
	if elem.Kind() == reflect.Uint8 {
		return arrayByteHexEncoder(offset, arrayLen, isOmitempty)
	}
	elemSize := elem.Size()
	elemEncoder := getFieldEncoder(deep, 0, elem, false, false)
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

func arrayByteHexEncoder(offset int, arrayLen uintptr, isOmitempty bool) UnsafeEncoder {
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		data := *(*[]byte)(unsafe.Pointer(&goSliceHeader{
			Data: uintptr(unsafe.Add(v, offset)),
			Len:  arrayLen,
			Cap:  arrayLen,
		}))
		if isOmitempty {
			var mask byte
			for i := range data {
				mask |= data[i]
			}
			if mask == 0 {
				return dst, nil
			}
		}
		return appendHexString(dst, data), nil
	}
}

func uintEncoder(offset int, isOmitempty bool) UnsafeEncoder {
	if math.MaxInt == math.MaxInt64 {
		return uint64Encoder(offset, isOmitempty)
	}
	return uint32Encoder(offset, isOmitempty)
}

func intEncoder(offset int, isOmitempty bool) UnsafeEncoder {
	if math.MaxInt == math.MaxInt64 {
		return int64Encoder(offset, isOmitempty)
	}
	return int32Encoder(offset, isOmitempty)
}

func uint64Encoder(offset int, isOmitempty bool) UnsafeEncoder {
	if isOmitempty {
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			n := *(*uint64)(unsafe.Add(v, offset))
			if n == 0 {
				return dst, nil
			}
			return dec.AppendUint64(dst, n), nil
		}
	}
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		n := *(*uint64)(unsafe.Add(v, offset))
		return dec.AppendUint64(dst, n), nil
	}
}

func int64Encoder(offset int, isOmitempty bool) UnsafeEncoder {
	if isOmitempty {
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			n := *(*int64)(unsafe.Add(v, offset))
			if n == 0 {
				return dst, nil
			}
			return dec.AppendInt64(dst, n), nil
		}
	}
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		n := *(*int64)(unsafe.Add(v, offset))
		return dec.AppendInt64(dst, n), nil
	}
}

func uint32Encoder(offset int, isOmitempty bool) UnsafeEncoder {
	if isOmitempty {
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			n := *(*uint32)(unsafe.Add(v, offset))
			if n == 0 {
				return dst, nil
			}
			return dec.AppendUint64(dst, uint64(n)), nil
		}
	}
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		n := *(*uint32)(unsafe.Add(v, offset))
		return dec.AppendUint64(dst, uint64(n)), nil
	}
}

func int32Encoder(offset int, isOmitempty bool) UnsafeEncoder {
	if isOmitempty {
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			n := *(*int32)(unsafe.Add(v, offset))
			if n == 0 {
				return dst, nil
			}
			return dec.AppendInt64(dst, int64(n)), nil
		}
	}
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		n := *(*int32)(unsafe.Add(v, offset))
		return dec.AppendInt64(dst, int64(n)), nil
	}
}

func uint16Encoder(offset int, isOmitempty bool) UnsafeEncoder {
	if isOmitempty {
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			n := *(*uint16)(unsafe.Add(v, offset))
			if n == 0 {
				return dst, nil
			}
			return dec.AppendUint64(dst, uint64(n)), nil
		}
	}
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		n := *(*uint16)(unsafe.Add(v, offset))
		return dec.AppendUint64(dst, uint64(n)), nil
	}
}

func int16Encoder(offset int, isOmitempty bool) UnsafeEncoder {
	if isOmitempty {
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			n := *(*int16)(unsafe.Add(v, offset))
			if n == 0 {
				return dst, nil
			}
			return dec.AppendInt64(dst, int64(n)), nil
		}
	}
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		n := *(*int16)(unsafe.Add(v, offset))
		return dec.AppendInt64(dst, int64(n)), nil
	}
}

func uint8Encoder(offset int, isOmitempty bool) UnsafeEncoder {
	if isOmitempty {
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			n := *(*uint8)(unsafe.Add(v, offset))
			if n == 0 {
				return dst, nil
			}
			return dec.AppendUint8(dst, n), nil
		}
	}
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		n := *(*uint8)(unsafe.Add(v, offset))
		return dec.AppendUint8(dst, n), nil
	}
}

func int8Encoder(offset int, isOmitempty bool) UnsafeEncoder {
	if isOmitempty {
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			n := *(*int8)(unsafe.Add(v, offset))
			if n == 0 {
				return dst, nil
			}
			return dec.AppendInt8(dst, n), nil
		}
	}
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		n := *(*int8)(unsafe.Add(v, offset))
		return dec.AppendInt8(dst, n), nil
	}
}

func float32Encoder(offset int, isOmitempty bool) UnsafeEncoder {
	if isOmitempty {
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			n := *(*float32)(unsafe.Add(v, offset))
			if n == 0 {
				return dst, nil
			}
			return strconv.AppendFloat(dst, float64(n), 'f', -1, 32), nil
		}
	}
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		n := *(*float32)(unsafe.Add(v, offset))
		return strconv.AppendFloat(dst, float64(n), 'f', -1, 32), nil
	}
}

func float64Encoder(offset int, isOmitempty bool) UnsafeEncoder {
	if isOmitempty {
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			n := *(*float64)(unsafe.Add(v, offset))
			if n == 0 {
				return dst, nil
			}
			return strconv.AppendFloat(dst, n, 'f', -1, 64), nil
		}
	}
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		n := *(*float64)(unsafe.Add(v, offset))
		return strconv.AppendFloat(dst, n, 'f', -1, 64), nil
	}
}

func boolEncoder(offset int, isOmitempty bool) UnsafeEncoder {
	if isOmitempty {
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			if *(*bool)(unsafe.Add(v, offset)) {
				return append(dst, 't', 'r', 'u', 'e'), nil
			}
			return dst, nil
		}
	}
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		if *(*bool)(unsafe.Add(v, offset)) {
			return append(dst, 't', 'r', 'u', 'e'), nil
		}
		return append(dst, 'f', 'a', 'l', 's', 'e'), nil
	}
}

func appendMarshalerEncoder(offset int, t reflect.Type) UnsafeEncoder {
	newValue := newRValuerForRType(t)
	return func(dst []byte, v unsafe.Pointer) (newDst []byte, err error) {
		v = unsafe.Add(v, offset)
		val := newValue(v).Interface()
		newDst, err = val.(AppendMarshaler).AppendMarshalJSON(dst)
		if err != nil {
			return dst, nil
		}
		return newDst, nil
	}
}

func marshalerEncoder(offset int, t reflect.Type) UnsafeEncoder {
	newValue := newRValuerForRType(t)
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		v = unsafe.Add(v, offset)
		val := newValue(v).Interface()
		data, err := val.(Marshaler).MarshalJSON()
		if err != nil {
			return dst, nil
		}
		return append(dst, data...), nil
	}
}

func textMarshalerEncoder(offset int, t reflect.Type) UnsafeEncoder {
	newValue := newRValuerForRType(t)
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		v = unsafe.Add(v, offset)
		val := newValue(v).Interface()
		data, err := val.(TextMarshaler).MarshalText()
		if err != nil {
			return dst, nil
		}
		return appendString(dst, data), nil
	}
}
