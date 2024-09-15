package zgo

import "unsafe"

//go:linkname growSlice runtime.growslice
func growSlice(oldPtr unsafe.Pointer, newLen, oldCap, num uint, et *Type) Slice

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

func Grow[S ~[]E, E any](s S, n int) S {
	if n -= cap(s) - len(s); n > 0 {
		if cap(s) == 0 {
			return make(S, 0, n)
		}
		var e E
		eType := UnpackEface(e).Type

		a := (*Slice)(unsafe.Pointer(&s))
		newCap := a.Cap + uint(n)
		b := growSlice(a.Data, newCap, a.Cap, a.Len, eType)
		a.Data = b.Data
		a.Cap = b.Cap
	}
	return s
}

func GrowBytes(s []byte, n int) []byte {
	if n -= cap(s) - len(s); n > 0 {
		if cap(s) == 0 {
			return make([]byte, 0, n)
		}
		a := (*Slice)(unsafe.Pointer(&s))
		newCap := a.Cap + uint(n)
		b := growSlice(a.Data, newCap, a.Cap, a.Len, byteType)
		a.Data = b.Data
		a.Cap = b.Cap
	}
	return s
}

func GrowLen[S ~[]E, E any](s S, n int) S {
	n += len(s)
	if d := n - cap(s); d > 0 {
		if cap(s) == 0 {
			return make(S, n)
		}
		var e E
		eType := UnpackEface(e).Type

		a := (*Slice)(unsafe.Pointer(&s))
		newCap := a.Cap + uint(d)
		b := growSlice(a.Data, newCap, a.Cap, a.Len, eType)
		a.Data = b.Data
		a.Cap = b.Cap
		a.Len = uint(n)
		return s
	}
	return s[:n]
}

func GrowBytesLen[S ~[]E, E any](s S, n int) S {
	n += len(s)
	if d := n - cap(s); d > 0 {
		if cap(s) == 0 {
			return make(S, n)
		}
		a := (*Slice)(unsafe.Pointer(&s))
		newCap := a.Cap + uint(d)
		b := growSlice(a.Data, newCap, a.Cap, a.Len, byteType)
		a.Data = b.Data
		a.Cap = b.Cap
		a.Len = uint(n)
		return s
	}
	return s[:n]
}
