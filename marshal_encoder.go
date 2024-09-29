package jessy

import (
	"io"
	"sync"

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
}

var encoderEndline = []byte{'\n'}

func (e *Encoder) Encode(value any) (err error) {
	buf := encodeBufferPool.Get().(*encodeBuffer)

	buf.marshalBuf, err = encodeAny(buf.marshalBuf, value, e.flags)
	if err == nil {
		if len(e.indentPrefix) == 0 && len(e.indentValue) == 0 {
			_, err = e.Write(buf.marshalBuf)
		} else {
			buf.indentBuf = zstr.AppendIndent(buf.indentBuf, buf.marshalBuf, e.indentPrefix, e.indentValue)
			_, err = e.Write(buf.indentBuf)
		}
	}
	if err == nil {
		_, err = e.Write(encoderEndline)
	}

	buf.marshalBuf = buf.marshalBuf[:0]
	buf.indentBuf = buf.indentBuf[:0]
	encodeBufferPool.Put(buf)

	return
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

var encodeBufferPool = sync.Pool{New: func() any { return new(encodeBuffer) }}

type encodeBuffer struct {
	marshalBuf []byte
	indentBuf  []byte
}
