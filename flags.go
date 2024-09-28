package jessy

// possible values: EscapeHTML, OmitEmpty, NeedQuotes
type Flags uint32

func (flags Flags) Has(flag Flags) bool {
	return flags&flag != 0
}

func (flags Flags) Exclude(exclude Flags) Flags {
	return flags &^ exclude
}

// encoder flags
const (
	// start options
	SortMapKeys Flags = 1 << iota
	EscapeHTML
	ValidateString
	ValidateTextMarshaler
	CompactMarshaler
	PrettySpaces

	// while encoding
	OmitEmpty
	NeedQuotes

	encodeFlagsLen

	// configs
	EncodeFastest  = 0
	EncodeStandard = SortMapKeys | EscapeHTML | ValidateString | ValidateTextMarshaler | CompactMarshaler
)
