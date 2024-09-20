package jessy

import (
	"math"
	"strconv"
	"unsafe"

	"github.com/avpetkun/jessy-go/zstr"
)

func uintEncoder(flags Flags) UnsafeEncoder {
	if math.MaxInt == math.MaxInt64 {
		return uint64Encoder(flags)
	}
	return uint32Encoder(flags)
}

func intEncoder(flags Flags) UnsafeEncoder {
	if math.MaxInt == math.MaxInt64 {
		return int64Encoder(flags)
	}
	return int32Encoder(flags)
}

func uint64Encoder(flags Flags) UnsafeEncoder {
	omitEmpty := flags.Has(OmitEmpty)
	needQuotes := flags.Has(NeedQuotes)

	if needQuotes {
		if omitEmpty {
			return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
				n := *(*uint64)(v)
				if n == 0 {
					return dst, nil
				}
				dst = append(dst, '"')
				dst = zstr.AppendUint64(dst, n)
				dst = append(dst, '"')
				return dst, nil
			}
		}
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			n := *(*uint64)(v)
			if n == 0 {
				return append(dst, '"', '0', '"'), nil
			}
			dst = append(dst, '"')
			dst = zstr.AppendUint64(dst, n)
			dst = append(dst, '"')
			return dst, nil
		}
	}

	if omitEmpty {
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			n := *(*uint64)(v)
			if n == 0 {
				return dst, nil
			}
			return zstr.AppendUint64(dst, n), nil
		}
	}
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		n := *(*uint64)(v)
		if n == 0 {
			return append(dst, '0'), nil
		}
		return zstr.AppendUint64(dst, n), nil
	}
}

func int64Encoder(flags Flags) UnsafeEncoder {
	omitEmpty := flags.Has(OmitEmpty)
	needQuotes := flags.Has(NeedQuotes)

	if needQuotes {
		if omitEmpty {
			return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
				n := *(*int64)(v)
				if n == 0 {
					return dst, nil
				}
				dst = append(dst, '"')
				dst = zstr.AppendInt64(dst, n)
				dst = append(dst, '"')
				return dst, nil
			}
		}
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			n := *(*int64)(v)
			if n == 0 {
				return append(dst, '"', '0', '"'), nil
			}
			dst = append(dst, '"')
			dst = zstr.AppendInt64(dst, n)
			dst = append(dst, '"')
			return dst, nil
		}
	}

	if omitEmpty {
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			n := *(*int64)(v)
			if n == 0 {
				return dst, nil
			}
			return zstr.AppendInt64(dst, n), nil
		}
	}
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		n := *(*int64)(v)
		if n == 0 {
			return append(dst, '0'), nil
		}
		return zstr.AppendInt64(dst, n), nil
	}
}

func uint32Encoder(flags Flags) UnsafeEncoder {
	omitEmpty := flags.Has(OmitEmpty)
	needQuotes := flags.Has(NeedQuotes)

	if needQuotes {
		if omitEmpty {
			return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
				n := *(*uint32)(v)
				if n == 0 {
					return dst, nil
				}
				dst = append(dst, '"')
				dst = zstr.AppendUint64(dst, uint64(n))
				dst = append(dst, '"')
				return dst, nil
			}
		}
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			n := *(*uint32)(v)
			if n == 0 {
				return append(dst, '"', '0', '"'), nil
			}
			dst = append(dst, '"')
			dst = zstr.AppendUint64(dst, uint64(n))
			dst = append(dst, '"')
			return dst, nil
		}
	}

	if omitEmpty {
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			n := *(*uint32)(v)
			if n == 0 {
				return dst, nil
			}
			return zstr.AppendUint64(dst, uint64(n)), nil
		}
	}
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		n := *(*uint32)(v)
		if n == 0 {
			return append(dst, '0'), nil
		}
		return zstr.AppendUint64(dst, uint64(n)), nil
	}
}

