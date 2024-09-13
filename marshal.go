package jessy

import (
	"math"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"
	"unsafe"

	"github.com/avpetkun/jessy-go/dec"
)

var MarshalMaxDeep = 10

var encodersCache sync.Map

func Marshal(value any) (data []byte, err error) {
	return AppendMarshal(nil, value)
}

func AppendMarshal(dst []byte, value any) (data []byte, err error) {
	eface := *(*goEmptyInterface)(unsafe.Pointer(&value))
	var enc encoder
	if val, ok := encodersCache.Load(eface.Type); ok {
		enc = val.(encoder)
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

func getFieldEncoder(deep, offset int, t reflect.Type, isEmbedded, isOmitempty bool) encoder {
	if deep++; deep == MarshalMaxDeep {
		return nopEncoder
	}
	if enc := tryMarshalerEncoder(offset, t); enc != nil {
		return enc
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
		Key     string
		KeyLen  int
		Encoder encoder
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

func nopEncoder(dst []byte, v unsafe.Pointer) ([]byte, error) {
	return dst, nil
}

func pointerEncoder(deep, offset int, t reflect.Type, isEmbedded, isOmitempty bool) encoder {
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

func mapEncoder(deep, offset int, t reflect.Type, isOmitempty bool) encoder {
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		return append(dst, 'n', 'u', 'l', 'l'), nil
	}
}

func stringEncoder(offset int, isOmitempty bool) encoder {
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

func sliceEncoder(deep, offset int, t reflect.Type, isOmitempty bool) encoder {
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

func arrayEncoder(deep, offset int, t reflect.Type, isOmitempty bool) encoder {
	arrayLen := uintptr(t.Len())
	elem := t.Elem()
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

func tryMarshalerEncoder(offset int, t reflect.Type) encoder {
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
	default:
		return nil
	}
}

func appendMarshalerEncoder(offset int, t reflect.Type) encoder {
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

func marshalerEncoder(offset int, t reflect.Type) encoder {
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

func textMarshalerEncoder(offset int, t reflect.Type) encoder {
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

// from encoding/json.appendString
func appendString(dst, src []byte) []byte {
	const hex = "0123456789abcdef"

	dst = append(dst, '"')
	start := 0
	srcLen := len(src)
	for i := 0; i < srcLen; i++ {
		if b := src[i]; b < utf8.RuneSelf {
			if safeSet[b] {
				continue
			}
			dst = append(dst, src[start:i]...)
			switch b {
			case '\\', '"':
				dst = append(dst, '\\', b)
			case '\b':
				dst = append(dst, '\\', 'b')
			case '\f':
				dst = append(dst, '\\', 'f')
			case '\n':
				dst = append(dst, '\\', 'n')
			case '\r':
				dst = append(dst, '\\', 'r')
			case '\t':
				dst = append(dst, '\\', 't')
			default:
				// This encodes bytes < 0x20 except for \b, \f, \n, \r and \t.
				// If escapeHTML is set, it also escapes <, >, and &
				// because they can lead to security holes when
				// user-controlled strings are rendered into JSON
				// and served to some browsers.
				dst = append(dst, '\\', 'u', '0', '0', hex[b>>4], hex[b&0xF])
			}
			i++
			start = i
			continue
		}
		// TODO(https://go.dev/issue/56948): Use generic utf8 functionality.
		// For now, cast only a small portion of byte slices to a string
		// so that it can be stack allocated. This slows down []byte slightly
		// due to the extra copy, but keeps string performance roughly the same.
		n := len(src) - i
		if n > utf8.UTFMax {
			n = utf8.UTFMax
		}
		c, size := utf8.DecodeRuneInString(string(src[i : i+n]))
		if c == utf8.RuneError && size == 1 {
			dst = append(dst, src[start:i]...)
			dst = append(dst, `\ufffd`...)
			i += size
			start = i
			continue
		}
		// U+2028 is LINE SEPARATOR.
		// U+2029 is PARAGRAPH SEPARATOR.
		// They are both technically valid characters in JSON strings,
		// but don't work in JSONP, which has to be evaluated as JavaScript,
		// and can lead to security holes there. It is valid JSON to
		// escape them, so we do so unconditionally.
		// See https://en.wikipedia.org/wiki/JSON#Safety.
		if c == '\u2028' || c == '\u2029' {
			dst = append(dst, src[start:i]...)
			dst = append(dst, '\\', 'u', '2', '0', '2', hex[c&0xF])
			i += size
			start = i
			continue
		}
		i += size
	}
	dst = append(dst, src[start:]...)
	dst = append(dst, '"')
	return dst
}

var safeSet = [utf8.RuneSelf]bool{
	' ':      true,
	'!':      true,
	'"':      false,
	'#':      true,
	'$':      true,
	'%':      true,
	'&':      true,
	'\'':     true,
	'(':      true,
	')':      true,
	'*':      true,
	'+':      true,
	',':      true,
	'-':      true,
	'.':      true,
	'/':      true,
	'0':      true,
	'1':      true,
	'2':      true,
	'3':      true,
	'4':      true,
	'5':      true,
	'6':      true,
	'7':      true,
	'8':      true,
	'9':      true,
	':':      true,
	';':      true,
	'<':      true,
	'=':      true,
	'>':      true,
	'?':      true,
	'@':      true,
	'A':      true,
	'B':      true,
	'C':      true,
	'D':      true,
	'E':      true,
	'F':      true,
	'G':      true,
	'H':      true,
	'I':      true,
	'J':      true,
	'K':      true,
	'L':      true,
	'M':      true,
	'N':      true,
	'O':      true,
	'P':      true,
	'Q':      true,
	'R':      true,
	'S':      true,
	'T':      true,
	'U':      true,
	'V':      true,
	'W':      true,
	'X':      true,
	'Y':      true,
	'Z':      true,
	'[':      true,
	'\\':     false,
	']':      true,
	'^':      true,
	'_':      true,
	'`':      true,
	'a':      true,
	'b':      true,
	'c':      true,
	'd':      true,
	'e':      true,
	'f':      true,
	'g':      true,
	'h':      true,
	'i':      true,
	'j':      true,
	'k':      true,
	'l':      true,
	'm':      true,
	'n':      true,
	'o':      true,
	'p':      true,
	'q':      true,
	'r':      true,
	's':      true,
	't':      true,
	'u':      true,
	'v':      true,
	'w':      true,
	'x':      true,
	'y':      true,
	'z':      true,
	'{':      true,
	'|':      true,
	'}':      true,
	'~':      true,
	'\u007f': true,
}
