package jessy

import (
	"bytes"
	"reflect"
	"slices"
	"sort"
	"sync"
	"unsafe"

	"github.com/avpetkun/jessy-go/zgo"
)

func mapEncoder(deep, indent uint32, t reflect.Type, flags Flags, isDirectIface bool) UnsafeEncoder {
	encodeMap := mapUnpackedEncoder(deep, indent, t, flags)
	if isDirectIface {
		return encodeMap
	}
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		v = *(*unsafe.Pointer)(v)
		if v == nil {
			if flags.Has(NeedQuotes) {
				return append(dst, '"', '"'), nil
			}
			if flags.Has(OmitEmpty) {
				return dst, nil
			}
			return append(dst, 'n', 'u', 'l', 'l'), nil
		}
		return encodeMap(dst, v)
	}
}

func mapUnpackedEncoder(deep, indent uint32, t reflect.Type, flags Flags) UnsafeEncoder {
	if flags.Has(PrettySpaces) {
		if flags.Has(SortMapKeys) {
			return mapEncoderSortedPretty(deep, indent, t, flags)
		}
		return mapEncoderUnsortedPretty(deep, indent, t, flags)
	}
	if flags.Has(SortMapKeys) {
		return mapEncoderSorted(deep, indent, t, flags)
	}
	return mapEncoderUnsorted(deep, indent, t, flags)
}

func mapEncoderUnsorted(deep, indent uint32, t reflect.Type, flags Flags) UnsafeEncoder {
	omitEmpty := flags.Has(OmitEmpty)

	encodeKey := createItemTypeEncoder(deep, indent+1, (flags | NeedQuotes), t.Key())
	encodeVal := createItemTypeEncoder(deep, indent+1, flags, t.Elem())
	getIterator := zgo.NewMapIteratorFromRType(t)

	return func(dst []byte, value unsafe.Pointer) ([]byte, error) {
		it, count := getIterator(value)
		if it == nil {
			if omitEmpty {
				return dst, nil
			}
			return append(dst, 'n', 'u', 'l', 'l'), nil
		}
		if count == 0 {
			it.Release()
			return append(dst, '{', '}'), nil
		}

		dst = append(dst, '{')
		dstInitLen := len(dst)

		var err error
		for range count {
			keyIndex := len(dst)
			dst, err = encodeKey(dst, it.Key)
			if err != nil {
				it.Release()
				return dst, err
			}
			if len(dst) == keyIndex {
				it.Next()
				continue
			}
			dst = append(dst, ':')
			valIndex := len(dst)
			dst, err = encodeVal(dst, it.Elem)
			if err != nil {
				it.Release()
				return dst, err
			}
			if len(dst) == valIndex {
				dst = dst[:keyIndex]
				it.Next()
				continue
			}
			dst = append(dst, ',')

			it.Next()
		}
		it.Release()

		if count = len(dst); count != dstInitLen {
			dst[count-1] = '}'
		} else {
			dst = append(dst, '}')
		}
		return dst, nil
	}
}

var _ sort.Interface = (*mapSortBuf)(nil)

type mapSortBuf struct {
	Pos [][]byte
	Buf bytes.Buffer
}

func (p *mapSortBuf) Len() int           { return len(p.Pos) }
func (p *mapSortBuf) Less(i, j int) bool { return bytes.Compare(p.Pos[i], p.Pos[j]) == -1 }
func (p *mapSortBuf) Swap(i, j int)      { p.Pos[i], p.Pos[j] = p.Pos[j], p.Pos[i] }

var mapSortBufPool = sync.Pool{New: func() any { return new(mapSortBuf) }}

func mapEncoderSorted(deep, indent uint32, t reflect.Type, flags Flags) UnsafeEncoder {
	omitEmpty := flags.Has(OmitEmpty)

	encodeKey := createItemTypeEncoder(deep, indent+1, (flags | NeedQuotes), t.Key())
	encodeVal := createItemTypeEncoder(deep, indent+1, flags, t.Elem())
	getIterator := zgo.NewMapIteratorFromRType(t)

	return func(dst []byte, value unsafe.Pointer) ([]byte, error) {
		it, count := getIterator(value)
		if it == nil {
			if omitEmpty {
				return dst, nil
			}
			return append(dst, 'n', 'u', 'l', 'l'), nil
		}
		if count == 0 {
			it.Release()
			return append(dst, '{', '}'), nil
		}

		dst = append(dst, '{')
		dstInitLen := len(dst)

		buf := mapSortBufPool.Get().(*mapSortBuf)
		buf.Pos = slices.Grow(buf.Pos, count)

		var err error
		for range count {
			keyIndex := len(dst)
			dst, err = encodeKey(dst, it.Key)
			if err != nil {
				it.Release()
				buf.Pos = buf.Pos[:0]
				mapSortBufPool.Put(buf)
				return dst, err
			}
			if len(dst) == keyIndex {
				dst = dst[:keyIndex]
				it.Next()
				continue
			}
			dst = append(dst, ':')
			valIndex := len(dst)
			dst, err = encodeVal(dst, it.Elem)
			if err != nil {
				it.Release()
				buf.Pos = buf.Pos[:0]
				mapSortBufPool.Put(buf)
				return dst, err
			}
			if len(dst) == valIndex {
				dst = dst[:keyIndex]
				it.Next()
				continue
			}
			dst = append(dst, ',')

			buf.Pos = append(buf.Pos, dst[keyIndex:])
			it.Next()
		}
		it.Release()

		dstNewLen := len(dst)
		mapSize := dstNewLen - dstInitLen
		if mapSize == 0 {
			dst = append(dst, '}')
		} else {
			sort.Sort(buf)

			buf.Buf.Grow(mapSize)
			for i := range buf.Pos {
				buf.Buf.Write(buf.Pos[i])
				buf.Pos[i] = nil
			}
			buf.Pos = buf.Pos[:0]

			copy(dst[dstInitLen:], buf.Buf.Bytes())
			buf.Buf.Reset()

			dst[dstNewLen-1] = '}'
		}

		mapSortBufPool.Put(buf)

		return dst, nil
	}
}

