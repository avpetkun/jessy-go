package zgo

import (
	"reflect"
	"unsafe"
)

//go:linkname ifaceIndir reflect.ifaceIndir
//go:noescape
func ifaceIndir(*Type) bool

type Value struct {
	Type *Type
	VPtr unsafe.Pointer
	Flag uintptr // flag
}

func (v Value) Native() reflect.Value {
	return *(*reflect.Value)(unsafe.Pointer(&v))
}

func NewValueFromRType(rType reflect.Type, valuePtr unsafe.Pointer) Value {
	eface := UnpackEface(rType)
	typ := (*Type)(eface.Value)
	flag := uintptr(rType.Kind())
	if ifaceIndir(typ) {
		flag |= 1 << 7
	}
	return Value{typ, valuePtr, flag}
}

func NewRValueFromRType(rType reflect.Type, valuePtr unsafe.Pointer) reflect.Value {
	return NewValueFromRType(rType, valuePtr).Native()
}

func NewRValuerFromRType(rType reflect.Type) func(ptr unsafe.Pointer) reflect.Value {
	rVal := NewValueFromRType(rType, nil)
	return func(valPtr unsafe.Pointer) reflect.Value {
		val := rVal
		val.VPtr = valPtr
		return val.Native()
	}
}

func NewAnyInterfacerFromRType(rType reflect.Type) func(valPtr unsafe.Pointer) any {
	i := NewRValueFromRType(rType, nil).Interface()
	valType := (*EmptyInterface)(unsafe.Pointer(&i)).Type

	return func(valPtr unsafe.Pointer) (i any) {
		eface := (*EmptyInterface)(unsafe.Pointer(&i))
		eface.Type = valType
		eface.Value = valPtr
		return
	}
}

func NewInterfacerFromRType[I any](rType reflect.Type) func(valPtr unsafe.Pointer) I {
	i := NewRValueFromRType(rType, nil).Interface().(I)
	valType := (*EmptyInterface)(unsafe.Pointer(&i)).Type

	return func(valPtr unsafe.Pointer) (i I) {
		eface := (*EmptyInterface)(unsafe.Pointer(&i))
		eface.Type = valType
		eface.Value = valPtr
		return
	}
}
