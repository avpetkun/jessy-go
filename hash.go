package jessy

import (
	"math"
	"reflect"
	"slices"
	"sort"
	"sync"
	"unsafe"

	"github.com/avpetkun/jessy-go/zgo"
)

func Hash(value any) (hashSum uint64, err error) {
	eface := zgo.UnpackEface(value)
	if eface.Data == nil {
		return 0, nil
	}
	h := newHashSum64()
	err = getTypeHashEncoder(eface.Type)(&h, eface.Data)
	hashSum = h.Sum()
	return
}

type hashEncoder func(h *hashSum64, value unsafe.Pointer) error

var hashEncodersCache sync.Map

func getTypeHashEncoder(typ *zgo.Type) hashEncoder {
	if val, ok := hashEncodersCache.Load(typ); ok {
		return val.(hashEncoder)
	}
	encoder := createTypeHashEncoder(0, typ.Native(), false, false)
	hashEncodersCache.Store(typ, encoder)
	return encoder
}

func nopHashEncoder(h *hashSum64, v unsafe.Pointer) error {
	return nil
}

func createItemTypeHashEncoder(deep int, t reflect.Type) hashEncoder {
	if t.Kind() == reflect.Pointer {
		return pointerHashEncoder(deep, t.Elem(), false, false)
	}
	return createTypeHashEncoder(deep, t, false, false)
}

func createTypeHashEncoder(deep int, t reflect.Type, wasStruct, byPointer bool) hashEncoder {
	if deep++; deep >= MarshalMaxDeep {
		return nopHashEncoder
	}

	if t.Kind() == reflect.Pointer {
		if t.Elem().Kind() == reflect.Struct {
			return createTypeHashEncoder(deep, t.Elem(), wasStruct, true)
		}
		return pointerHashEncoder(deep, t.Elem(), wasStruct, true)
	}

	switch t.Kind() {
	case reflect.Struct:
		return structHashEncoder(deep, t, byPointer)
	case reflect.String:
		return stringHashEncoder
	case reflect.Map:
		if wasStruct {
			return pointerHashEncoder(deep, t, false, false)
		}
		return mapHashEncoderSorted(deep, t)
	case reflect.Slice:
		return sliceHashEncoder(deep, t)
	case reflect.Array:
		return arrayHashEncoder(deep, t)

	case reflect.Bool:
		return boolHashEncoder
	case reflect.Int, reflect.Uint:
		return uintHashEncoder()
	case reflect.Int8, reflect.Uint8:
		return uint8HashEncoder
	case reflect.Int16, reflect.Uint16:
		return uint16HashEncoder
	case reflect.Int32, reflect.Uint32, reflect.Float32:
		return uint32HashEncoder
	case reflect.Int64, reflect.Uint64, reflect.Float64:
		return uint64HashEncoder

	case reflect.Interface:
		return func(h *hashSum64, value unsafe.Pointer) error {
			eface := (*zgo.EmptyInterface)(value)
			if eface.Data == nil {
				h.Byte(0)
				return nil
			}
			return getTypeHashEncoder(eface.Type)(h, eface.Data)
		}
	}

	return nopHashEncoder
}

func structHashEncoder(deep int, t reflect.Type, byPointer bool) hashEncoder {
	fieldsCount := t.NumField()
	if fieldsCount == 0 {
		return func(h *hashSum64, value unsafe.Pointer) error {
			h.Byte(0)
			return nil
		}
	}
	type Field struct {
		Key     uint64
		Offset  uintptr
		Encoder hashEncoder
	}

	fields := make([]Field, 0, fieldsCount)
	for i := range fieldsCount {
		f := t.Field(i)
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
		var fieldEncoder hashEncoder
		if makeUnpack {
			fieldEncoder = pointerHashEncoder(deep, ft, wasStruct, byPointer)
		} else {
			fieldEncoder = createTypeHashEncoder(deep, ft, wasStruct, byPointer)
		}

		keySum := newHashSum64()
		keySum.Write([]byte(f.Name))
		keySum.Byte(':')

		fields = append(fields, Field{
			Key:     keySum.Sum(),
			Offset:  f.Offset,
			Encoder: fieldEncoder,
		})
	}

	sort.Slice(fields, func(i, j int) bool {
		return fields[i].Key < fields[j].Key
	})

	return func(h *hashSum64, value unsafe.Pointer) (err error) {
		for i := range fields {
			h.WriteUint64(fields[i].Key)
			err = fields[i].Encoder(h, unsafe.Add(value, fields[i].Offset))
			if err != nil {
				return err
			}
		}
		return
	}
}

