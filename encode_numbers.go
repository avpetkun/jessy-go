package jessy

import (
	"errors"
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
				n := float64(*(*float32)(v))
				if n == 0 {
					return dst, nil
				}
				if math.IsNaN(n) || math.IsInf(n, 0) {
					return dst, errFloatNum
				}
				dst = append(dst, '"')
				dst = appendFloat32(dst, n)
				dst = append(dst, '"')
				return dst, nil
			}
		}
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			n := float64(*(*float32)(v))
			if n == 0 {
				return append(dst, '"', '0', '"'), nil
			}
			if math.IsNaN(n) || math.IsInf(n, 0) {
				return dst, errFloatNum
			}
			dst = append(dst, '"')
			dst = appendFloat32(dst, n)
			dst = append(dst, '"')
			return dst, nil
		}
	}

	if omitEmpty {
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			n := float64(*(*float32)(v))
			if n == 0 {
				return dst, nil
			}
			if math.IsNaN(n) || math.IsInf(n, 0) {
				return dst, errFloatNum
			}
			return appendFloat32(dst, n), nil
		}
	}
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		n := float64(*(*float32)(v))
		if n == 0 {
			return append(dst, '0'), nil
		}
		if math.IsNaN(n) || math.IsInf(n, 0) {
			return dst, errFloatNum
		}
		return appendFloat32(dst, n), nil
	}
}

func float64Encoder(flags Flags) UnsafeEncoder {
	omitEmpty := flags.Has(OmitEmpty)
	needQuotes := flags.Has(NeedQuotes)

	if needQuotes {
		if omitEmpty {
			return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
				n := *(*float64)(v)
				if n == 0 {
					return dst, nil
				}
				if math.IsNaN(n) || math.IsInf(n, 0) {
					return dst, errFloatNum
				}
				dst = append(dst, '"')
				dst = appendFloat64(dst, n)
				dst = append(dst, '"')
				return dst, nil
			}
		}
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			n := *(*float64)(v)
			if n == 0 {
				return append(dst, '"', '0', '"'), nil
			}
			if math.IsNaN(n) || math.IsInf(n, 0) {
				return dst, errFloatNum
			}
			dst = append(dst, '"')
			dst = appendFloat64(dst, n)
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
			if math.IsNaN(n) || math.IsInf(n, 0) {
				return dst, errFloatNum
			}
			return appendFloat64(dst, n), nil
		}
	}
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		n := *(*float64)(v)
		if n == 0 {
			return append(dst, '0'), nil
		}
		if math.IsNaN(n) || math.IsInf(n, 0) {
			return dst, errFloatNum
		}
		return appendFloat64(dst, n), nil
	}
}

func complex64Encoder(flags Flags) UnsafeEncoder {
	if flags.Has(OmitEmpty) {
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			n := *(*complex64)(v)
			if n == 0 {
				return dst, nil
			}
			r, i := float64(real(n)), float64(imag(n))
			if math.IsNaN(r) || math.IsInf(r, 0) {
				return dst, errComplexNum
			}
			if math.IsNaN(i) || math.IsInf(i, 0) {
				return dst, errComplexNum
			}
			return appendQuotedComplex(dst, r, i), nil
		}
	}
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		n := *(*complex64)(v)
		if n == 0 {
			return append(dst, '"', '0', '+', '0', 'i', '"'), nil
		}
		r, i := float64(real(n)), float64(imag(n))
		if math.IsNaN(r) || math.IsInf(r, 0) {
			return dst, errComplexNum
		}
		if math.IsNaN(i) || math.IsInf(i, 0) {
			return dst, errComplexNum
		}
		return appendQuotedComplex(dst, r, i), nil
	}
}

func complex128Encoder(flags Flags) UnsafeEncoder {
	if flags.Has(OmitEmpty) {
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			n := *(*complex128)(v)
			if n == 0 {
				return dst, nil
			}
			r, i := real(n), imag(n)
			if math.IsNaN(r) || math.IsInf(r, 0) {
				return dst, errComplexNum
			}
			if math.IsNaN(i) || math.IsInf(i, 0) {
				return dst, errComplexNum
			}
			return appendQuotedComplex(dst, r, i), nil
		}
	}
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		n := *(*complex128)(v)
		if n == 0 {
			return append(dst, '"', '0', '+', '0', 'i', '"'), nil
		}
		r, i := real(n), imag(n)
		if math.IsNaN(r) || math.IsInf(r, 0) {
			return dst, errComplexNum
		}
		if math.IsNaN(i) || math.IsInf(i, 0) {
			return dst, errComplexNum
		}
		return appendQuotedComplex(dst, r, i), nil
	}
}

var errFloatNum = errors.New("float number is NaN or +-Inf")
var errComplexNum = errors.New("complex float parts is NaN or +-Inf")

func appendQuotedComplex(dst []byte, real, imag float64) []byte {
	dst = append(dst, '"')

	dst = appendFloat64(dst, real)

	dst = append(dst, '+')
	imIdx := len(dst)

	dst = appendFloat64(dst, imag)
	dst = append(dst, 'i')
	dst = append(dst, '"')

	if dst[imIdx] == '-' {
		copy(dst[imIdx-1:], dst[imIdx:])
		dst = dst[:len(dst)-1]
	}
	return dst
}

// from encoding/json
func appendFloat32(b []byte, f float64) []byte {
	abs := math.Abs(f)
	fmt := byte('f')
	if abs != 0 && (float32(abs) < 1e-6 || float32(abs) >= 1e21) {
		fmt = 'e'
	}
	b = strconv.AppendFloat(b, f, fmt, -1, 32)
	if fmt == 'e' {
		// clean up e-09 to e-9
		n := len(b)
		if n >= 4 && b[n-4] == 'e' && b[n-3] == '-' && b[n-2] == '0' {
			b[n-2] = b[n-1]
			b = b[:n-1]
		}
	}
	return b
}

// from encoding/json
func appendFloat64(b []byte, f float64) []byte {
	abs := math.Abs(f)
	fmt := byte('f')
	if abs != 0 && (abs < 1e-6 || abs >= 1e21) {
		fmt = 'e'
	}
	b = strconv.AppendFloat(b, f, fmt, -1, 64)
	if fmt == 'e' {
		// clean up e-09 to e-9
		n := len(b)
		if n >= 4 && b[n-4] == 'e' && b[n-3] == '-' && b[n-2] == '0' {
			b[n-2] = b[n-1]
			b = b[:n-1]
		}
	}
	return b
}
