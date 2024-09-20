package jessy

import "unsafe"

func boolEncoder(flags Flags) UnsafeEncoder {
	omitEmpty := flags.Has(OmitEmpty)
	needQuotes := flags.Has(NeedQuotes)

	if needQuotes {
		if omitEmpty {
			return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
				if *(*bool)(v) {
					return append(dst, '"', 't', 'r', 'u', 'e', '"'), nil
				}
				return dst, nil
			}
		}
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			if *(*bool)(v) {
				return append(dst, '"', 't', 'r', 'u', 'e', '"'), nil
			}
			return append(dst, '"', 'f', 'a', 'l', 's', 'e', '"'), nil
		}
	}

	if omitEmpty {
		return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
			if *(*bool)(v) {
				return append(dst, 't', 'r', 'u', 'e'), nil
			}
			return dst, nil
		}
	}
	return func(dst []byte, v unsafe.Pointer) ([]byte, error) {
		if *(*bool)(v) {
			return append(dst, 't', 'r', 'u', 'e'), nil
		}
		return append(dst, 'f', 'a', 'l', 's', 'e'), nil
	}
}
