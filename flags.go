package jessy

// possible values: EscapeHTML, OmitEmpty, NeedQuotes
type Flags uint32

func (flags Flags) Has(flag Flags) bool {
	return flags&flag != 0
}

func (flags Flags) excludes(flagsToExclude ...Flags) Flags {
	for _, f := range flagsToExclude {
		flags &^= f
	}
	return flags
}

// encoder flags
const (
	// start options
	SortMapKeys Flags = 1 << iota
	EscapeHTML
	ValidateString
	ValidateTextMarshaller

	// while encoding
	OmitEmpty
	NeedQuotes

	encodeFlagsLen

	// configs
	EncodeFastest  = 0
	EncodeStandard = SortMapKeys | EscapeHTML | ValidateString | ValidateTextMarshaller
)
