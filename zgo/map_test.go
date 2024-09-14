package zgo

import (
	"reflect"
	"testing"
	"unsafe"
)

func TestMapIter(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}

	it, count := NewValueMapIterator(m)
	for range count {
		println("key", *(*string)(it.Key), "val", *(*int)(it.Elem))
		it.Next()
	}

	getIterator := NewPointerMapIteratorForType(reflect.TypeOf(m))
	it, count = getIterator(unsafe.Pointer(&m))
	for range count {
		println("key", *(*string)(it.Key), "val", *(*int)(it.Elem))
		it.Next()
	}
}
