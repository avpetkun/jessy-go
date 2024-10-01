// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// package std implements encoding and decoding of JSON as defined in
// RFC 7159. The mapping between JSON and Go values is described
// in the documentation for the Marshal and Unmarshal functions.
//
// See "JSON and Go" for an introduction to this package:
// https://golang.org/doc/articles/json_and_go.html
package std

import (
	"reflect"
	"sync"
)

type structFields struct {
	List         []field
	byExactName  map[string]*field
	byFoldedName map[string]*field
}

// A field represents a single field found in a struct.
type field struct {
	name      string
	NameBytes []byte

	NameNonEsc  string
	NameEscHTML string

	Tag       bool
	index     []int
	Typ       reflect.Type
	OmitEmpty bool
	quoted    bool
}

var fieldCache sync.Map // map[reflect.Type]structFields

// cachedTypeFields is like typeFields but uses a cache to avoid repeated work.
func cachedTypeFields(t reflect.Type) structFields {
	if f, ok := fieldCache.Load(t); ok {
		return f.(structFields)
	}
	f, _ := fieldCache.LoadOrStore(t, typeFields(t))
	return f.(structFields)
}