func int32Encoder(flags Flags) UnsafeEncoder {
	omitEmpty := flags.Has(OmitEmpty)
	needQuotes := flags.Has(NeedQuotes)

	if needQuotes {
		if omitEmpty {
			return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
				n := *(*int32)(v)
				if n == 0 {
					return dst, nil
				}
				dst = append(dst, '"')
				dst = zstr.AppendInt64(dst, int64(n))
				dst = append(dst, '"')
				return dst, nil
			}
		}
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			n := *(*int32)(v)
			if n == 0 {
				return append(dst, '"', '0', '"'), nil
			}
			dst = append(dst, '"')
			dst = zstr.AppendInt64(dst, int64(n))
			dst = append(dst, '"')
			return dst, nil
		}
	}

	if omitEmpty {
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			n := *(*int32)(v)
			if n == 0 {
				return dst, nil
			}
			return zstr.AppendInt64(dst, int64(n)), nil
		}
	}
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		n := *(*int32)(v)
		if n == 0 {
			return append(dst, '0'), nil
		}
		return zstr.AppendInt64(dst, int64(n)), nil
	}
}

func uint16Encoder(flags Flags) UnsafeEncoder {
	omitEmpty := flags.Has(OmitEmpty)
	needQuotes := flags.Has(NeedQuotes)

	if needQuotes {
		if omitEmpty {
			return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
				n := *(*uint16)(v)
				if n == 0 {
					return dst, nil
				}
				dst = append(dst, '"')
				dst = zstr.AppendUint64(dst, uint64(n))
				dst = append(dst, '"')
				return dst, nil
			}
		}
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			n := *(*uint16)(v)
			if n == 0 {
				return append(dst, '"', '0', '"'), nil
			}
			dst = append(dst, '"')
			dst = zstr.AppendUint64(dst, uint64(n))
			dst = append(dst, '"')
			return dst, nil
		}
	}

	if omitEmpty {
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			n := *(*uint16)(v)
			if n == 0 {
				return dst, nil
			}
			return zstr.AppendUint64(dst, uint64(n)), nil
		}
	}
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		n := *(*uint16)(v)
		if n == 0 {
			return append(dst, '0'), nil
		}
		return zstr.AppendUint64(dst, uint64(n)), nil
	}
}

func int16Encoder(flags Flags) UnsafeEncoder {
	omitEmpty := flags.Has(OmitEmpty)
	needQuotes := flags.Has(NeedQuotes)

	if needQuotes {
		if omitEmpty {
			return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
				n := *(*int16)(v)
				if n == 0 {
					return dst, nil
				}
				dst = append(dst, '"')
				dst = zstr.AppendInt64(dst, int64(n))
				dst = append(dst, '"')
				return dst, nil
			}
		}
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			n := *(*int16)(v)
			if n == 0 {
				return append(dst, '"', '0', '"'), nil
			}
			dst = append(dst, '"')
			dst = zstr.AppendInt64(dst, int64(n))
			dst = append(dst, '"')
			return dst, nil
		}
	}

	if omitEmpty {
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			n := *(*int16)(v)
			if n == 0 {
				return dst, nil
			}
			return zstr.AppendInt64(dst, int64(n)), nil
		}
	}
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		n := *(*int16)(v)
		if n == 0 {
			return append(dst, '0'), nil
		}
		return zstr.AppendInt64(dst, int64(n)), nil
	}
}

func uint8Encoder(flags Flags) UnsafeEncoder {
	omitEmpty := flags.Has(OmitEmpty)
	needQuotes := flags.Has(NeedQuotes)

	if needQuotes {
		if omitEmpty {
			return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
				n := *(*uint8)(v)
				if n == 0 {
					return dst, nil
				}
				dst = append(dst, '"')
				dst = zstr.AppendUint8(dst, n)
				dst = append(dst, '"')
				return dst, nil
			}
		}
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			n := *(*uint8)(v)
			if n == 0 {
				return append(dst, '"', '0', '"'), nil
			}
			dst = append(dst, '"')
			dst = zstr.AppendUint8(dst, n)
			dst = append(dst, '"')
			return dst, nil
		}
	}

	if omitEmpty {
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			n := *(*uint8)(v)
			if n == 0 {
				return dst, nil
			}
			return zstr.AppendUint8(dst, n), nil
		}
	}
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		n := *(*uint8)(v)
		if n == 0 {
			return append(dst, '0'), nil
		}
		return zstr.AppendUint8(dst, n), nil
	}
}

