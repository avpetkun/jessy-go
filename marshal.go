package jessy

import (
	"bytes"
	"math"
	"reflect"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"unsafe"

	"github.com/avpetkun/jessy-go/zgo"
	"github.com/avpetkun/jessy-go/zstr"
)

var (
	MarshalMaxDeep uint = 10

	customEncoders []customEncoder
)

type UnsafeEncoder func(dst []byte, value unsafe.Pointer) ([]byte, error)
type ValueEncoder[T any] func(dst []byte, value T) ([]byte, error)

type customEncoder struct {
	reflect.Type
	Encoder func(flags Flags) UnsafeEncoder
}

func AddUnsafeEncoder[T any](encoder func(flags Flags) UnsafeEncoder) {
	customEncoders = append(customEncoders, customEncoder{
		Type:    reflect.TypeFor[T](),
		Encoder: encoder,
	})
}

func AddValueEncoder[T any](encoder func(flags Flags) ValueEncoder[T]) {
	AddUnsafeEncoder[T](func(flags Flags) UnsafeEncoder {
		valEnc := encoder(flags)
		return func(dst []byte, value unsafe.Pointer) ([]byte, error) {
			return valEnc(dst, *(*T)(value))
		}
	})
}

func Marshal(value any) (data []byte, err error) {
	return AppendMarshalFlags(nil, value, EncodeStandard)
}

func AppendMarshal(dst []byte, value any) (data []byte, err error) {
	return AppendMarshalFlags(dst, value, EncodeStandard)
}

func MarshalFast(value any) (data []byte, err error) {
	return AppendMarshalFlags(nil, value, EncodeFastest)
}

func AppendMarshalFast(dst []byte, value any) (data []byte, err error) {
	return AppendMarshalFlags(dst, value, EncodeFastest)
}

func AppendMarshalFlags(dst []byte, value any, flags Flags) (data []byte, err error) {
	eface := zgo.UnpackEface(value)
	if eface.Type == nil {
		return append(dst, 'n', 'u', 'l', 'l'), nil
	}
	enc := getValueTypeEncoder(eface.Type, flags)
	return enc(dst, eface.Value)
}

var encodersTypesCache [encodeFlagsLen]sync.Map

func getValueTypeEncoder(typ *zgo.Type, flags Flags) UnsafeEncoder {
	if val, ok := encodersTypesCache[flags].Load(typ); ok {
		return val.(UnsafeEncoder)
	}
	t := typ.Native()
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	enc := getValueEncoder(0, 0, t, flags)
	encodersTypesCache[flags].Store(typ, enc)
	return enc
}

func getValueEncoder(deep, offset uint, t reflect.Type, flags Flags) UnsafeEncoder {
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
			return textMarshalerEncoder(offset, t, flags)
		case tp.Implements(typeTextMarshaler):
			return textMarshalerEncoder(offset, tp, flags)
		}
	}
	switch t.Kind() {
	case reflect.Struct:
		return structEncoder(deep, offset, t, flags, false)
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
			if eface.Value == nil {
				return append(dst, 'n', 'u', 'l', 'l'), nil
			}
			return getValueTypeEncoder(eface.Type, flags)(dst, eface.Value)
		}
	default:
		return nopEncoder
	}
}

func getEmbeddedStructEncoder(deep, offset uint, t reflect.Type, flags Flags) UnsafeEncoder {
	switch t.Kind() {
	case reflect.Pointer:
		return embeddedPointerEncoder(deep, offset, t, flags)
	case reflect.Struct:
		return structEncoder(deep, offset, t, flags, true)
	default:
		return nopEncoder
	}
}

func pointerEncoder(deep, offset uint, t reflect.Type, flags Flags) UnsafeEncoder {
	omitEmpty := flags.Has(OmitEmpty)
	flags = flags.excludes(OmitEmpty)

	t = t.Elem()
	tUnderlying := t
	for tUnderlying.Kind() == reflect.Pointer {
		tUnderlying = tUnderlying.Elem()
	}

	switch tUnderlying.Kind() {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		elemEncoder := getValueEncoder(deep, 0, tUnderlying, flags)
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			if v == nil {
				if omitEmpty {
					return dst, nil
				}
				return append(dst, 'n', 'u', 'l', 'l'), nil
			}
			return elemEncoder(dst, v)
		}
	default:
		elemEncoder := getValueEncoder(deep, 0, t, flags)
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			v = *(*unsafe.Pointer)(unsafe.Add(v, offset))
			if v == nil {
				if omitEmpty {
					return dst, nil
				}
				return append(dst, 'n', 'u', 'l', 'l'), nil
			}
			return elemEncoder(dst, v)
		}
	}
}

