package jessy

import "reflect"

func tReallyImplements(t, inter reflect.Type) bool {
	if t.Implements(inter) {
		if t.Kind() == reflect.Struct {
			for i := range t.NumField() {
				f := t.Field(i)
				if f.Anonymous && f.Type.Implements(inter) {
					return false
				}
			}
		}
		return true
	}
	return false
}

func tImplementsAny(t reflect.Type) bool {
	switch {
	case t.Implements(typeAppendMarshaler):
	case t.Implements(typeMarshaler):
	case t.Implements(typeTextMarshaler):
	default:
		tp := reflect.PointerTo(t)
		switch {
		case tp.Implements(typeAppendMarshaler):
		case tp.Implements(typeMarshaler):
		case tp.Implements(typeTextMarshaler):
		default:
			return false
		}
	}
	return true
}
