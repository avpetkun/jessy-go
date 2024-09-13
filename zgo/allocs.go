package zgo

import "unsafe"

//go:linkname Mallocgc runtime.mallocgc
func Mallocgc(size uintptr, typ *Type, needzero bool) unsafe.Pointer