var encodersKeyCache [encodeFlagsLen]sync.Map

func getKeyTypeEncoder(typ *zgo.Type, flags Flags) UnsafeEncoder {
	if val, ok := encodersKeyCache[flags].Load(typ); ok {
		return val.(UnsafeEncoder)
	}
	t := typ.Native()
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	enc := getKeyEncoder(t, flags)
	encodersKeyCache[flags].Store(typ, enc)
	return enc
}

func getKeyEncoder(t reflect.Type, flags Flags) UnsafeEncoder {
	if t.Kind() == reflect.Pointer {
		return keyPointerEncoder(t, flags)
	}
	for i := range customEncoders {
		if customEncoders[i].Type == t {
			return customEncoders[i].Encoder(flags)
		}
	}
	if t.Implements(typeTextMarshaler) {
		return textMarshalerEncoder(0, t, flags)
	} else if tp := reflect.PointerTo(t); tp.Implements(typeTextMarshaler) {
		return textMarshalerEncoder(0, tp, flags)
	}
	switch t.Kind() {
	case reflect.Array:
		if t.Elem().Kind() == reflect.Uint8 {
			return arrayByteHexEncoder(0, uint(t.Len()), flags)
		}
	case reflect.String:
		return stringEncoder(0, flags)
	case reflect.Bool:
		return boolEncoder(0, flags)
	case reflect.Int:
		return intEncoder(0, flags)
	case reflect.Int8:
		return int8Encoder(0, flags)
	case reflect.Int16:
		return int16Encoder(0, flags)
	case reflect.Int32:
		return int32Encoder(0, flags)
	case reflect.Int64:
		return int64Encoder(0, flags)
	case reflect.Uint:
		return uintEncoder(0, flags)
	case reflect.Uint8:
		return uint8Encoder(0, flags)
	case reflect.Uint16:
		return uint16Encoder(0, flags)
	case reflect.Uint32:
		return uint32Encoder(0, flags)
	case reflect.Uint64:
		return uint64Encoder(0, flags)
	case reflect.Float32:
		return float32Encoder(0, flags)
	case reflect.Float64:
		return float64Encoder(0, flags)
	case reflect.Interface:
		return func(dst []byte, value unsafe.Pointer) ([]byte, error) {
			eface := (*zgo.EmptyInterface)(value)
			return getKeyTypeEncoder(eface.Type, flags)(dst, eface.Value)
		}
	}
	return nopEncoder
}

func nopEncoder(dst []byte, v unsafe.Pointer) ([]byte, error) {
	return dst, nil
}

func structEncoder(deep, offset uint, t reflect.Type, flags Flags, inEmbedded bool) UnsafeEncoder {
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
}

func mapEncoder(deep, offset uint, t reflect.Type, flags Flags) UnsafeEncoder {
	if flags.Has(SortMapKeys) {
		return mapEncoderSorted(deep, offset, t, flags)
	}
	return mapEncoderUnsorted(deep, offset, t, flags)
}

var _ sort.Interface = (*mapSortBuf)(nil)

type mapSortBuf struct {
	Pos [][]byte
	Buf bytes.Buffer
}

func (p *mapSortBuf) Len() int {
	return len(p.Pos)
}
func (p *mapSortBuf) Less(i, j int) bool {
	return bytes.Compare(p.Pos[i], p.Pos[j]) == -1
}
func (p *mapSortBuf) Swap(i, j int) {
	p.Pos[i], p.Pos[j] = p.Pos[j], p.Pos[i]
}

