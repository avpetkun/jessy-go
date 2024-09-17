package zgo

import (
	"reflect"
	"sync"
	"unsafe"
)

//go:linkname mapIterInitType runtime.mapiterinit
func mapIterInitType(t *Type, m *Map, it *MapIterator)

//go:linkname mapIterInitPointer runtime.mapiterinit
func mapIterInitPointer(t unsafe.Pointer, m *Map, it *MapIterator)

//go:linkname mapIterNext runtime.mapiternext
func mapIterNext(it *MapIterator)

func NewMapIteratorFromValue(value any) (it *MapIterator, count int) {
	it = mapIteratorPool.Get().(*MapIterator)
	eface := *(*EmptyInterface)(unsafe.Pointer(&value))
	if eface.Type == nil || eface.Value == nil {
		return
	}
	hmap := (*Map)(eface.Value)
	mapIterInitType(eface.Type, hmap, it)
	count = hmap.Count
	return
}

func NewMapIteratorFromRType(rType reflect.Type, isDirectValue bool) (getIterator func(valuePtr unsafe.Pointer) (it *MapIterator, count int)) {
	mapType := UnpackEface(rType).Value

	if isDirectValue {
		return func(value unsafe.Pointer) (it *MapIterator, count int) {
			if value == nil {
				return
			}
			hmap := (*Map)(value)
			it = mapIteratorPool.Get().(*MapIterator)
			mapIterInitPointer(mapType, hmap, it)
			count = hmap.Count
			return
		}
	}
	return func(value unsafe.Pointer) (it *MapIterator, count int) {
		value = *(*unsafe.Pointer)(value)
		if value == nil {
			return
		}
		hmap := (*Map)(value)
		it = mapIteratorPool.Get().(*MapIterator)
		mapIterInitPointer(mapType, hmap, it)
		count = hmap.Count
		return
	}
}

var mapIteratorPool = sync.Pool{New: func() any { return new(MapIterator) }}

type MapIterator struct {
	Key         unsafe.Pointer // Must be in first position.  Write nil to indicate iteration end (see cmd/compile/internal/walk/range.go).
	Elem        unsafe.Pointer // Must be in second position (see cmd/compile/internal/walk/range.go).
	T           *MapType
	H           *Map
	Buckets     unsafe.Pointer    // bucket ptr at hash_iter initialization time
	Bptr        *unsafe.Pointer   // current bucket
	Overflow    *[]unsafe.Pointer // keeps overflow buckets of hmap.buckets alive
	OldOverflow *[]unsafe.Pointer // keeps overflow buckets of hmap.oldbuckets alive
	StartBucket uintptr           // bucket iteration started at
	Offset      uint8             // intra-bucket offset to start from during iteration (should be big enough to hold bucketCnt-1)
	Wrapped     bool              // already wrapped around from end of bucket array to beginning
	B           uint8
	I           uint8
	Bucket      uintptr
	CheckBucket uintptr
}

func (it *MapIterator) Next() {
	mapIterNext(it)
}

func (it *MapIterator) Release() {
	*it = MapIterator{}
	mapIteratorPool.Put(it)
}

type Map struct {
	Count      int // # live cells == size of map.  Must be first (used by len() builtin)
	Flags      uint8
	B          uint8          // log_2 of # of buckets (can hold up to loadFactor * 2^B items)
	NOverflow  uint16         // approximate number of overflow buckets; see incrnoverflow for details
	Hash0      uint32         // hash seed
	Buckets    unsafe.Pointer // array of 2^B Buckets. may be nil if count==0.
	OldBuckets unsafe.Pointer // previous bucket array of half the size, non-nil only when growing
	NEvacuate  uintptr        // progress counter for evacuation (buckets less than this have been evacuated)
	Extra      unsafe.Pointer // optional fields
}

type MapType struct {
	Type
	Key    *Type
	Elem   *Type
	Bucket *Type // internal type representing a hash bucket
	// function for hashing keys (ptr to key, seed) -> hash
	Hasher     func(unsafe.Pointer, uintptr) uintptr
	KeySize    uint8  // size of key slot
	ValueSize  uint8  // size of elem slot
	BucketSize uint16 // size of bucket
	Flags      uint32
}