//
//
//

func mapEncoderUnsortedPretty(deep, indent uint32, t reflect.Type, flags Flags) UnsafeEncoder {
	omitEmpty := flags.Has(OmitEmpty)

	encodeKey := createItemTypeEncoder(deep, indent+1, (flags | NeedQuotes), t.Key())
	encodeVal := createItemTypeEncoder(deep, indent+1, flags, t.Elem())
	getIterator := zgo.NewMapIteratorFromRType(t)

	deepSpaces0 := getIndent(indent)
	deepSpaces1 := getIndent(indent + 1)

	return func(dst []byte, value unsafe.Pointer) ([]byte, error) {
		it, count := getIterator(value)
		if it == nil {
			if omitEmpty {
				return dst, nil
			}
			return append(dst, 'n', 'u', 'l', 'l'), nil
		}
		if count == 0 {
			it.Release()
			return append(dst, '{', '}'), nil
		}

		dst = append(dst, '{', '\n')
		dstInitLen := len(dst)

		var err error
		for range count {
			keyIndex := len(dst)
			dst = append(dst, deepSpaces1...)
			dst, err = encodeKey(dst, it.Key)
			if err != nil {
				it.Release()
				return dst, err
			}
			if len(dst) == keyIndex {
				it.Next()
				continue
			}
			dst = append(dst, ':', ' ')
			valIndex := len(dst)
			dst, err = encodeVal(dst, it.Elem)
			if err != nil {
				it.Release()
				return dst, err
			}
			if len(dst) == valIndex {
				dst = dst[:keyIndex]
				it.Next()
				continue
			}
			dst = append(dst, ',', '\n')

			it.Next()
		}
		it.Release()

		if count = len(dst); count != dstInitLen {
			dst = dst[:count-2]
			dst = append(dst, '\n')
			dst = append(dst, deepSpaces0...)
			dst = append(dst, '}')
		} else {
			dst[count-1] = '}'
		}
		return dst, nil
	}
}

func mapEncoderSortedPretty(deep, indent uint32, t reflect.Type, flags Flags) UnsafeEncoder {
	omitEmpty := flags.Has(OmitEmpty)

	encodeKey := createItemTypeEncoder(deep, indent+1, (flags | NeedQuotes), t.Key())
	encodeVal := createItemTypeEncoder(deep, indent+1, flags, t.Elem())
	getIterator := zgo.NewMapIteratorFromRType(t)

	deepSpaces0 := getIndent(indent)
	deepSpaces1 := getIndent(indent + 1)

	return func(dst []byte, value unsafe.Pointer) ([]byte, error) {
		it, count := getIterator(value)
		if it == nil {
			if omitEmpty {
				return dst, nil
			}
			return append(dst, 'n', 'u', 'l', 'l'), nil
		}
		if count == 0 {
			it.Release()
			return append(dst, '{', '}'), nil
		}

		dst = append(dst, '{', '\n')
		dstInitLen := len(dst)

		buf := mapSortBufPool.Get().(*mapSortBuf)
		buf.Pos = slices.Grow(buf.Pos, count)

		var err error
		for range count {
			keyIndex := len(dst)
			dst = append(dst, deepSpaces1...)
			dst, err = encodeKey(dst, it.Key)
			if err != nil {
				it.Release()
				buf.Pos = buf.Pos[:0]
				mapSortBufPool.Put(buf)
				return dst, err
			}
			if len(dst) == keyIndex {
				dst = dst[:keyIndex]
				it.Next()
				continue
			}
			dst = append(dst, ':', ' ')
			valIndex := len(dst)
			dst, err = encodeVal(dst, it.Elem)
			if err != nil {
				it.Release()
				buf.Pos = buf.Pos[:0]
				mapSortBufPool.Put(buf)
				return dst, err
			}
			if len(dst) == valIndex {
				dst = dst[:keyIndex]
				it.Next()
				continue
			}
			dst = append(dst, ',', '\n')

			buf.Pos = append(buf.Pos, dst[keyIndex:])
			it.Next()
		}
		it.Release()

		dstNewLen := len(dst)
		mapSize := dstNewLen - dstInitLen
		if mapSize == 0 {
			dst = append(dst, '}')
		} else {
			sort.Sort(buf)

			buf.Buf.Grow(mapSize)
			for i := range buf.Pos {
				buf.Buf.Write(buf.Pos[i])
				buf.Pos[i] = nil
			}
			buf.Pos = buf.Pos[:0]

			copy(dst[dstInitLen:], buf.Buf.Bytes())
			buf.Buf.Reset()

			dst = dst[:dstNewLen-2]
			dst = append(dst, '\n')
			dst = append(dst, deepSpaces0...)
			dst = append(dst, '}')
		}

		mapSortBufPool.Put(buf)

		return dst, nil
	}
}