func mapEncoderSorted(deep, offset uint, t reflect.Type, flags Flags) UnsafeEncoder {
	omitEmpty := flags.Has(OmitEmpty)
	flags = flags.excludes(OmitEmpty)

	encodeKey := getKeyEncoder(t.Key(), (flags | NeedQuotes))
	encodeVal := getValueEncoder(deep, 0, t.Elem(), flags)
	getIterator := zgo.NewMapIteratorFromRType(t)

	bufPool := sync.Pool{New: func() any { return new(mapSortBuf) }}

	return func(dst []byte, value unsafe.Pointer) ([]byte, error) {
		it, count := getIterator(unsafe.Add(value, offset))
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

func mapEncoderUnsorted(deep, offset uint, t reflect.Type, flags Flags) UnsafeEncoder {
	omitEmpty := flags.Has(OmitEmpty)
	flags = flags.excludes(OmitEmpty)

	encodeKey := getKeyEncoder(t.Key(), (flags | NeedQuotes))
	encodeVal := getValueEncoder(deep, 0, t.Elem(), flags)
	getIterator := zgo.NewMapIteratorFromRType(t)

	return func(dst []byte, value unsafe.Pointer) ([]byte, error) {
		it, count := getIterator(unsafe.Add(value, offset))
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

func keyPointerEncoder(t reflect.Type, flags Flags) UnsafeEncoder {
	elemEncoder := getKeyEncoder(t.Elem(), flags)
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		v = *(*unsafe.Pointer)(v)
		if v == nil {
			return append(dst, 'n', 'u', 'l', 'l'), nil
		}
		return elemEncoder(dst, v)
	}
}

func embeddedPointerEncoder(deep, offset uint, t reflect.Type, flags Flags) UnsafeEncoder {
	elemEncoder := getEmbeddedStructEncoder(deep, 0, t.Elem(), flags)
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		v = *(*unsafe.Pointer)(unsafe.Add(v, offset))
		if v == nil {
			return dst, nil
		}
		return elemEncoder(dst, v)
	}
}

func stringEncoder(offset uint, flags Flags) UnsafeEncoder {
	omitEmpty := flags.Has(OmitEmpty)
	escapeHTML := flags.Has(EscapeHTML)
	needValidate := flags.Has(ValidateString) || escapeHTML
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		h := (*zgo.String)(unsafe.Add(v, offset))
		if h.Len == 0 {
			if omitEmpty {
				return dst, nil
			}
			return append(dst, '"', '"'), nil
		}
		data := unsafe.Slice(h.Data, h.Len)
		if needValidate {
			return zstr.AppendQuotedString(dst, data, escapeHTML), nil
		}
		dst = append(dst, '"')
		dst = append(dst, data...)
		dst = append(dst, '"')
		return dst, nil
	}
}

func sliceEncoder(deep, offset uint, t reflect.Type, flags Flags) UnsafeEncoder {
	elem := t.Elem()
	if elem.Kind() == reflect.Uint8 {
		return sliceBase64Encoder(offset, flags)
	}

	omitEmpty := flags.Has(OmitEmpty)
	flags = flags.excludes(OmitEmpty)

	elemSize := uint(elem.Size())
	elemEncoder := getValueEncoder(deep, 0, elem, flags)

	return func(dst []byte, v unsafe.Pointer) (_ []byte, err error) {
		h := (*zgo.Slice)(unsafe.Add(v, offset))
		if h.Len == 0 {
			if omitEmpty {
				return dst, nil
			}
			return append(dst, '[', ']'), nil
		}
		dst = append(dst, '[')
		for i := range h.Len {
			if i > 0 {
				dst = append(dst, ',')
			}
			dst, err = elemEncoder(dst, unsafe.Add(h.Data, elemSize*i))
			if err != nil {
				return dst, err
			}
		}
		dst = append(dst, ']')
		return dst, nil
	}
}

func sliceBase64Encoder(offset uint, flags Flags) UnsafeEncoder {
	omitEmpty := flags.Has(OmitEmpty)
	return func(dst []byte, v unsafe.Pointer) (_ []byte, err error) {
		data := *(*[]byte)(unsafe.Add(v, offset))
		if len(data) == 0 {
			if omitEmpty {
				return dst, nil
			}
			return append(dst, '[', ']'), nil
		}
		return zstr.AppendBase64String(dst, data), nil
	}
}

func arrayEncoder(deep, offset uint, t reflect.Type, flags Flags) UnsafeEncoder {
	arrayLen := uint(t.Len())
	elem := t.Elem()
	if elem.Kind() == reflect.Uint8 {
		return arrayByteHexEncoder(offset, arrayLen, flags)
	}

	flags = flags.excludes(OmitEmpty)

	elemSize := uint(elem.Size())
	elemEncoder := getValueEncoder(deep, 0, elem, flags)
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

func arrayByteHexEncoder(offset uint, arrayLen uint, flags Flags) UnsafeEncoder {
	omitEmpty := flags.Has(OmitEmpty)
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		data := zgo.NewSliceBytes(unsafe.Add(v, offset), arrayLen, arrayLen)
		if omitEmpty {
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

func uintEncoder(offset uint, flags Flags) UnsafeEncoder {
	if math.MaxInt == math.MaxInt64 {
		return uint64Encoder(offset, flags)
	}
	return uint32Encoder(offset, flags)
}

func intEncoder(offset uint, flags Flags) UnsafeEncoder {
	if math.MaxInt == math.MaxInt64 {
		return int64Encoder(offset, flags)
	}
	return int32Encoder(offset, flags)
}

func uint64Encoder(offset uint, flags Flags) UnsafeEncoder {
	omitEmpty := flags.Has(OmitEmpty)
	needQuotes := flags.Has(NeedQuotes)
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		n := *(*uint64)(unsafe.Add(v, offset))
		if n == 0 && omitEmpty {
			return dst, nil
		}
		if needQuotes {
			dst = append(dst, '"')
			dst = zstr.AppendUint64(dst, n)
			dst = append(dst, '"')
			return dst, nil
		}
		return zstr.AppendUint64(dst, n), nil
	}
}

func int64Encoder(offset uint, flags Flags) UnsafeEncoder {
	omitEmpty := flags.Has(OmitEmpty)
	needQuotes := flags.Has(NeedQuotes)
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		n := *(*int64)(unsafe.Add(v, offset))
		if n == 0 && omitEmpty {
			return dst, nil
		}
		if needQuotes {
			dst = append(dst, '"')
			dst = zstr.AppendInt64(dst, n)
			dst = append(dst, '"')
			return dst, nil
		}
		return zstr.AppendInt64(dst, n), nil
	}
}

func uint32Encoder(offset uint, flags Flags) UnsafeEncoder {
	omitEmpty := flags.Has(OmitEmpty)
	needQuotes := flags.Has(NeedQuotes)
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		n := *(*uint32)(unsafe.Add(v, offset))
		if n == 0 && omitEmpty {
			return dst, nil
		}
		if needQuotes {
			dst = append(dst, '"')
			dst = zstr.AppendUint64(dst, uint64(n))
			dst = append(dst, '"')
			return dst, nil
		}
		return zstr.AppendUint64(dst, uint64(n)), nil
	}
}

func int32Encoder(offset uint, flags Flags) UnsafeEncoder {
	omitEmpty := flags.Has(OmitEmpty)
	needQuotes := flags.Has(NeedQuotes)
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		n := *(*int32)(unsafe.Add(v, offset))
		if n == 0 && omitEmpty {
			return dst, nil
		}
		if needQuotes {
			dst = append(dst, '"')
			dst = zstr.AppendInt64(dst, int64(n))
			dst = append(dst, '"')
			return dst, nil
		}
		return zstr.AppendInt64(dst, int64(n)), nil
	}
}

