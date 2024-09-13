package zgo

import "unsafe"

type SliceHeader struct {
	Data uintptr
	Len  uintptr
	Cap  uintptr
}

func MakeSliceBytes(data unsafe.Pointer, size, cap uintptr) []byte {
	return *(*[]byte)(unsafe.Pointer(&SliceHeader{
		Data: uintptr(data),
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

// problems ?
func MakeDirtyBytes(size int) []byte {
	usize := uintptr(size)
	p := Mallocgc(usize, nil, false)
	h := SliceHeader{
		Data: uintptr(p),
		Len:  usize,
		Cap:  usize,
	}
	return *(*[]byte)(unsafe.Pointer(&h))
}
