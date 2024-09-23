package jessy

import (
	"reflect"
	"unsafe"

	"github.com/avpetkun/jessy-go/zgo"
	"github.com/avpetkun/jessy-go/zstr"
)

func sliceEncoder(deep uint, t reflect.Type, flags Flags) UnsafeEncoder {
	elem := t.Elem()
	if elem.Kind() == reflect.Uint8 {
		return sliceBase64Encoder(flags)
	}

	omitEmpty := flags.Has(OmitEmpty)
	flags = flags.excludes(OmitEmpty)

	elemSize := uint(elem.Size())
	elemEncoder := createItemTypeEncoder(deep, flags, elem)

	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		h := (*zgo.Slice)(v)
		if h == nil || h.Len == 0 {
			if omitEmpty {
				return dst, nil
			}
			return append(dst, '[', ']'), nil
		}
		dst = append(dst, '[')
		var err error
		dstLen := len(dst)
		newLen := 0
		for i := range h.Len {
			dst, err = elemEncoder(dst, unsafe.Add(h.Data, elemSize*i))
			if err != nil {
				return dst, err
			}
			if newLen = len(dst); newLen != dstLen {
				dst = append(dst, ',')
				dstLen = newLen
			}
		}

		if dstLen = len(dst) - 1; dst[dstLen] == ',' {
			dst[dstLen] = ']'
		} else {
			dst = append(dst, ']')
		}
		return dst, nil
	}
}

func sliceBase64Encoder(flags Flags) UnsafeEncoder {
	omitEmpty := flags.Has(OmitEmpty)
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		data := *(*[]byte)(v)
		if len(data) == 0 {
			if omitEmpty {
				return dst, nil
			}
			return append(dst, '[', ']'), nil
		}
		return zstr.AppendBase64String(dst, data), nil
	}
}

func arrayEncoder(deep uint, t reflect.Type, flags Flags) UnsafeEncoder {
	arrayLen := uint(t.Len())
	elem := t.Elem()
	if elem.Kind() == reflect.Uint8 {
		return arrayByteHexEncoder(arrayLen, flags)
	}

	elemSize := uint(elem.Size())
	elemEncoder := createItemTypeEncoder(deep, flags.excludes(OmitEmpty), elem)

	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		dst = append(dst, '[')
		var err error
		dstLen := len(dst)
		newLen := 0
		for i := range arrayLen {
			dst, err = elemEncoder(dst, unsafe.Add(v, elemSize*i))
			if err != nil {
				return dst, err
			}
			if newLen = len(dst); newLen != dstLen {
				dst = append(dst, ',')
				dstLen = newLen
			}
		}

		if dstLen = len(dst) - 1; dst[dstLen] == ',' {
			dst[dstLen] = ']'
		} else {
			dst = append(dst, ']')
		}
		return dst, nil
	}
}

func arrayByteHexEncoder(arrayLen uint, flags Flags) UnsafeEncoder {
	omitEmpty := flags.Has(OmitEmpty)

	if omitEmpty {
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			data := zgo.NewSliceBytes(v, arrayLen, arrayLen)
			var mask byte
			for i := range data {
				mask |= data[i]
			}
			if mask == 0 {
				return dst, nil
			}
			return zstr.AppendHexString(dst, data), nil
		}
	}
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		data := zgo.NewSliceBytes(v, arrayLen, arrayLen)
		return zstr.AppendHexString(dst, data), nil
	}
}
