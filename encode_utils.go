package jessy

import (
	"reflect"

	"github.com/avpetkun/jessy-go/zgo"
)

func getIndent(n uint32) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = '\t'
	}
	return zgo.B2S(b)
}

func tReallyImplements(t, interfaceType reflect.Type) bool {
	if t.Implements(interfaceType) {
		if t.Kind() == reflect.Pointer {
			t = t.Elem()
		}
		if t.Kind() == reflect.Struct {
			for i := range t.NumField() {
				f := t.Field(i)
				if f.Anonymous && f.Type.Implements(interfaceType) {
					return false
				}
			}
		}
		return true
	}
	return false
}

func tReallyStruct(t reflect.Type) bool {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t.Kind() == reflect.Struct
}

func tImplementsAny(t reflect.Type) bool {
	if t.Implements(typeAppendMarshaler) || t.Implements(typeMarshaler) || t.Implements(typeTextMarshaler) {
		return true
	}
	t = reflect.PointerTo(t)
	return t.Implements(typeAppendMarshaler) || t.Implements(typeMarshaler) || t.Implements(typeTextMarshaler)
}
