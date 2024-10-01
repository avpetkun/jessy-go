package zstr

import (
	"math"
	"math/rand"
	"strconv"
	"testing"

	"github.com/avpetkun/jessy-go/require"
)

func TestFormatUint8(t *testing.T) {
	buf := make([]byte, 3)
	for v := range uint8(math.MaxUint8) {
		buf = AppendUint8(buf[:0], v)
		require.Equal(t, strconv.FormatUint(uint64(v), 10), string(buf))
	}
}

func TestFormatInt8(t *testing.T) {
	buf := make([]byte, 3)
	for v := math.MinInt8; v <= math.MaxInt8; v++ {
		buf = AppendInt8(buf[:0], int8(v))
		require.Equal(t, strconv.FormatInt(int64(v), 10), string(buf))
	}
}

func TestFormatUint64(t *testing.T) {
	buf := make([]byte, 0, 3)

	buf = AppendUint64(buf[:0], 0)
	require.Equal(t, "0", string(buf))

	buf = AppendUint64(buf[:0], math.MaxUint64)
	require.Equal(t, strconv.FormatUint(math.MaxUint64, 10), string(buf))

	for range 100 {
		v := rand.Uint64()
		buf = AppendUint64(buf[:0], v)
		require.Equal(t, strconv.FormatUint(v, 10), string(buf))
	}
}

func TestFormatInt64(t *testing.T) {
	buf := make([]byte, 0, 3)

	buf = AppendInt64(buf[:0], 0)
	require.Equal(t, "0", string(buf))

	buf = AppendInt64(buf[:0], math.MaxInt64)
	require.Equal(t, strconv.FormatInt(math.MaxInt64, 10), string(buf))

	buf = AppendInt64(buf[:0], math.MinInt64)
	require.Equal(t, strconv.FormatInt(math.MinInt64, 10), string(buf))

	for range 100 {
		v := rand.Int63()
		buf = AppendInt64(buf[:0], v)
		require.Equal(t, strconv.FormatInt(v, 10), string(buf))
	}
}

func BenchmarkFormatUint8(b *testing.B) {
	buf := make([]byte, 3)
	b.Run("strconv", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			buf = strconv.AppendUint(buf[:0], 123, 10)
		}
	})
	b.Run("custom", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			buf = AppendUint8(buf[:0], 123)
		}
	})
}

func BenchmarkFormatInt8(b *testing.B) {
	buf := make([]byte, 4)
	b.Run("strconv", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			buf = strconv.AppendInt(buf[:0], -123, 10)
		}
	})
	b.Run("custom", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			buf = AppendInt8(buf[:0], -123)
		}
	})
}

func BenchmarkFormatUint64(b *testing.B) {
	benchs := []struct {
		Name  string
		Value uint64
	}{
		{Name: "12345", Value: 12345},
		{Name: "12345678910", Value: 12345678910},
		{Name: "18446744073709551615", Value: math.MaxUint64},
	}
	buf := make([]byte, 20)
	for _, bench := range benchs {
		b.Run(bench.Name, func(b *testing.B) {
			b.Run("strconv", func(b *testing.B) {
				b.ResetTimer()
				for range b.N {
					buf = strconv.AppendUint(buf[:0], bench.Value, 10)
				}
			})
			b.Run("custom", func(b *testing.B) {
				b.ResetTimer()
				for range b.N {
					buf = AppendUint64(buf[:0], bench.Value)
				}
			})
		})
	}
}

func BenchmarkFormatInt64(b *testing.B) {
	benchs := []struct {
		Name  string
		Value int64
	}{
		{Name: "12345", Value: 12345},
		{Name: "-12345", Value: -12345},
		{Name: "12345678910", Value: 12345678910},
		{Name: "-12345678910", Value: -12345678910},
		{Name: "9223372036854775807", Value: math.MaxInt64},
		{Name: "-9223372036854775808", Value: math.MinInt64},
	}
	buf := make([]byte, 20)
	for _, bench := range benchs {
		b.Run(bench.Name, func(b *testing.B) {
			b.Run("strconv", func(b *testing.B) {
				b.ResetTimer()
				for range b.N {
					buf = strconv.AppendInt(buf[:0], bench.Value, 10)
				}
			})
			b.Run("custom", func(b *testing.B) {
				b.ResetTimer()
				for range b.N {
					buf = AppendInt64(buf[:0], bench.Value)
				}
			})
		})
	}
}
