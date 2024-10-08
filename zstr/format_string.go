package zstr

import (
	"encoding/base64"
	"unicode/utf8"
)

func AppendBase64String(dst, data []byte) []byte {
	size := base64.StdEncoding.EncodedLen(len(data)) + 2

	i := len(dst)
	dst = growCap(dst, size)[:i+size]

	dst[i] = '"'
	base64.StdEncoding.Encode(dst[i+1:], data)
	dst[len(dst)-1] = '"'

	return dst
}

const toHex = "0123456789abcdef"

func AppendHexString(dst, data []byte) []byte {
	size := len(data)*2 + 4

	i := len(dst)
	dst = growCap(dst, size)[:i+size]

	dst[i] = '"'
	dst[i+1] = '0'
	dst[i+2] = 'x'
	i += 3

	for _, v := range data {
		dst[i] = toHex[v>>4]
		dst[i+1] = toHex[v&0x0f]
		i += 2
	}
	dst[i] = '"'

	return dst
}

func AppendHex(dst, data []byte) []byte {
	size := len(data)*2 + 2

	i := len(dst)
	dst = growCap(dst, size)[:i+size]

	dst[i] = '0'
	dst[i+1] = 'x'
	i += 2

	for _, v := range data {
		dst[i] = toHex[v>>4]
		dst[i+1] = toHex[v&0x0f]
		i += 2
	}

	return dst
}

func AppendHTMLEscape(dst, src []byte) []byte {
	// The characters can only appear in string literals,
	// so just scan the string one byte at a time.
	start := 0
	for i, c := range src {
		if c == '<' || c == '>' || c == '&' {
			dst = append(dst, src[start:i]...)
			dst = append(dst, '\\', 'u', '0', '0', toHex[c>>4], toHex[c&0xF])
			start = i + 1
		}
		// Convert U+2028 and U+2029 (E2 80 A8 and E2 80 A9).
		if c == 0xE2 && i+2 < len(src) && src[i+1] == 0x80 && src[i+2]&^1 == 0xA8 {
			dst = append(dst, src[start:i]...)
			dst = append(dst, '\\', 'u', '2', '0', '2', toHex[src[i+2]&0xF])
			start = i + len("\u2029")
		}
	}
	return append(dst, src[start:]...)
}

// from encoding/json.AppendQuotedString
func AppendQuotedString(dst, src []byte, escapeHtml bool) []byte {
	dst = growCap(dst, len(src)+2)
	dst = append(dst, '"')
	start := 0
	srcLen := len(src)
	for i := 0; i < srcLen; i++ {
		if b := src[i]; b < utf8.RuneSelf {
			if htmlSafeSet[b] || (!escapeHtml && safeSet[b]) {
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
				dst = append(dst, '\\', 'u', '0', '0', toHex[b>>4], toHex[b&0xF])
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
			dst = append(dst, '\\', 'u', '2', '0', '2', toHex[c&0xF])
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

// from encoding/json.Compact
func AppendCompactJSON(dst, src []byte, escapeHTML bool) []byte {
	var inString bool
	var skipNext bool
	start := 0

	for i, c := range src {
		if escapeHTML && (c == '<' || c == '>' || c == '&') {
			if start < i {
				dst = append(dst, src[start:i]...)
			}
			dst = append(dst, '\\', 'u', '0', '0', toHex[c>>4], toHex[c&0xF])
			start = i + 1
			continue
		}
		// Convert U+2028 and U+2029 (E2 80 A8 and E2 80 A9).
		if escapeHTML && c == 0xE2 && i+2 < len(src) && src[i+1] == 0x80 && src[i+2]&^1 == 0xA8 {
			if start < i {
				dst = append(dst, src[start:i]...)
			}
			dst = append(dst, '\\', 'u', '2', '0', '2', toHex[src[i+2]&0xF])
			start = i + 3
			continue
		}
		if !inString {
			if c == '"' {
				inString = true
			} else if isSpace(c) {
				if start < i {
					dst = append(dst, src[start:i]...)
				}
				start = i + 1
			}
			continue
		}
		// skip already escaped char
		if skipNext {
			skipNext = false
			continue
		}
		// char is escaped
		if c == '\\' {
			skipNext = true
			continue
		}
		if c == '"' {
			inString = false
		}
	}
	if start < len(src) {
		dst = append(dst, src[start:]...)
	}
	return dst
}

func AppendIndent(dst, src []byte, prefix, indent string) []byte {
	deep := 0
	start := 0
	lastIndentLen := 0
	inString := false

	dst = growCap(dst, len(src)*2)
	dst = append(dst, prefix...)

	for i := range src {
		if inString {
			inString = src[i] != '"' || src[i-1] == '\\'
			continue
		}
		switch src[i] {
		case '"':
			inString = true
		case '{', '[':
			deep++
			dst = append(dst, src[start:i+1]...)
			start = i + 1
			dstLen := len(dst)
			dst = appendNewline(dst, prefix, indent, deep)
			lastIndentLen = len(dst) - dstLen
		case '}', ']':
			dst = append(dst, src[start:i]...)
			if lastIndentLen != 0 && i-start == 0 {
				dst = dst[:len(dst)-lastIndentLen]
				deep--
			} else {
				deep--
				dst = appendNewline(dst, prefix, indent, deep)
			}
			dst = append(dst, src[i])
			start = i + 1
			lastIndentLen = 0
		case ',':
			dst = append(dst, src[start:i+1]...)
			dst = appendNewline(dst, prefix, indent, deep)
			start = i + 1
		case ':':
			dst = append(dst, src[start:i+1]...)
			dst = append(dst, ' ')
			start = i + 1
		}
	}

	dst = append(dst, src[start:]...)

	return dst
}

func appendNewline(dst []byte, prefix, indent string, deep int) []byte {
	dst = append(dst, '\n')
	dst = append(dst, prefix...)
	for range deep {
		dst = append(dst, indent...)
	}
	return dst
}

const spaceMask = (1 << ' ') | (1 << '\t') | (1 << '\r') | (1 << '\n')

func isSpace(c byte) bool {
	return spaceMask&(1<<c) != 0
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

var htmlSafeSet = [utf8.RuneSelf]bool{
	' ':      true,
	'!':      true,
	'"':      false,
	'#':      true,
	'$':      true,
	'%':      true,
	'&':      false,
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
	'<':      false,
	'=':      true,
	'>':      false,
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