func uint16Encoder(offset uint, flags Flags) UnsafeEncoder {
	omitEmpty := flags.Has(OmitEmpty)
	needQuotes := flags.Has(NeedQuotes)
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		n := *(*uint16)(unsafe.Add(v, offset))
		if n == 0 && omitEmpty {
			return dst, nil
		}
		if needQuotes {
			dst = append(dst, '"')
			dst = zstr.AppendUint64(dst, uint64(n))
			dst = append(dst, '"')
			return dst, nil
		}
		return zstr.AppendUint64(dst, uint64(n)), nil
	}
}

func int16Encoder(offset uint, flags Flags) UnsafeEncoder {
	omitEmpty := flags.Has(OmitEmpty)
	needQuotes := flags.Has(NeedQuotes)
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		n := *(*int16)(unsafe.Add(v, offset))
		if n == 0 && omitEmpty {
			return dst, nil
		}
		if needQuotes {
			dst = append(dst, '"')
			dst = zstr.AppendInt64(dst, int64(n))
			dst = append(dst, '"')
			return dst, nil
		}
		return zstr.AppendInt64(dst, int64(n)), nil
	}
}

func uint8Encoder(offset uint, flags Flags) UnsafeEncoder {
	omitEmpty := flags.Has(OmitEmpty)
	needQuotes := flags.Has(NeedQuotes)
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		n := *(*uint8)(unsafe.Add(v, offset))
		if n == 0 && omitEmpty {
			return dst, nil
		}
		if needQuotes {
			dst = append(dst, '"')
			dst = zstr.AppendUint8(dst, n)
			dst = append(dst, '"')
			return dst, nil
		}
		return zstr.AppendUint8(dst, n), nil
	}
}

