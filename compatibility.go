package jessy

import (
	"bytes"
	"encoding/json"
)

func Indent(dst *bytes.Buffer, src []byte, prefix, indent string) error {
	return json.Indent(dst, src, prefix, indent)
}

func HTMLEscape(dst *bytes.Buffer, src []byte) {
	json.HTMLEscape(dst, src)
}

func Compact(dst *bytes.Buffer, src []byte) error {
	return json.Compact(dst, src)
}

func Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
