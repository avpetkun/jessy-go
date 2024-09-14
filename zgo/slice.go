package zgo

import "unsafe"

type Slice struct {
	Data unsafe.Pointer
	Len  uintptr
	Cap  uintptr
}

func MakeSliceBytes(data unsafe.Pointer, size, cap uintptr) []byte {
	return *(*[]byte)(unsafe.Pointer(&Slice{
		Data: data,
		Len:  size,
		Cap:  cap,
	}))
}

func AppendBytesFrame(s []byte, n int) (newS, frame []byte) {
	s0 := len(s)
	s1 := s0 + n
	if s1 > cap(s) {
		s = append(s, make([]byte, n)...)
	}
	newS = s[:s1]
	frame = s[s0:s1]
	return
}