func pointerHashEncoder(deep int, t reflect.Type, wasStruct, byPointer bool) hashEncoder {
	elemEncoder := createTypeHashEncoder(deep, t, wasStruct, byPointer)

	return func(h *hashSum64, v unsafe.Pointer) error {
		v = *(*unsafe.Pointer)(v)
		if v == nil {
			h.Byte(0)
			return nil
		}
		return elemEncoder(h, v)
	}
}

func stringHashEncoder(w *hashSum64, v unsafe.Pointer) error {
	h := (*zgo.String)(v)
	if h.Len == 0 {
		w.Byte(0)
	} else {
		w.Write(unsafe.Slice(h.Data, h.Len))
	}
	return nil
}

func boolHashEncoder(h *hashSum64, v unsafe.Pointer) error {
	if *(*bool)(v) {
		h.Byte(1)
	} else {
		h.Byte(0)
	}
	return nil
}

func uintHashEncoder() hashEncoder {
	if math.MaxInt == math.MaxInt64 {
		return uint64HashEncoder
	}
	return uint32HashEncoder
}

func uint8HashEncoder(h *hashSum64, v unsafe.Pointer) error {
	n := *(*uint8)(v)
	h.Byte(n)
	return nil
}

func uint16HashEncoder(h *hashSum64, v unsafe.Pointer) error {
	n := *(*uint16)(v)
	h.WriteUint16(n)
	return nil
}

func uint32HashEncoder(h *hashSum64, v unsafe.Pointer) error {
	n := *(*uint32)(v)
	h.WriteUint32(n)
	return nil
}

func uint64HashEncoder(h *hashSum64, v unsafe.Pointer) error {
	n := *(*uint64)(v)
	h.WriteUint64(n)
	return nil
}

func arrayHashEncoder(deep int, t reflect.Type) hashEncoder {
	arrayLen := uint(t.Len())
	elem := t.Elem()
	if elem.Kind() == reflect.Uint8 {
		return arrayByteHexHashEncoder(arrayLen)
	}

	elemSize := uint(elem.Size())
	elemEncoder := createItemTypeHashEncoder(deep, elem)

	return func(h *hashSum64, v unsafe.Pointer) (err error) {
		for i := range arrayLen {
			err = elemEncoder(h, unsafe.Add(v, elemSize*i))
			if err != nil {
				return err
			}
		}
		return
	}
}

func arrayByteHexHashEncoder(arrayLen uint) hashEncoder {
	return func(h *hashSum64, v unsafe.Pointer) error {
		data := zgo.NewSliceBytes(v, arrayLen, arrayLen)
		h.Byte(0)
		h.Write(data)
		return nil
	}
}

func sliceHashEncoder(deep int, t reflect.Type) hashEncoder {
	elem := t.Elem()
	if elem.Kind() == reflect.Uint8 {
		return sliceBase64HexEncoder
	}

	elemSize := uint(elem.Size())
	elemEncoder := createItemTypeHashEncoder(deep, elem)

	return func(w *hashSum64, v unsafe.Pointer) (err error) {
		h := (*zgo.Slice)(v)
		if h == nil || h.Len == 0 {
			w.Byte(0)
			return
		}
		for i := range h.Len {
			err = elemEncoder(w, unsafe.Add(h.Data, elemSize*i))
			if err != nil {
				return err
			}
		}
		return
	}
}

