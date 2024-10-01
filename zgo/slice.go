package zgo

import "unsafe"

type Slice struct {
	Data unsafe.Pointer
	Len  uint
	Cap  uint
}

func NewSliceBytes(data unsafe.Pointer, size, cap uint) []byte {
	return *(*[]byte)(unsafe.Pointer(&Slice{
		Data: data,
		Len:  size,
		Cap:  cap,
	}))
}
