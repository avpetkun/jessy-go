package format

import (
	"fmt"
	"math"
)

var errEmptyNumber = fmt.Errorf("parsing empty number")

func ParseUint64(s []byte) (v uint64, err error) {
	if len(s) == 0 {
		return 0, errEmptyNumber
	}
	var c byte
	var n uint64
	for _, c = range s {
		c -= '0'
		if c > 9 {
			v = 0
			err = fmt.Errorf("invalid number symbol '%s' in number %s", string(c), s)
			return
		}
		n = v*10 + uint64(c)
		if n < v {
			err = fmt.Errorf("too big uint64 number '%s'", s)
			return
		}
		v = n
	}
	return
}

func ParseInt64(s []byte) (v int64, err error) {
	if len(s) == 0 {
		return 0, errEmptyNumber
	}
	var neg bool
	if s[0] == '-' {
		s = s[1:]
		neg = true
	}
	var c byte
	var n int64
	for _, c = range s {
		c -= '0'
		if c > 9 {
			v = 0
			err = fmt.Errorf("invalid number symbol '%s' in number %s", string(c), s)
			return
		}
		n = v*10 + int64(c)
		if n < v && n != math.MinInt64 {
			err = fmt.Errorf("too big int64 number '%s'", s)
			return
		}
		v = n
	}
	if neg {
		v = -v
	}
	return
}
