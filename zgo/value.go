package zgo

import (
	"reflect"
	"unsafe"
)

type Value struct {
	Type *Type
	VPtr unsafe.Pointer
	Flag uintptr // flag
}

func (v Value) Native() reflect.Value {
	return *(*reflect.Value)(unsafe.Pointer(&v))
}

func NewValueFromRType(rType reflect.Type, valuePtr unsafe.Pointer) Value {
	typ := TypeFromRType(rType)
	flag := uintptr(rType.Kind())
	if typ.IfaceIndir() {
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
		eface.Data = valPtr
		return
	}
}

func NewInterfacerFromRType[I any](rType reflect.Type) func(valPtr unsafe.Pointer) I {
	if reflect.TypeFor[I]() == rType {
		return func(valPtr unsafe.Pointer) I {
			return *(*I)(valPtr)
		}
	}

	var i I
	i, ok := NewRValueFromRType(rType, unsafe.Pointer(&i)).Interface().(I)
	if !ok {
		return nil
	}
	valType := (*EmptyInterface)(unsafe.Pointer(&i)).Type

	return func(valPtr unsafe.Pointer) (i I) {
		eface := (*EmptyInterface)(unsafe.Pointer(&i))
		eface.Type = valType
		eface.Data = valPtr
		return
	}
}
