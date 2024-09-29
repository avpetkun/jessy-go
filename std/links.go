package std

import (
	"reflect"
	_ "unsafe"
)

//go:linkname isValidNumber encoding/json.isValidNumber
func isValidNumber(s string) bool

//go:linkname typeFields encoding/json.typeFields
func typeFields(t reflect.Type) structFields

//go:linkname unquoteBytes encoding/json.unquoteBytes
func unquoteBytes(s []byte) (t []byte, ok bool)
