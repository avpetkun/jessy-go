package jessy

import (
	"math/big"
	"reflect"
	"unsafe"

	"github.com/avpetkun/jessy-go/zstr"
)

var typeBigInt = reflect.TypeFor[big.Int]()

func bigIntEncoder(flags Flags) UnsafeEncoder {
	omitEmpty := flags.Has(OmitEmpty)

	if flags.Has(NeedQuotes) {
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			b := *(*big.Int)(v)
			bits := b.Bits()
			if len(bits) == 0 {
				if omitEmpty {
					return dst, nil
				}
				return append(dst, '"', '0', '"'), nil
			}
			dst = append(dst, '"')
			if len(bits) == 1 {
				if b.Sign() == -1 {
					dst = append(dst, '-')
				}
				dst = zstr.AppendUint64(dst, uint64(bits[0]))
			} else {
				dst = b.Append(dst, 10)
			}
			dst = append(dst, '"')
			return dst, nil
		}
	}
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		b := *(*big.Int)(v)
		bits := b.Bits()
		if len(bits) == 0 {
			if omitEmpty {
				return dst, nil
			}
			return append(dst, '0'), nil
		}
		if len(bits) == 1 {
			if b.Sign() == -1 {
				dst = append(dst, '-')
			}
			return zstr.AppendUint64(dst, uint64(bits[0])), nil
		}
		return b.Append(dst, 10), nil
	}
}
