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

func Marshal(value any) (data []byte, err error) {
	enc := getValueEncoder(reflect.TypeOf(value))
	return enc(nil, reflect.ValueOf(value).UnsafePointer())
}

var encoders sync.Map

func getValueEncoder(t reflect.Type) encoder {
	if val, ok := encoders.Load(t); ok {
		return val.(encoder)
	}
	enc := getFieldEncoder(0, 0, t, false, false)
	encoders.Store(t, enc)
	return enc
}

var MaxDeep = 10

func getFieldEncoder(deep, offset int, t reflect.Type, isEmbedded, isOmitempty bool) encoder {
	if deep++; deep == MaxDeep {
		return nopEncoder
	}
	switch t.Kind() {
	case reflect.Pointer:
		return pointerEncoder(deep, offset, t, isEmbedded, isOmitempty)
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

type encoder func(dst []byte, v unsafe.Pointer) ([]byte, error)

func structEncoder(deep, offset int, t reflect.Type, isEmbedded bool) encoder {
	type Field struct {
		Name    string
		Encoder encoder
	}
	fieldEncoders := []Field{}
	for i := range t.NumField() {
		f := t.Field(i)

		name := f.Tag.Get("json")
		action := ""
		if i := strings.IndexByte(name, ','); i != -1 {
			action = name[i+1:]
			name = name[:i]
		}
		if name == "-" {
			continue
		} else if name == "" {
			name = f.Name
		}
		omitempty := action == "omitempty"

		if f.Anonymous {
			fieldEncoders = append(fieldEncoders, Field{
				Encoder: getFieldEncoder(deep, int(f.Offset), f.Type, true, omitempty),
			})
		} else if f.IsExported() {
			fieldEncoders = append(fieldEncoders, Field{
				Name:    name,
				Encoder: getFieldEncoder(deep, int(f.Offset), f.Type, false, omitempty),
			})
		}
	}
	return func(dst []byte, v unsafe.Pointer) (_ []byte, err error) {
		vOffset := unsafe.Add(v, offset)
		if !isEmbedded {
			dst = append(dst, '{')
		}
		var dstLen int
		for i, f := range fieldEncoders {
			if i > 0 {
				dst = append(dst, ',')
			}
			if len(f.Name) != 0 {
				dst = append(dst, '"')
				dst = append(dst, f.Name...)
				dst = append(dst, `":`...)
			}
			dstLen = len(dst)
			dst, err = f.Encoder(dst, vOffset)
			if err != nil {
				return dst, err
			}
			if len(dst) == dstLen {
				if i > 0 {
					dst = dst[:dstLen-len(f.Name)-4]
				} else {
					dst = dst[:dstLen-len(f.Name)-3]
				}
			}
		}
		if !isEmbedded {
			dst = append(dst, '}')
		}
		return dst, nil
	}
}

func nopEncoder(dst []byte, v unsafe.Pointer) ([]byte, error) {
	return dst, nil
}

func pointerEncoder(deep, offset int, t reflect.Type, isEmbedded, isOmitempty bool) encoder {
	elemEncoder := getFieldEncoder(deep, 0, t.Elem(), isEmbedded, isOmitempty)
	if isOmitempty {
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			v = unsafe.Add(v, offset)
			if *(*uintptr)(v) == 0 {
				return dst, nil
			}
			return elemEncoder(dst, v)
		}
	}
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		v = unsafe.Add(v, offset)
		if *(*uintptr)(v) == 0 {
			return append(dst, "null"...), nil
		}
		return elemEncoder(dst, v)
	}
}

func mapEncoder(deep, offset int, t reflect.Type, isOmitempty bool) encoder {
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		return append(dst, "null"...), nil
	}
}

func stringEncoder(offset int, isOmitempty bool) encoder {
	type StringHeader struct {
		Data *byte
		Len  int
	}
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		h := (*StringHeader)(unsafe.Add(v, offset))
		if h.Len == 0 {
			if isOmitempty {
				return dst, nil
			}
			return append(dst, `""`...), nil
		}
		data := unsafe.Slice(h.Data, h.Len)
		dst = append(dst, '"')
		dst = append(dst, data...)
		dst = append(dst, '"')
		return dst, nil
	}
}

func sliceEncoder(deep, offset int, t reflect.Type, isOmitempty bool) encoder {
	type SliceHeader struct {
		Data uintptr
		Len  uintptr
		Cap  int
	}
	elem := t.Elem()
	elemSize := elem.Size()
	elemEncoder := getFieldEncoder(deep, 0, elem, false, false)
	return func(dst []byte, v unsafe.Pointer) (_ []byte, err error) {
		h := (*SliceHeader)(unsafe.Add(v, offset))
		if h.Len == 0 {
			if isOmitempty {
				return dst, nil
			}
			return append(dst, `[]`...), nil
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

func arrayEncoder(deep, offset int, t reflect.Type, isOmitempty bool) encoder {
	uoffset := uintptr(offset)
	arrayLen := uintptr(t.Len())
	elem := t.Elem()
	elemSize := elem.Size()
	elemEncoder := getFieldEncoder(deep, 0, elem, false, false)
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

func uintEncoder(offset int, isOmitempty bool) encoder {
	if math.MaxInt == math.MaxInt64 {
		return uint64Encoder(offset, isOmitempty)
	}
	return uint32Encoder(offset, isOmitempty)
}

func intEncoder(offset int, isOmitempty bool) encoder {
	if math.MaxInt == math.MaxInt64 {
		return int64Encoder(offset, isOmitempty)
	}
	return int32Encoder(offset, isOmitempty)
}

func uint64Encoder(offset int, isOmitempty bool) encoder {
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

func int64Encoder(offset int, isOmitempty bool) encoder {
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

func uint32Encoder(offset int, isOmitempty bool) encoder {
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

func int32Encoder(offset int, isOmitempty bool) encoder {
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

func uint16Encoder(offset int, isOmitempty bool) encoder {
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

func int16Encoder(offset int, isOmitempty bool) encoder {
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

func uint8Encoder(offset int, isOmitempty bool) encoder {
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

func int8Encoder(offset int, isOmitempty bool) encoder {
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

func float32Encoder(offset int, isOmitempty bool) encoder {
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

func float64Encoder(offset int, isOmitempty bool) encoder {
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

func boolEncoder(offset int, isOmitempty bool) encoder {
	if isOmitempty {
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			if *(*bool)(unsafe.Add(v, offset)) {
				return append(dst, "true"...), nil
			}
			return dst, nil
		}
	}
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		if *(*bool)(unsafe.Add(v, offset)) {
			return append(dst, "true"...), nil
		}
		return append(dst, "false"...), nil
	}
}
