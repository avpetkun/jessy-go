package jessy

import (
	"bytes"

	"github.com/avpetkun/jessy-go/zstr"
)

// HTMLEscape appends to dst the JSON-encoded src with <, >, &, U+2028 and U+2029
// characters inside string literals changed to \u003c, \u003e, \u0026, \u2028, \u2029
// so that the JSON will be safe to embed inside HTML <script> tags.
// For historical reasons, web browsers don't honor standard HTML
// escaping within <script> tags, so an alternative JSON encoding must be used.
func HTMLEscape(dst *bytes.Buffer, src []byte) {
	dst.Grow(len(src))
	dst.Write(zstr.AppendHTMLEscape(dst.AvailableBuffer(), src))
}

// Indent appends to dst an indented form of the JSON-encoded src.
// Each element in a JSON object or array begins on a new,
// indented line beginning with prefix followed by one or more
// copies of indent according to the indentation nesting.
// The data appended to dst does not begin with the prefix nor
// any indentation, to make it easier to embed inside other formatted JSON data.
// Although leading space characters (space, tab, carriage return, newline)
// at the beginning of src are dropped, trailing space characters
// at the end of src are preserved and copied to dst.
// For example, if src has no trailing spaces, neither will dst;
// if src ends in a trailing newline, so will dst.
func Indent(dst *bytes.Buffer, src []byte, prefix, indent string) (err error) {
	dst.Grow(len(src) * 2)
	b := dst.AvailableBuffer()
	b = zstr.AppendIndent(b, src, prefix, indent)
	_, err = dst.Write(b)
	return
}

// Compact appends to dst the JSON-encoded src with
// insignificant space characters elided.
func Compact(dst *bytes.Buffer, src []byte) (err error) {
	dst.Grow(len(src) * 2)
	b := dst.AvailableBuffer()
	b = zstr.AppendCompactJSON(b, src, false)
	_, err = dst.Write(b)
	return
}