func sliceBase64HexEncoder(h *hashSum64, v unsafe.Pointer) error {
	data := *(*[]byte)(v)
	h.Byte(0)
	h.Write(data)
	return nil
}

func mapHashEncoderSorted(deep int, t reflect.Type) hashEncoder {
	encodeKey := createItemTypeHashEncoder(deep, t.Key())
	encodeVal := createItemTypeHashEncoder(deep, t.Elem())
	getIterator := zgo.NewMapIteratorFromRType(t)

	type Buf struct{ Pos []uint64 }
	bufPool := sync.Pool{New: func() any { return new(Buf) }}

	return func(h *hashSum64, value unsafe.Pointer) (err error) {
		it, count := getIterator(value)
		if count == 0 {
			if it != nil {
				it.Release()
			}
			h.Byte(0)
			return
		}

		buf := bufPool.Get().(*Buf)
		pos := slices.Grow(buf.Pos, count)[:count]

		prevHash := *h
		for i := range count {
			err = encodeKey(h, it.Key)
			if err != nil {
				it.Release()
				buf.Pos = pos[:0]
				bufPool.Put(buf)
				return
			}
			err = encodeVal(h, it.Elem)
			if err != nil {
				it.Release()
				buf.Pos = pos[:0]
				bufPool.Put(buf)
				return
			}
			pos[i] = h.Sum()
			*h = prevHash
			it.Next()
		}
		it.Release()

		slices.Sort(pos)

		for _, v := range pos {
			h.WriteUint64(v)
		}
		buf.Pos = pos[:0]
		bufPool.Put(buf)

		return nil
	}
}

//
//
//
//
//

const hashOffset64 = 14695981039346656037
const hashPrime64 = 1099511628211

type hashSum64 uint64

func newHashSum64() hashSum64 { return hashOffset64 }

func (s hashSum64) Sum() uint64 { return uint64(s) }

func (s *hashSum64) Reset() { *s = hashOffset64 }

func (s *hashSum64) Byte(c byte) {
	*s = (*s * hashPrime64) ^ hashSum64(c)
}

func (s *hashSum64) Write(data []byte) {
	hash := *s
	for _, c := range data {
		hash *= hashPrime64
		hash ^= hashSum64(c)
	}
	*s = hash
}

func (s *hashSum64) WriteUint16(v uint16) {
	h := *s
	h = (h * hashPrime64) ^ hashSum64(v&0xFF)
	h = (h * hashPrime64) ^ hashSum64((v>>8)&0xFF)
	*s = h
}

func (s *hashSum64) WriteUint32(v uint32) {
	h := *s
	h = (h * hashPrime64) ^ hashSum64(v&0xFF)
	h = (h * hashPrime64) ^ hashSum64((v>>8)&0xFF)
	h = (h * hashPrime64) ^ hashSum64((v>>16)&0xFF)
	h = (h * hashPrime64) ^ hashSum64((v>>24)&0xFF)
	*s = h
}

func (s *hashSum64) WriteUint64(v uint64) {
	h := uint64(*s)
	h = (h * hashPrime64) ^ (v & 0xFF)
	h = (h * hashPrime64) ^ ((v >> 8) & 0xFF)
	h = (h * hashPrime64) ^ ((v >> 16) & 0xFF)
	h = (h * hashPrime64) ^ ((v >> 24) & 0xFF)
	h = (h * hashPrime64) ^ ((v >> 32) & 0xFF)
	h = (h * hashPrime64) ^ ((v >> 40) & 0xFF)
	h = (h * hashPrime64) ^ ((v >> 48) & 0xFF)
	h = (h * hashPrime64) ^ ((v >> 56) & 0xFF)
	*s = hashSum64(h)
}
