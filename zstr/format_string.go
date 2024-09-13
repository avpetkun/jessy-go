package zstr

import (
	"encoding/base64"
	"encoding/hex"
	"unicode/utf8"
)

func AppendBase64String(dst, data []byte) []byte {
	size := base64.StdEncoding.EncodedLen(len(data)) + 2

	dst, out := bytesAllocFrame(dst, size)

	out[0] = '"'
	base64.StdEncoding.Encode(out[1:], data)
	out[size-1] = '"'

	return dst
}

func AppendHexString(dst, data []byte) []byte {
	size := len(data)*2 + 4
	dst, out := bytesAllocFrame(dst, size)

	out[0] = '"'
	out[1] = '0'
	out[2] = 'x'
	hex.Encode(out[3:], data)
	out[size-1] = '"'

	return dst
}

func AppendHex(dst, data []byte) []byte {
	size := len(data)*2 + 2
	dst, out := bytesAllocFrame(dst, size)

	out[0] = '0'
	out[1] = 'x'
	hex.Encode(out[2:], data)

	return dst
}

// from encoding/json.AppendString
func AppendString(dst, src []byte) []byte {
	const hex = "0123456789abcdef"

	dst = append(dst, '"')
	start := 0
	srcLen := len(src)
	for i := 0; i < srcLen; i++ {
		if b := src[i]; b < utf8.RuneSelf {
			if safeSet[b] {
				continue
			}
			dst = append(dst, src[start:i]...)
			switch b {
			case '\\', '"':
				dst = append(dst, '\\', b)
			case '\b':
				dst = append(dst, '\\', 'b')
			case '\f':
				dst = append(dst, '\\', 'f')
			case '\n':
				dst = append(dst, '\\', 'n')
			case '\r':
				dst = append(dst, '\\', 'r')
			case '\t':
				dst = append(dst, '\\', 't')
			default:
				// This encodes bytes < 0x20 except for \b, \f, \n, \r and \t.
				// If escapeHTML is set, it also escapes <, >, and &
				// because they can lead to security holes when
				// user-controlled strings are rendered into JSON
				// and served to some browsers.
				dst = append(dst, '\\', 'u', '0', '0', hex[b>>4], hex[b&0xF])
			}
			i++
			start = i
			continue
		}
		// TODO(https://go.dev/issue/56948): Use generic utf8 functionality.
		// For now, cast only a small portion of byte slices to a string
		// so that it can be stack allocated. This slows down []byte slightly
		// due to the extra copy, but keeps string performance roughly the same.
		n := len(src) - i
		if n > utf8.UTFMax {
			n = utf8.UTFMax
		}
		c, size := utf8.DecodeRuneInString(string(src[i : i+n]))
		if c == utf8.RuneError && size == 1 {
			dst = append(dst, src[start:i]...)
			dst = append(dst, `\ufffd`...)
			i += size
			start = i
			continue
		}
		// U+2028 is LINE SEPARATOR.
		// U+2029 is PARAGRAPH SEPARATOR.
		// They are both technically valid characters in JSON strings,
		// but don't work in JSONP, which has to be evaluated as JavaScript,
		// and can lead to security holes there. It is valid JSON to
		// escape them, so we do so unconditionally.
		// See https://en.wikipedia.org/wiki/JSON#Safety.
		if c == '\u2028' || c == '\u2029' {
			dst = append(dst, src[start:i]...)
			dst = append(dst, '\\', 'u', '2', '0', '2', hex[c&0xF])
			i += size
			start = i
			continue
		}
		i += size
	}
	dst = append(dst, src[start:]...)
	dst = append(dst, '"')
	return dst
}

var safeSet = [utf8.RuneSelf]bool{
	' ':      true,
	'!':      true,
	'"':      false,
	'#':      true,
	'$':      true,
	'%':      true,
	'&':      true,
	'\'':     true,
	'(':      true,
	')':      true,
	'*':      true,
	'+':      true,
	',':      true,
	'-':      true,
	'.':      true,
	'/':      true,
	'0':      true,
	'1':      true,
	'2':      true,
	'3':      true,
	'4':      true,
	'5':      true,
	'6':      true,
	'7':      true,
	'8':      true,
	'9':      true,
	':':      true,
	';':      true,
	'<':      true,
	'=':      true,
	'>':      true,
	'?':      true,
	'@':      true,
	'A':      true,
	'B':      true,
	'C':      true,
	'D':      true,
	'E':      true,
	'F':      true,
	'G':      true,
	'H':      true,
	'I':      true,
	'J':      true,
	'K':      true,
	'L':      true,
	'M':      true,
	'N':      true,
	'O':      true,
	'P':      true,
	'Q':      true,
	'R':      true,
	'S':      true,
	'T':      true,
	'U':      true,
	'V':      true,
	'W':      true,
	'X':      true,
	'Y':      true,
	'Z':      true,
	'[':      true,
	'\\':     false,
	']':      true,
	'^':      true,
	'_':      true,
	'`':      true,
	'a':      true,
	'b':      true,
	'c':      true,
	'd':      true,
	'e':      true,
	'f':      true,
	'g':      true,
	'h':      true,
	'i':      true,
	'j':      true,
	'k':      true,
	'l':      true,
	'm':      true,
	'n':      true,
	'o':      true,
	'p':      true,
	'q':      true,
	'r':      true,
	's':      true,
	't':      true,
	'u':      true,
	'v':      true,
	'w':      true,
	'x':      true,
	'y':      true,
	'z':      true,
	'{':      true,
	'|':      true,
	'}':      true,
	'~':      true,
	'\u007f': true,
}
