package zstr

func growWindow(s []byte, n int) []byte {
	l := len(s)
	e := l + n
	if e > cap(s) {
		s = append(s[:cap(s)], make([]byte, e-cap(s))...)
		return s[l:]
	}
	return s[l:e]
}

func growCap(s []byte, n int) []byte {
	if n -= cap(s) - len(s); n > 0 {
		s = append(s[:cap(s)], make([]byte, n)...)[:len(s)]
	}
	return s
}
