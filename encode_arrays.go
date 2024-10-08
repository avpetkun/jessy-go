package jessy

import (
	"reflect"
	"unsafe"

	"github.com/avpetkun/jessy-go/zgo"
	"github.com/avpetkun/jessy-go/zstr"
)

func sliceEncoder(deep, indent uint32, t reflect.Type, flags Flags) UnsafeEncoder {
	elem := t.Elem()
	if elem.Kind() == reflect.Uint8 && !tImplementsAny(elem) {
		return sliceBase64Encoder(flags)
	}

	prettySpaces := flags.Has(PrettySpaces)
	omitEmpty := flags.Has(OmitEmpty)

	elemSize := uint(elem.Size())
	elemEncoder := createItemTypeEncoder(deep, indent+1, flags, elem)

	if prettySpaces {
		deepSpaces0 := getIndent(indent)
		deepSpaces1 := getIndent(indent + 1)
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			h := (*zgo.Slice)(v)
			if h == nil || h.Len == 0 {
				if omitEmpty {
					return dst, nil
				}
				dst = append(dst, '[', ']')
				return dst, nil
			}
			dst = append(dst, '[', '\n')
			var err error
			for i := range h.Len {
				dst = append(dst, deepSpaces1...)
				dstLen := len(dst)
				dst, err = elemEncoder(dst, unsafe.Add(h.Data, elemSize*i))
				if err != nil {
					return dst, err
				}
				if len(dst) == dstLen {
					dst = dst[:dstLen-1-int(indent)]
				} else {
					dst = append(dst, ',', '\n')
				}
			}

			if i := len(dst) - 2; dst[i] == ',' {
				dst = dst[:i]
			}
			dst = append(dst, '\n')
			dst = append(dst, deepSpaces0...)
			dst = append(dst, ']')

			return dst, nil
		}
	}

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

func arrayEncoder(deep, indent uint32, t reflect.Type, flags Flags) UnsafeEncoder {
	arrayLen := uint(t.Len())
	elem := t.Elem()

	elemSize := uint(elem.Size())
	elemEncoder := createItemTypeEncoder(deep, indent, flags, elem)

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
