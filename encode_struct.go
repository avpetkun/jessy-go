package jessy

import (
	"reflect"
	"sort"
	"strings"
	"unsafe"
)

type StructField struct {
	Key     string
	KeyLen  int
	Offset  uintptr
	Encoder UnsafeEncoder
}

func structEncoder(deep, indent uint32, flags Flags, t reflect.Type, ifaceIndir, embedded bool) UnsafeEncoder {
	if deep++; deep >= MarshalMaxDeep {
		return nopEncoder
	}
	prettySpaces := flags.Has(PrettySpaces)

	fieldsCount := t.NumField()
	if fieldsCount == 0 {
		return func(dst []byte, value unsafe.Pointer) ([]byte, error) {
			return append(dst, '{', '}'), nil
		}
	}

	fields := make([]StructField, 0, fieldsCount)
	for i := range fieldsCount {
		f := t.Field(i)

		anonymousStruct := f.Anonymous && tReallyStruct(f.Type) && !tImplementsAny(f.Type)
		if !f.IsExported() && !anonymousStruct {
			continue
		}

		parts := strings.Split(f.Tag.Get("json"), ",")
		name := parts[0]
		if name == "-" {
			continue
		}
		if name == "" {
			name = f.Name
		}

		fieldFlags := flags
		for _, action := range parts[1:] {
			switch action {
			case "omitempty":
				fieldFlags |= OmitEmpty
			case "string":
				fieldFlags |= NeedQuotes
			}
		}

		nextIndent := indent + 1
		if embedded {
			nextIndent--
		}
		fieldEncoder := createTypeEncoder(deep, nextIndent, fieldFlags, f.Type, ifaceIndir, anonymousStruct)

		if anonymousStruct {
			fields = append(fields, StructField{
				KeyLen:  0,
				Offset:  f.Offset,
				Encoder: fieldEncoder,
			})
		} else {
			key := `"` + name + `":`
			if prettySpaces {
				key += " "
			}
			fields = append(fields, StructField{
				Key:     key,
				KeyLen:  len(key),
				Offset:  f.Offset,
				Encoder: fieldEncoder,
			})
		}
	}

	if len(fields) == 0 {
		return nopStructEncoder
	}

	sort.Slice(fields, func(i, j int) bool {
		return fields[i].Key < fields[j].Key
	})

	if prettySpaces {
		return structEncoderPretty(indent, fields, embedded)
	}
	return structEncoderMinimal(fields, embedded)
}

func nopStructEncoder(dst []byte, value unsafe.Pointer) ([]byte, error) {
	if value == nil {
		return append(dst, 'n', 'u', 'l', 'l'), nil
	}
	return append(dst, '{', '}'), nil
}

func structEncoderPretty(indent uint32, fields []StructField, embedded bool) UnsafeEncoder {
	deepSpace0 := getIndent(indent)
	deepSpace1 := getIndent(indent + 1)
	if embedded {
		return func(dst []byte, value unsafe.Pointer) ([]byte, error) {
			var err error
			var was uint32
			for i := range fields {
				if was != 0 {
					dst = append(dst, deepSpace0...)
				}
				dst = append(dst, fields[i].Key...)
				dstLen := len(dst)
				dst, err = fields[i].Encoder(dst, unsafe.Add(value, fields[i].Offset))
				if err != nil {
					return dst, err
				}
				if len(dst) == dstLen {
					dst = dst[:dstLen-fields[i].KeyLen-int(was)]
				} else {
					dst = append(dst, ',', '\n')
					was = indent
				}
			}
			if i := len(dst); i != 0 && dst[i-1] == '\n' {
				dst = dst[:i-2]
			}
			return dst, nil
		}
	}
	return func(dst []byte, value unsafe.Pointer) ([]byte, error) {
		dst = append(dst, '{', '\n')
		var err error
		for i := range fields {
			dst = append(dst, deepSpace1...)
			dst = append(dst, fields[i].Key...)
			dstLen := len(dst)
			dst, err = fields[i].Encoder(dst, unsafe.Add(value, fields[i].Offset))
			if err != nil {
				return dst, err
			}
			if len(dst) == dstLen {
				dst = dst[:dstLen-fields[i].KeyLen-1-int(indent)]
			} else {
				dst = append(dst, ',', '\n')
			}
		}
		if i := len(dst) - 2; dst[i] == ',' {
			dst = dst[:i]
		}
		dst = append(dst, '\n')
		dst = append(dst, deepSpace0...)
		dst = append(dst, '}')
		return dst, nil
	}
}

func structEncoderMinimal(fields []StructField, embedded bool) UnsafeEncoder {
	if embedded {
		return func(dst []byte, value unsafe.Pointer) ([]byte, error) {
			var err error
			for i := range fields {
				dst = append(dst, fields[i].Key...)
				dstLen := len(dst)
				dst, err = fields[i].Encoder(dst, unsafe.Add(value, fields[i].Offset))
				if err != nil {
					return dst, err
				}
				if len(dst) == dstLen {
					dst = dst[:dstLen-fields[i].KeyLen]
				} else {
					dst = append(dst, ',')
				}
			}
			if i := len(dst); i != 0 && dst[i-1] == ',' {
				dst = dst[:i-1]
			}
			return dst, nil
		}
	}
	return func(dst []byte, value unsafe.Pointer) ([]byte, error) {
		dst = append(dst, '{')
		var err error
		for i := range fields {
			dst = append(dst, fields[i].Key...)
			dstLen := len(dst)
			dst, err = fields[i].Encoder(dst, unsafe.Add(value, fields[i].Offset))
			if err != nil {
				return dst, err
			}
			if len(dst) == dstLen {
				dst = dst[:dstLen-fields[i].KeyLen]
			} else {
				dst = append(dst, ',')
			}
		}
		if i := len(dst) - 1; dst[i] == ',' {
			dst[i] = '}'
		} else {
			dst = append(dst, '}')
		}
		return dst, nil
	}
}
