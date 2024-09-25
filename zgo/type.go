package zgo

import (
	"reflect"
	"unsafe"
)

//go:linkname toRType reflect.toType
//go:noescape
func toRType(*Type) reflect.Type

const (
	// TODO (khr, drchase) why aren't these in TFlag?  Investigate, fix if possible.
	kindDirectIface = 1 << 5
	kindGCProg      = 1 << 6 // Type.gc points to GC program
	kindMask        = (1 << 5) - 1
)

type Type struct {
	Size        uintptr
	PtrBytes    uintptr // number of (prefix) bytes in the type that can contain pointers
	Hash        uint32  // hash of type; avoids computation in hash tables
	TFlag       uint8   // extra type information flags
	Align_      uint8   // alignment of variable with this type
	FieldAlign_ uint8   // alignment of struct field with this type
	Kind_       uint8   // enumeration for C
	// function for comparing objects of this type
	// (ptr to object A, ptr to object B) -> ==?
	Equal func(unsafe.Pointer, unsafe.Pointer) bool
	// GCData stores the GC type data for the garbage collector.
	// If the KindGCProg bit is set in kind, GCData is a GC program.
	// Otherwise it is a ptrmask bitmap. See mbitmap.go for details.
	GCData    *byte
	Str       int32 // string form
	PtrToThis int32 // type for pointer to this type, may be zero
}

func NewTypeFromRType(rt reflect.Type) *Type {
	return (*Type)(UnpackEface(rt).Data)
}

func NewTypeFor[T any]() *Type {
	var v T
	return UnpackEface(v).Type
}

func (t *Type) Native() reflect.Type {
	return toRType(t)
}

func (t *Type) Kind() reflect.Kind {
	return reflect.Kind(t.Kind_ & kindMask)
}

// IfaceIndir reports whether t is stored indirectly in an interface value.
func (t *Type) IfaceIndir() bool {
	return t.Kind_&kindDirectIface == 0
}

// isDirectIface reports whether t is stored directly in an interface value.
func (t *Type) IsDirectIface() bool {
	return t.Kind_&kindDirectIface != 0
}
