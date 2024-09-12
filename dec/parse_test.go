package dec

import (
	"strconv"
	"testing"
)

func TestParseUint64(t *testing.T) {
	cases := []struct {
		V uint64
		S string
	}{
		{12345678910, "12345678910"},
		{0, "0"}, // min uint64
		{18446744073709551615, "18446744073709551615"}, // max uint64
	}
	for _, c := range cases {
		v, err := ParseUint64([]byte(c.S))
		if err != nil {
			t.Error(err)
			continue
		}
		if v != c.V {
			t.Errorf("expected %d actual %d", c.V, v)
		}
	}
}

func TestParseInt64(t *testing.T) {
	cases := []struct {
		V int64
		S string
	}{
		{12345678910, "12345678910"},
		{-12345678910, "-12345678910"},
		{-9223372036854775808, "-9223372036854775808"}, // min int64
		{9223372036854775807, "9223372036854775807"},   // max int64
	}
	for _, c := range cases {
		v, err := ParseInt64([]byte(c.S))
		if err != nil {
			t.Error(err)
			continue
		}
		if v != c.V {
			t.Errorf("expected %d actual %d", c.V, v)
		}
	}
}

func BenchmarkParseUint64(b *testing.B) {
	b.Run("strconv", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			_, _ = strconv.ParseUint("123456789", 10, 64)
		}
	})
	b.Run("custom", func(b *testing.B) {
		bs := []byte("123456789")
		b.ResetTimer()
		for range b.N {
			_, _ = ParseUint64(bs)
		}
	})
}

func BenchmarkParseInt64(b *testing.B) {
	b.Run("strconv", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			_, _ = strconv.ParseInt("-123456789", 10, 64)
		}
	})
	b.Run("custom", func(b *testing.B) {
		bs := []byte("-123456789")
		b.ResetTimer()
		for range b.N {
			_, _ = ParseInt64(bs)
		}
	})
}
