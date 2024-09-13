package zstr

import "unsafe"

func bytesAllocFrame(s []byte, n int) (newS, frame []byte) {
	s0 := len(s)
	s1 := s0 + n
	if s1 > cap(s) {
		s = append(s, make([]byte, n)...)
	}
	newS = s[:s1]
	frame = s[s0:s1]
	return
}

func B2S(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}

func S2B(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}
