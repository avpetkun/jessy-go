package jessy

import (
	"math"
	"reflect"
	"strconv"
	"sync"
	"unsafe"

	"github.com/avpetkun/jessy-go/dec"
)

func Marshal(value any) (data []byte, err error) {
	enc := getValueEncoder(reflect.TypeOf(value))
	return enc(nil, reflect.ValueOf(value).UnsafePointer())
}

var encoders sync.Map

func getValueEncoder(t reflect.Type) encoder {
	if val, ok := encoders.Load(t); ok {
		return val.(encoder)
	}
	enc := getFieldEncoder(0, 0, t)
	encoders.Store(t, enc)
	return enc
}

var MaxDeep = 10

func getFieldEncoder(deep, offset int, t reflect.Type) encoder {
	if deep++; deep == MaxDeep {
		return nullEncoder
	}
	switch t.Kind() {
	case reflect.Pointer:
		return pointerEncoder(deep, offset, t)
	case reflect.Struct:
		return structEncoder(deep, offset, t)
	case reflect.Map:
		return mapEncoder(deep, offset, t)
	case reflect.Array:
		return arrayEncoder(deep, offset, t)
	case reflect.Slice:
		return sliceEncoder(deep, offset, t)
	case reflect.String:
		return stringEncoder(offset)
	case reflect.Bool:
		return boolEncoder(offset)
	case reflect.Int:
		return intEncoder(offset)
	case reflect.Int8:
		return int8Encoder(offset)
	case reflect.Int16:
		return int16Encoder(offset)
	case reflect.Int32:
		return int32Encoder(offset)
	case reflect.Int64:
		return int64Encoder(offset)
	case reflect.Uint:
		return uintEncoder(offset)
	case reflect.Uint8:
		return uint8Encoder(offset)
	case reflect.Uint16:
		return uint16Encoder(offset)
	case reflect.Uint32:
		return uint32Encoder(offset)
	case reflect.Uint64:
		return uint64Encoder(offset)
	case reflect.Float32:
		return float32Encoder(offset)
	case reflect.Float64:
		return float64Encoder(offset)
	default:
		return nullEncoder
	}
}

type encoder func(dst []byte, v unsafe.Pointer) ([]byte, error)

func structEncoder(deep, offset int, t reflect.Type) encoder {
	type Field struct {
		Name    string
		Encoder encoder
	}
	fieldEncoders := []Field{}
	for i := range t.NumField() {
		f := t.Field(i)
		fieldEncoders = append(fieldEncoders, Field{
			Name:    f.Name,
			Encoder: getFieldEncoder(deep, int(f.Offset), f.Type),
		})
	}
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		vOffset := unsafe.Add(v, offset)
		var err error
		dst = append(dst, '{')
		for i, f := range fieldEncoders {
			if i > 0 {
				dst = append(dst, ',')
			}

			dst = append(dst, '"')
			dst = append(dst, f.Name...)
			dst = append(dst, `":`...)

			dst, err = f.Encoder(dst, vOffset)
			if err != nil {
				return dst, err
			}
		}
		dst = append(dst, '}')
		return dst, nil
	}
}

func nullEncoder(dst []byte, v unsafe.Pointer) ([]byte, error) {
	return append(dst, "null"...), nil
}

func pointerEncoder(deep, offset int, t reflect.Type) encoder {
	elemEncoder := getFieldEncoder(deep, 0, t.Elem())
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		v = unsafe.Add(v, offset)
		if *(*uintptr)(v) == 0 {
			return append(dst, "null"...), nil
		}
		return elemEncoder(dst, unsafe.Add(v, offset))
	}
}

func mapEncoder(deep, offset int, t reflect.Type) encoder {
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		return append(dst, "null"...), nil
	}
}

func stringEncoder(offset int) encoder {
	type StringHeader struct {
		Data *byte
		Len  int
	}
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		h := (*StringHeader)(unsafe.Add(v, offset))
		data := unsafe.Slice(h.Data, h.Len)
		dst = append(dst, '"')
		dst = append(dst, data...)
		dst = append(dst, '"')
		return dst, nil
	}
}

func sliceEncoder(deep, offset int, t reflect.Type) encoder {
	type SliceHeader struct {
		Data uintptr
		Len  uintptr
		Cap  int
	}
	elem := t.Elem()
	elemSize := elem.Size()
	elemEncoder := getFieldEncoder(deep, 0, elem)
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		var err error
		dst = append(dst, '[')
		h := (*SliceHeader)(unsafe.Add(v, offset))
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

func arrayEncoder(deep, offset int, t reflect.Type) encoder {
	uoffset := uintptr(offset)
	arrayLen := uintptr(t.Len())
	elem := t.Elem()
	elemSize := elem.Size()
	elemEncoder := getFieldEncoder(deep, 0, elem)
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		voffset := uintptr(v) + uoffset
		var err error
		dst = append(dst, '[')
		for i := range arrayLen {
			if i > 0 {
				dst = append(dst, ',')
			}
			vp := unsafe.Pointer(voffset + elemSize*i)
			dst, err = elemEncoder(dst, vp)
			if err != nil {
				return dst, err
			}
		}
		dst = append(dst, ']')
		return dst, nil
	}
}

func uintEncoder(offset int) encoder {
	if math.MaxInt == math.MaxInt64 {
		return uint64Encoder(offset)
	}
	return uint32Encoder(offset)
}

func intEncoder(offset int) encoder {
	if math.MaxInt == math.MaxInt64 {
		return int64Encoder(offset)
	}
	return int32Encoder(offset)
}

func uint64Encoder(offset int) encoder {
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		p := (*uint64)(unsafe.Add(v, offset))
		return dec.AppendUint64(dst, *p), nil
	}
}

func int64Encoder(offset int) encoder {
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		p := (*int64)(unsafe.Add(v, offset))
		return dec.AppendInt64(dst, *p), nil
	}
}

func uint32Encoder(offset int) encoder {
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		p := (*uint32)(unsafe.Add(v, offset))
		return dec.AppendUint64(dst, uint64(*p)), nil
	}
}

func int32Encoder(offset int) encoder {
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		p := (*int32)(unsafe.Add(v, offset))
		return dec.AppendInt64(dst, int64(*p)), nil
	}
}

func uint16Encoder(offset int) encoder {
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		p := (*uint16)(unsafe.Add(v, offset))
		return dec.AppendUint64(dst, uint64(*p)), nil
	}
}

func int16Encoder(offset int) encoder {
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		p := (*int16)(unsafe.Add(v, offset))
		return dec.AppendInt64(dst, int64(*p)), nil
	}
}

func uint8Encoder(offset int) encoder {
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		p := (*uint8)(unsafe.Add(v, offset))
		return dec.AppendUint8(dst, *p), nil
	}
}

func int8Encoder(offset int) encoder {
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		p := (*int8)(unsafe.Add(v, offset))
		return dec.AppendInt8(dst, *p), nil
	}
}

func float32Encoder(offset int) encoder {
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		f := (*float32)(unsafe.Add(v, offset))
		return strconv.AppendFloat(dst, float64(*f), 'f', -1, 32), nil
	}
}

func float64Encoder(offset int) encoder {
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		f := (*float64)(unsafe.Add(v, offset))
		return strconv.AppendFloat(dst, *f, 'f', -1, 64), nil
	}
}

func boolEncoder(offset int) encoder {
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		p := (*bool)(unsafe.Add(v, offset))
		if *p {
			return append(dst, "true"...), nil
		}
		return append(dst, "false"...), nil
	}
}
