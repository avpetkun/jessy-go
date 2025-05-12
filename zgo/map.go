package zgo

import (
	"reflect"
	"sync"
	"unsafe"
)

func NewMapIteratorFromValue(value any) (it *MapIterator, count int) {
	it = mapIteratorPool.Get().(*MapIterator)
	eface := *(*EmptyInterface)(unsafe.Pointer(&value))
	if eface.Type == nil || eface.Data == nil {
		return
	}
	mt := (*MapType)(unsafe.Pointer(eface.Type))
	hmap := (*Map)(eface.Data)
	it.Init(mt, hmap)
	count = hmap.Len()
	return
}

func NewMapIteratorFromRType(rType reflect.Type) (getIterator func(valuePtr unsafe.Pointer) (it *MapIterator, count int)) {
	mapType := (*MapType)(UnpackEface(rType).Data)
	return func(value unsafe.Pointer) (it *MapIterator, count int) {
		if value == nil {
			return
		}
		hmap := (*Map)(value)
		it = mapIteratorPool.Get().(*MapIterator)
		it.Init(mapType, hmap)
		count = hmap.Len()
		return
	}
}

var mapIteratorPool = sync.Pool{New: func() any { return new(MapIterator) }}

func (it *MapIterator) Release() {
	*it = MapIterator{}
	mapIteratorPool.Put(it)
}
