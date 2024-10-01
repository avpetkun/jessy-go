package zgo

import (
	"reflect"
	"testing"
	"unsafe"
)

func TestMapIter(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}

	it, count := NewMapIteratorFromValue(m)
	if it == nil {
		t.Fatal("map m is nil")
	}
	for range count {
		println("key", *(*string)(it.Key), "val", *(*int)(it.Elem))
		it.Next()
	}
	it.Release()

	getIterator := NewMapIteratorFromRType(reflect.TypeOf(m))
	it, count = getIterator(*(*unsafe.Pointer)(unsafe.Pointer(&m)))
	if it == nil {
		t.Fatal("map m is nil")
	}
	for range count {
		println("key", *(*string)(it.Key), "val", *(*int)(it.Elem))
		it.Next()
	}
	it.Release()
}
