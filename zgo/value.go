package zgo

import (
	"reflect"
	"unsafe"
)

//go:linkname ifaceIndir reflect.ifaceIndir
//go:noescape
func ifaceIndir(*Type) bool

type Value struct {
	typ     *Type
	ptr     unsafe.Pointer
	uintptr // flag
}

func NewRValuerForRType(rt reflect.Type) func(ptr unsafe.Pointer) reflect.Value {
	eface := UnpackEface(rt)
	gt := (*Type)(eface.Value)

	flag := uintptr(rt.Kind())
	if ifaceIndir(gt) {
		flag |= 1 << 7
	}

	return func(ptr unsafe.Pointer) reflect.Value {
		val := Value{gt, ptr, flag}
		return *(*reflect.Value)(unsafe.Pointer(&val))
	}
}

func NewAnyInterfacerFromRType(rt reflect.Type) func(valPtr unsafe.Pointer) any {
	eface := UnpackEface(rt)
	gt := (*Type)(eface.Value)

	flag := uintptr(rt.Kind())
	if ifaceIndir(gt) {
		flag |= 1 << 7
	}

	var valType *Type

	return func(valPtr unsafe.Pointer) (i any) {
		if valType == nil {
			gVal := Value{gt, valPtr, flag}
			rVal := (*reflect.Value)(unsafe.Pointer(&gVal))
			i = rVal.Interface()
			valType = (*EmptyInterface)(unsafe.Pointer(&i)).Type
			return
		}
		eface := (*EmptyInterface)(unsafe.Pointer(&i))
		eface.Type = valType
		eface.Value = valPtr
		return
	}
}

func NewInterfacerFromRType[I any](rt reflect.Type) func(valPtr unsafe.Pointer) I {
	eface := UnpackEface(rt)
	gt := (*Type)(eface.Value)

	flag := uintptr(rt.Kind())
	if ifaceIndir(gt) {
		flag |= 1 << 7
	}

	var valType *Type

	return func(valPtr unsafe.Pointer) (i I) {
		if valType == nil {
			gVal := Value{gt, valPtr, flag}
			rVal := (*reflect.Value)(unsafe.Pointer(&gVal))
			i = rVal.Interface().(I)
			valType = (*EmptyInterface)(unsafe.Pointer(&i)).Type
			return
		}
		eface := (*EmptyInterface)(unsafe.Pointer(&i))
		eface.Type = valType
		eface.Value = valPtr
		return
	}
}
