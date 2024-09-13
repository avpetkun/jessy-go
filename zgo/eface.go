package zgo

import "unsafe"

type EmptyInterface struct {
	Type  *Type
	Value unsafe.Pointer
}

func UnpackEface(value any) EmptyInterface {
	return *(*EmptyInterface)(unsafe.Pointer(&value))
}

func PackEface(typ *Type, value unsafe.Pointer) (i any) {
	eface := (*EmptyInterface)(unsafe.Pointer(&i))
	eface.Type = typ
	eface.Value = value
	return
}
