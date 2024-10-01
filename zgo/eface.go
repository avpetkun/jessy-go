package zgo

import "unsafe"

type EmptyInterface struct {
	Type *Type
	Data unsafe.Pointer
}

func UnpackEface(value any) EmptyInterface {
	return *(*EmptyInterface)(unsafe.Pointer(&value))
}

func PackEface(typ *Type, data unsafe.Pointer) (i any) {
	eface := (*EmptyInterface)(unsafe.Pointer(&i))
	eface.Type = typ
	eface.Data = data
	return
}