func int8Encoder(offset uint, flags Flags) UnsafeEncoder {
	omitEmpty := flags.Has(OmitEmpty)
	needQuotes := flags.Has(NeedQuotes)
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		n := *(*int8)(unsafe.Add(v, offset))
		if n == 0 && omitEmpty {
			return dst, nil
		}
		if needQuotes {
			dst = append(dst, '"')
			dst = zstr.AppendInt8(dst, n)
			dst = append(dst, '"')
			return dst, nil
		}
		return zstr.AppendInt8(dst, n), nil
	}
}

func float32Encoder(offset uint, flags Flags) UnsafeEncoder {
	omitEmpty := flags.Has(OmitEmpty)
	needQuotes := flags.Has(NeedQuotes)
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		n := *(*float32)(unsafe.Add(v, offset))
		if n == 0 && omitEmpty {
			return dst, nil
		}
		if needQuotes {
			dst = append(dst, '"')
			dst = strconv.AppendFloat(dst, float64(n), 'f', -1, 32)
			dst = append(dst, '"')
			return dst, nil
		}
		return strconv.AppendFloat(dst, float64(n), 'f', -1, 32), nil
	}
}

func float64Encoder(offset uint, flags Flags) UnsafeEncoder {
	omitEmpty := flags.Has(OmitEmpty)
	needQuotes := flags.Has(NeedQuotes)
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		n := *(*float64)(unsafe.Add(v, offset))
		if n == 0 && omitEmpty {
			return dst, nil
		}
		if needQuotes {
			dst = append(dst, '"')
			dst = strconv.AppendFloat(dst, n, 'f', -1, 64)
			dst = append(dst, '"')
			return dst, nil
		}
		return strconv.AppendFloat(dst, n, 'f', -1, 64), nil
	}
}

func boolEncoder(offset uint, flags Flags) UnsafeEncoder {
	omitEmpty := flags.Has(OmitEmpty)
	needQuotes := flags.Has(NeedQuotes)
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		if *(*bool)(unsafe.Add(v, offset)) {
			if needQuotes {
				return append(dst, '"', 't', 'r', 'u', 'e', '"'), nil
			}
			return append(dst, 't', 'r', 'u', 'e'), nil
		}
		if omitEmpty {
			return dst, nil
		}
		if needQuotes {
			return append(dst, '"', 'f', 'a', 'l', 's', 'e', '"'), nil
		}
		return append(dst, 'f', 'a', 'l', 's', 'e'), nil
	}
}

func appendMarshalerEncoder(offset uint, t reflect.Type) UnsafeEncoder {
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

func marshalerEncoder(offset uint, t reflect.Type) UnsafeEncoder {
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

func textMarshalerEncoder(offset uint, t reflect.Type, flags Flags) UnsafeEncoder {
	escapeHTML := flags.Has(EscapeHTML)
	needValidate := flags.Has(ValidateTextMarshaller) || escapeHTML

	getInterface := zgo.NewInterfacerFromRType[TextMarshaler](t)
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		v = unsafe.Add(v, offset)
		data, err := getInterface(v).MarshalText()
		if err != nil {
			return dst, nil
		}
		if needValidate {
			return zstr.AppendQuotedString(dst, data, escapeHTML), nil
		}
		dst = append(dst, '"')
		dst = append(dst, data...)
		dst = append(dst, '"')
		return dst, nil
	}
}