func int8Encoder(flags Flags) UnsafeEncoder {
	omitEmpty := flags.Has(OmitEmpty)
	needQuotes := flags.Has(NeedQuotes)

	if needQuotes {
		if omitEmpty {
			return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
				n := *(*int8)(v)
				if n == 0 {
					return dst, nil
				}
				dst = append(dst, '"')
				dst = zstr.AppendInt8(dst, n)
				dst = append(dst, '"')
				return dst, nil
			}
		}
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			n := *(*int8)(v)
			if n == 0 {
				return append(dst, '"', '0', '"'), nil
			}
			dst = append(dst, '"')
			dst = zstr.AppendInt8(dst, n)
			dst = append(dst, '"')
			return dst, nil
		}
	}

	if omitEmpty {
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			n := *(*int8)(v)
			if n == 0 {
				return dst, nil
			}
			return zstr.AppendInt8(dst, n), nil
		}
	}
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		n := *(*int8)(v)
		if n == 0 {
			return append(dst, '0'), nil
		}
		return zstr.AppendInt8(dst, n), nil
	}
}

func float32Encoder(flags Flags) UnsafeEncoder {
	omitEmpty := flags.Has(OmitEmpty)
	needQuotes := flags.Has(NeedQuotes)

	if needQuotes {
		if omitEmpty {
			return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
				n := *(*float32)(v)
				if n == 0 && omitEmpty {
					return dst, nil
				}
				dst = append(dst, '"')
				dst = strconv.AppendFloat(dst, float64(n), 'f', -1, 32)
				dst = append(dst, '"')
				return dst, nil
			}
		}
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			n := *(*float32)(v)
			if n == 0 {
				return append(dst, '"', '0', '"'), nil
			}
			dst = append(dst, '"')
			dst = strconv.AppendFloat(dst, float64(n), 'f', -1, 32)
			dst = append(dst, '"')
			return dst, nil
		}
	}

	if omitEmpty {
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			n := *(*float32)(v)
			if n == 0 {
				return dst, nil
			}
			return strconv.AppendFloat(dst, float64(n), 'f', -1, 32), nil
		}
	}
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		n := *(*float32)(v)
		if n == 0 {
			return append(dst, '0'), nil
		}
		return strconv.AppendFloat(dst, float64(n), 'f', -1, 32), nil
	}
}

func float64Encoder(flags Flags) UnsafeEncoder {
	omitEmpty := flags.Has(OmitEmpty)
	needQuotes := flags.Has(NeedQuotes)

	if needQuotes {
		if omitEmpty {
			return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
				n := *(*float64)(v)
				if n == 0 && omitEmpty {
					return dst, nil
				}
				dst = append(dst, '"')
				dst = strconv.AppendFloat(dst, n, 'f', -1, 64)
				dst = append(dst, '"')
				return dst, nil
			}
		}
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			n := *(*float64)(v)
			if n == 0 {
				return append(dst, '"', '0', '"'), nil
			}
			dst = append(dst, '"')
			dst = strconv.AppendFloat(dst, n, 'f', -1, 64)
			dst = append(dst, '"')
			return dst, nil
		}
	}

	if omitEmpty {
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			n := *(*float64)(v)
			if n == 0 {
				return dst, nil
			}
			return strconv.AppendFloat(dst, n, 'f', -1, 64), nil
		}
	}
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		n := *(*float64)(v)
		if n == 0 {
			return append(dst, '0'), nil
		}
		return strconv.AppendFloat(dst, n, 'f', -1, 64), nil
	}
}
