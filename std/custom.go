package std

import "fmt"

func parseUint64(s []byte) (v uint64, err error) {
	var c byte
	for _, c = range s {
		c -= '0'
		if c > 9 {
			v = 0
			err = fmt.Errorf("invalid uint64 number: %s", s)
			return
		}
		v = v*10 + uint64(c)
	}
	return
}

func parseInt64(s []byte) (v int64, err error) {
	var neg bool
	if s[0] == '-' {
		s = s[1:]
		neg = true
	}
	var c byte
	for _, c = range s {
		c -= '0'
		if c > 9 {
			v = 0
			err = fmt.Errorf("invalid int64 number: %s", s)
			return
		}
		v = v*10 + int64(c)
	}
	if neg {
		v = -v
	}
	return
}
