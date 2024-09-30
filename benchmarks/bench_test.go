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
			buf := make([]byte, 0, 3000000000)
			b.ResetTimer()
			for range b.N {
				buf, _ = jessy.Append(buf[:0], value)
			}
		})
		b.Run("append-fast", func(b *testing.B) {
			buf := make([]byte, 0, 3000000000)
			b.ResetTimer()
			for range b.N {
				buf, _ = jessy.AppendFast(buf[:0], value)
			}
		})
		b.Run("append-fast-pretty", func(b *testing.B) {
			buf := make([]byte, 0, 3000000000)
			b.ResetTimer()
			for range b.N {
				buf, _ = jessy.AppendPrettyFast(buf[:0], value)
			}
		})
		b.Run("encode-std", func(b *testing.B) {
			e := jessy.NewEncoder(io.Discard)
			e.Grow(3000000000)
			b.ResetTimer()
			for range b.N {
				e.Encode(value)
			}
		})
		b.Run("encode-fast", func(b *testing.B) {
			e := jessy.NewEncoderWithFlags(io.Discard, jessy.EncodeFastest)
			e.Grow(3000000000)
			b.ResetTimer()
			for range b.N {
				e.Encode(value)
			}
		})
	})

	b.Run("sonic", func(b *testing.B) {
		std := sonic.ConfigStd
		def := sonic.ConfigDefault
		fast := sonic.ConfigFastest
		stdEnc := std.NewEncoder(io.Discard)
		defEnc := def.NewEncoder(io.Discard)
		fastEnc := fast.NewEncoder(io.Discard)
		b.ResetTimer()
		b.Run("std-marshal", func(b *testing.B) {
			for range b.N {
				std.Marshal(value)
			}
		})
		b.Run("std-encode", func(b *testing.B) {
			for range b.N {
				stdEnc.Encode(value)
			}
		})
		b.Run("default-marshal", func(b *testing.B) {
			for range b.N {
				def.Marshal(value)
			}
		})
		b.Run("default-encode", func(b *testing.B) {
			for range b.N {
				defEnc.Encode(value)
			}
		})
		b.Run("fast-marshal", func(b *testing.B) {
			for range b.N {
				fast.Marshal(value)
			}
		})
		b.Run("fast-encode", func(b *testing.B) {
			for range b.N {
				fastEnc.Encode(value)
			}
		})
	})

	b.Run("jettison", func(b *testing.B) {
		fastOpts := []jettison.Option{
			jettison.UnsortedMap(),
			jettison.NoCompact(),
			jettison.NoHTMLEscaping(),
			jettison.NoStringEscaping(),
			jettison.NoUTF8Coercion(),
		}
		b.Run("marshal-full", func(b *testing.B) {
			for range b.N {
				jettison.Marshal(value)
			}
		})
		b.Run("marshal-fast", func(b *testing.B) {
			for range b.N {
				jettison.MarshalOpts(value, fastOpts...)
			}
		})
		b.Run("append-full", func(b *testing.B) {
			buf := make([]byte, 0, 3000000000)
			b.ResetTimer()
			for range b.N {
				buf, _ = jettison.Append(buf[:0], value)
			}
		})
		b.Run("append-fast", func(b *testing.B) {
			buf := make([]byte, 0, 3000000000)
			b.ResetTimer()
			for range b.N {
				buf, _ = jettison.AppendOpts(buf[:0], value, fastOpts...)
			}
		})
	})

	b.Run("jsoniter", func(b *testing.B) {
		def := jsoniter.ConfigDefault
		fast := jsoniter.ConfigFastest
		compat := jsoniter.ConfigCompatibleWithStandardLibrary
		defEnc := def.NewEncoder(io.Discard)
		fastEnc := fast.NewEncoder(io.Discard)
		compatEnc := compat.NewEncoder(io.Discard)
		b.ResetTimer()

		b.Run("marshal-default", func(b *testing.B) {
			for range b.N {
				def.Marshal(value)
			}
		})
		b.Run("marshal-fast", func(b *testing.B) {
			for range b.N {
				fast.Marshal(value)
			}
		})
		b.Run("marshal-compat", func(b *testing.B) {
			for range b.N {
				compat.Marshal(value)
			}
		})

		b.Run("encode-default", func(b *testing.B) {
			for range b.N {
				defEnc.Encode(value)
			}
		})
		b.Run("encode-fast", func(b *testing.B) {
			for range b.N {
				fastEnc.Encode(value)
			}
		})
		b.Run("encode-compat", func(b *testing.B) {
			for range b.N {
				compatEnc.Encode(value)
			}
		})

		b.Run("borrow-default", func(b *testing.B) {
			for range b.N {
				s := def.BorrowStream(io.Discard)
				s.WriteVal(value)
				def.ReturnStream(s)
			}
		})
		b.Run("borrow-fast", func(b *testing.B) {
			for range b.N {
				s := fast.BorrowStream(io.Discard)
				s.WriteVal(value)
				fast.ReturnStream(s)
			}
		})
		b.Run("borrow-compat", func(b *testing.B) {
			for range b.N {
				s := compat.BorrowStream(io.Discard)
				s.WriteVal(value)
				compat.ReturnStream(s)
			}
		})
	})

	b.Run("gojson", func(b *testing.B) {
		fastOpts := []gojson.EncodeOptionFunc{
			gojson.DisableHTMLEscape(), gojson.DisableNormalizeUTF8(), gojson.UnorderedMap(),
		}
		b.Run("marshal-full", func(b *testing.B) {
			for range b.N {
				gojson.Marshal(value)
			}
		})
		b.Run("marshal-fast", func(b *testing.B) {
			for range b.N {
				gojson.MarshalWithOption(value, fastOpts...)
			}
		})
		b.Run("encoder-full", func(b *testing.B) {
			enc := gojson.NewEncoder(io.Discard)
			b.ResetTimer()
			for range b.N {
				enc.Encode(value)
			}
		})
		b.Run("encoder-fast", func(b *testing.B) {
			enc := gojson.NewEncoder(io.Discard)
			b.ResetTimer()
			for range b.N {
				enc.EncodeWithOption(value, fastOpts...)
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
			enc := json.NewEncoder(io.Discard)
			b.ResetTimer()
			for range b.N {
				enc.Encode(value)
			}
		})
		b.Run("encoder-fast", func(b *testing.B) {
			enc := json.NewEncoder(io.Discard)
			enc.SetEscapeHTML(false)
			b.ResetTimer()
			for range b.N {
				enc.Encode(value)
			}
		})
	})
}
