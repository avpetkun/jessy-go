package jessy

import (
	"reflect"
	"unsafe"
)

//go:linkname toRType reflect.toType
//go:noescape
func toRType(*goType) reflect.Type

//go:linkname ifaceIndir reflect.ifaceIndir
//go:noescape
func ifaceIndir(*goType) bool

type goEmptyInterface struct {
	Type  *goType
	Value unsafe.Pointer
}

func goUnpackEface(value any) goEmptyInterface {
	return *(*goEmptyInterface)(unsafe.Pointer(&value))
}

type goType struct {
	Size       uintptr
	PtrBytes   uintptr      // number of (prefix) bytes in the type that can contain pointers
	Hash       uint32       // hash of type; avoids computation in hash tables
	TFlag      uint8        // extra type information flags
	Align      uint8        // alignment of variable with this type
	FieldAlign uint8        // alignment of struct field with this type
	Kind       reflect.Kind // enumeration for C
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

func (gt *goType) Native() reflect.Type { return toRType(gt) }

type goValue struct {
	typ     *goType
	ptr     unsafe.Pointer
	uintptr // flag
}

func newRValuerForRType(rt reflect.Type) func(ptr unsafe.Pointer) reflect.Value {
	eface := goUnpackEface(rt)
	gt := (*goType)(eface.Value)

	flag := uintptr(rt.Kind())
	if ifaceIndir(gt) {
		flag |= 1 << 7
	}

	return func(ptr unsafe.Pointer) reflect.Value {
		val := goValue{gt, ptr, flag}
		return *(*reflect.Value)(unsafe.Pointer(&val))
	}
}

type goStringHeader struct {
	Data *byte
	Len  int
}

type goSliceHeader struct {
	Data uintptr
	Len  uintptr
	Cap  uintptr
}
