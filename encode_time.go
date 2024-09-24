package jessy

import (
	"reflect"
	"time"
	"unsafe"
)

var timeType = reflect.TypeFor[time.Time]()

func timeEncoder(flags Flags) UnsafeEncoder {
	if flags.Has(OmitEmpty) {
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			t := *(*time.Time)(v)

			if t.IsZero() {
				return dst, nil
			}

			dst = append(dst, '"')
			dst = t.AppendFormat(dst, time.RFC3339Nano)
			dst = append(dst, '"')

			return dst, nil
		}
	}
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		t := *(*time.Time)(v)

		dst = append(dst, '"')
		dst = t.AppendFormat(dst, time.RFC3339Nano)
		dst = append(dst, '"')

		return dst, nil
	}
}
