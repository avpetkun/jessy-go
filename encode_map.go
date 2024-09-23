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

func mapEncoder(deep uint, t reflect.Type, flags Flags) UnsafeEncoder {
	if flags.Has(SortMapKeys) {
		return mapEncoderSorted(deep, t, flags)
	}
	return mapEncoderUnsorted(deep, t, flags)
}

func mapEncoderUnsorted(deep uint, t reflect.Type, flags Flags) UnsafeEncoder {
	omitEmpty := flags.Has(OmitEmpty)
	flags = flags.excludes(OmitEmpty)

	encodeKey := createItemTypeEncoder(deep, (flags | NeedQuotes), t.Key())
	encodeVal := createItemTypeEncoder(deep, flags, t.Elem())
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

func mapEncoderSorted(deep uint, t reflect.Type, flags Flags) UnsafeEncoder {
	omitEmpty := flags.Has(OmitEmpty)
	flags = flags.excludes(OmitEmpty)

	encodeKey := createItemTypeEncoder(deep, (flags | NeedQuotes), t.Key())
	encodeVal := createItemTypeEncoder(deep, flags, t.Elem())
	getIterator := zgo.NewMapIteratorFromRType(t)

	bufPool := sync.Pool{New: func() any { return new(mapSortBuf) }}

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

		buf := bufPool.Get().(*mapSortBuf)
		buf.Pos = slices.Grow(buf.Pos, count)

		var err error
		for range count {
			keyIndex := len(dst)
			dst, err = encodeKey(dst, it.Key)
			if err != nil {
				it.Release()
				buf.Pos = buf.Pos[:0]
				bufPool.Put(buf)
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
				bufPool.Put(buf)
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

		bufPool.Put(buf)

		return dst, nil
	}
}
