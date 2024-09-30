package benchmarks

import (
	"encoding/json"
	"io"
	"testing"

	"github.com/avpetkun/jessy-go"
	"github.com/bytedance/sonic"
	gojson "github.com/goccy/go-json"
	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
	"github.com/wI2L/jettison"
)

func BenchmarkMediumStruct(b *testing.B) {
	s := getMediumTestStruct()
	runValueBenchmarks(b, s)
}

func BenchmarkMediumSlice(b *testing.B) {
	slice := make([]*Struct, 10)
	for i := range slice {
		slice[i] = getMediumTestStruct()
	}

	runValueBenchmarks(b, slice)
}

func BenchmarkBigSlice(b *testing.B) {
	slice := make([]*Struct, 100_000)
	for i := range slice {
		slice[i] = getMediumTestStruct()
	}
	runValueBenchmarks(b, slice)
}

func BenchmarkBigMap(b *testing.B) {
	const N = 100_000

	m := make(map[string]*Struct, N)
	for range N {
		m[uuid.NewString()] = getMediumTestStruct()
	}

	runValueBenchmarks(b, m)
}

func runValueBenchmarks(b *testing.B, value any) {
	buf := make([]byte, 0, 3000000000)

	jessyStdEnc := jessy.NewEncoder(io.Discard)
	jessyFastEnc := jessy.NewEncoderWithFlags(io.Discard, jessy.EncodeFastest)

	sonicStd := sonic.ConfigStd
	sonicDef := sonic.ConfigDefault
	sonicFast := sonic.ConfigFastest
	sonicStdEnc := sonicStd.NewEncoder(io.Discard)
	sonicDefEnc := sonicDef.NewEncoder(io.Discard)
	sonicFastEnc := sonicFast.NewEncoder(io.Discard)

	jettisonFastOpts := []jettison.Option{
		jettison.UnsortedMap(),
		jettison.NoCompact(),
		jettison.NoHTMLEscaping(),
		jettison.NoStringEscaping(),
		jettison.NoUTF8Coercion(),
	}

	jsoniterDef := jsoniter.ConfigDefault
	jsoniterFast := jsoniter.ConfigFastest
	jsoniterCompat := jsoniter.ConfigCompatibleWithStandardLibrary
	jsoniterDefEnc := jsoniterDef.NewEncoder(io.Discard)
	jsoniterFastEnc := jsoniterFast.NewEncoder(io.Discard)
	jsoniterCompatEnc := jsoniterCompat.NewEncoder(io.Discard)

	gojsonEnc := gojson.NewEncoder(io.Discard)
	gojsonFastOpts := []gojson.EncodeOptionFunc{
		gojson.DisableHTMLEscape(), gojson.DisableNormalizeUTF8(), gojson.UnorderedMap(),
	}

	stdJsonEnc := json.NewEncoder(io.Discard)
	stdJsonFastEnc := json.NewEncoder(io.Discard)
	stdJsonFastEnc.SetEscapeHTML(false)

	b.ResetTimer()

	b.Run("jessy", func(b *testing.B) {
		b.Run("marshal-std", func(b *testing.B) {
			for range b.N {
				jessy.Marshal(value)
			}
		})
		b.Run("marshal-fast", func(b *testing.B) {
			for range b.N {
				jessy.MarshalFast(value)
			}
		})
		b.Run("append-std", func(b *testing.B) {
			for range b.N {
				buf, _ = jessy.Append(buf[:0], value)
			}
		})
		b.Run("append-fast", func(b *testing.B) {
			for range b.N {
				buf, _ = jessy.AppendFast(buf[:0], value)
			}
		})
		b.Run("append-fast-pretty", func(b *testing.B) {
			for range b.N {
				buf, _ = jessy.AppendPrettyFast(buf[:0], value)
			}
		})
		b.Run("encode-std", func(b *testing.B) {
			for range b.N {
				jessyStdEnc.Encode(value)
			}
		})
		b.Run("encode-fast", func(b *testing.B) {
			for range b.N {
				jessyFastEnc.Encode(value)
			}
		})
	})

	return

	b.Run("sonic", func(b *testing.B) {
		b.Run("std-marshal", func(b *testing.B) {
			for range b.N {
				sonicStd.Marshal(value)
			}
		})
		b.Run("std-encode", func(b *testing.B) {
			for range b.N {
				sonicStdEnc.Encode(value)
			}
		})
		b.Run("default-marshal", func(b *testing.B) {
			for range b.N {
				sonicDef.Marshal(value)
			}
		})
		b.Run("default-encode", func(b *testing.B) {
			for range b.N {
				sonicDefEnc.Encode(value)
			}
		})
		b.Run("fast-marshal", func(b *testing.B) {
			for range b.N {
				sonicFast.Marshal(value)
			}
		})
		b.Run("fast-encode", func(b *testing.B) {
			for range b.N {
				sonicFastEnc.Encode(value)
			}
		})
	})

	b.Run("jettison", func(b *testing.B) {
		b.Run("marshal-full", func(b *testing.B) {
			for range b.N {
				jettison.Marshal(value)
			}
		})
		b.Run("marshal-fast", func(b *testing.B) {
			for range b.N {
				jettison.MarshalOpts(value, jettisonFastOpts...)
			}
		})
		b.Run("append-full", func(b *testing.B) {
			for range b.N {
				buf, _ = jettison.Append(buf[:0], value)
			}
		})
		b.Run("append-fast", func(b *testing.B) {
			for range b.N {
				buf, _ = jettison.AppendOpts(buf[:0], value, jettisonFastOpts...)
			}
		})
	})

	b.Run("jsoniter", func(b *testing.B) {
		b.Run("marshal-default", func(b *testing.B) {
			for range b.N {
				jsoniterDef.Marshal(value)
			}
		})
		b.Run("marshal-fast", func(b *testing.B) {
			for range b.N {
				jsoniterFast.Marshal(value)
			}
		})
		b.Run("marshal-compat", func(b *testing.B) {
			for range b.N {
				jsoniterCompat.Marshal(value)
			}
		})

		b.Run("encode-default", func(b *testing.B) {
			for range b.N {
				jsoniterDefEnc.Encode(value)
			}
		})
		b.Run("encode-fast", func(b *testing.B) {
			for range b.N {
				jsoniterFastEnc.Encode(value)
			}
		})
		b.Run("encode-compat", func(b *testing.B) {
			for range b.N {
				jsoniterCompatEnc.Encode(value)
			}
		})

		b.Run("borrow-default", func(b *testing.B) {
			for range b.N {
				s := jsoniterDef.BorrowStream(io.Discard)
				s.WriteVal(value)
				jsoniterDef.ReturnStream(s)
			}
		})
		b.Run("borrow-fast", func(b *testing.B) {
			for range b.N {
				s := jsoniterFast.BorrowStream(io.Discard)
				s.WriteVal(value)
				jsoniterFast.ReturnStream(s)
			}
		})
		b.Run("borrow-compat", func(b *testing.B) {
			for range b.N {
				s := jsoniterCompat.BorrowStream(io.Discard)
				s.WriteVal(value)
				jsoniterCompat.ReturnStream(s)
			}
		})
	})

	b.Run("gojson", func(b *testing.B) {
		b.Run("marshal-full", func(b *testing.B) {
			for range b.N {
				gojson.Marshal(value)
			}
		})
		b.Run("marshal-fast", func(b *testing.B) {
			for range b.N {
				gojson.MarshalWithOption(value, gojsonFastOpts...)
			}
		})
		b.Run("encoder-full", func(b *testing.B) {
			for range b.N {
				gojsonEnc.Encode(value)
			}
		})
		b.Run("encoder-fast", func(b *testing.B) {
			for range b.N {
				gojsonEnc.EncodeWithOption(value, gojsonFastOpts...)
			}
		})
	})

	b.Run("encoding-json", func(b *testing.B) {
		b.Run("marshal", func(b *testing.B) {
			for range b.N {
				json.Marshal(value)
			}
		})
		b.Run("encoder-full", func(b *testing.B) {
			for range b.N {
				stdJsonEnc.Encode(value)
			}
		})
		b.Run("encoder-fast", func(b *testing.B) {
			b.ResetTimer()
			for range b.N {
				stdJsonFastEnc.Encode(value)
			}
		})
	})
}
