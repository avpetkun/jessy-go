package jessy

import (
	"io"
	"slices"

	"github.com/avpetkun/jessy-go/zstr"
)

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{Writer: w, flags: EncodeStandard}
}

func NewEncoderWithFlags(w io.Writer, flags Flags) *Encoder {
	return &Encoder{Writer: w, flags: flags}
}

type Encoder struct {
	io.Writer

	flags Flags

	indentPrefix string
	indentValue  string

	marshalBuf []byte
	indentBuf  []byte
}

var encoderEndline = []byte{'\n'}

func (e *Encoder) Encode(value any) (err error) {
	e.marshalBuf, err = encodeAny(e.marshalBuf[:0], value, e.flags)
	if err == nil {
		if len(e.indentPrefix) == 0 && len(e.indentValue) == 0 {
			_, err = e.Write(e.marshalBuf)
		} else {
			e.indentBuf = slices.Grow(e.indentBuf[:0], len(e.marshalBuf)*2)
			e.indentBuf = zstr.AppendIndent(e.indentBuf, e.marshalBuf, e.indentPrefix, e.indentValue)
			_, err = e.Write(e.indentBuf)
		}
	}
	if err == nil {
		_, err = e.Write(encoderEndline)
	}
	return
}

func (e *Encoder) Grow(size int) {
	e.marshalBuf = slices.Grow(e.marshalBuf, size)
}

func (e *Encoder) GrowIndent(size int) {
	e.indentBuf = slices.Grow(e.indentBuf, size)
}

func (e *Encoder) SetEscapeHTML(on bool) {
	if on {
		e.flags |= EscapeHTML
	} else {
		e.flags &^= EscapeHTML
	}
}

func (e *Encoder) SetFlags(flags Flags) {
	e.flags = flags
}

func (e *Encoder) SetStandardFlags() {
	e.flags = EncodeStandard
}

func (e *Encoder) SetFastestFlags() {
	e.flags = EncodeFastest
}

func (e *Encoder) SetPrettyFlags(on bool) {
	if on {
		e.flags |= PrettySpaces
	} else {
		e.flags &^= PrettySpaces
	}
}

func (e *Encoder) SetIndent(prefix, indent string) {
	e.indentPrefix = prefix
	e.indentValue = indent
}
