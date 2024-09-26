package jessy

import "reflect"

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
