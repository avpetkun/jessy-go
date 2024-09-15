package zgo

import (
	"fmt"
	"runtime"
	"slices"
	"testing"
)

func TestSliceGrow(t *testing.T) {
	sprint := func(s []byte) {
		print(s, " ")
		fmt.Println(s)
	}

	if true {
		s := make([]byte, 10, 20)
		s[0], s[8], s[9] = 1, 8, 9
		sprint(s)
		s = slices.Grow(s, 5)
		sprint(s)
		s = slices.Grow(s, 15)
		sprint(s)

		println()

		s = make([]byte, 10, 20)
		s[0], s[8], s[9] = 1, 8, 9
		sprint(s)
		s = Grow(s, 5)
		sprint(s)
		s = Grow(s, 15)
		sprint(s)

		println()

		s = make([]byte, 10, 20)
		s[0], s[8], s[9] = 1, 8, 9
		sprint(s)
		s = GrowBytes(s, 5)
		sprint(s)
		s = GrowBytes(s, 15)
		sprint(s)
	}

	println()

	if true {
		s := make([]byte, 10, 20)
		s[0], s[8], s[9] = 1, 8, 9
		sprint(s)
		s = stdGrowLen(s, 5)
		sprint(s)
		s = stdGrowLen(s, 15)
		sprint(s)

		println()

		s = make([]byte, 10, 20)
		s[0], s[8], s[9] = 1, 8, 9
		sprint(s)
		s = GrowLen(s, 5)
		sprint(s)
		s = GrowLen(s, 15)
		sprint(s)

		println()

		s = make([]byte, 10, 20)
		s[0], s[8], s[9] = 1, 8, 9
		sprint(s)
		s = GrowBytesLen(s, 5)
		sprint(s)
		s = GrowBytesLen(s, 15)
		sprint(s)
	}
}

func BenchmarkSliceGrow(b *testing.B) {
	b.Run("slices.Grow", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			s := make([]byte, 65_000, 66_000)
			s[0], s[8], s[9] = 1, 8, 9
			s = slices.Grow(s, 5)
			s = slices.Grow(s, 40_000)
			runtime.KeepAlive(s)
		}
	})
	b.Run("Grow", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			s := make([]byte, 65_000, 66_000)
			s[0], s[8], s[9] = 1, 8, 9
			s = Grow(s, 5)
			s = Grow(s, 40_000)
			runtime.KeepAlive(s)
		}
	})
	b.Run("GrowBytes", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			s := make([]byte, 65_000, 66_000)
			s[0], s[8], s[9] = 1, 8, 9
			s = GrowBytes(s, 5)
			s = GrowBytes(s, 40_000)
			runtime.KeepAlive(s)
		}
	})
}

func stdGrowLen[S ~[]E, E any](s S, n int) S {
	if d := len(s) + n - cap(s); d > 0 {
		return append(s[:cap(s)], make([]E, d)...)
	}
	return s[:len(s)+n]
}

// goos: darwin
// goarch: arm64
// cpu: Apple M1 Pro
// BenchmarkSliceGrow/slices.Grow    		10893 ns/op		204800 B/op		2 allocs/op
// BenchmarkSliceGrow/Grow           		 9406 ns/op		204801 B/op		2 allocs/op
// BenchmarkSliceGrow/GrowBytes      		 9385 ns/op		204801 B/op		2 allocs/op
// BenchmarkSliceGrowLen/stdGrowLen  		11011 ns/op		204800 B/op		2 allocs/op
// BenchmarkSliceGrowLen/GrowLen     		 9351 ns/op		204801 B/op		2 allocs/op
// BenchmarkSliceGrowLen/GrowBytesLen		 9272 ns/op		204801 B/op		2 allocs/op

func BenchmarkSliceGrowLen(b *testing.B) {
	b.Run("stdGrowLen", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			s := make([]byte, 65_000, 66_000)
			s[0], s[8], s[9] = 1, 8, 9
			s = stdGrowLen(s, 5)
			s = stdGrowLen(s, 40_000)
			runtime.KeepAlive(s)
		}
	})
	b.Run("GrowLen", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			s := make([]byte, 65_000, 66_000)
			s[0], s[8], s[9] = 1, 8, 9
			s = GrowLen(s, 5)
			s = GrowLen(s, 40_000)
			runtime.KeepAlive(s)
		}
	})
	b.Run("GrowBytesLen", func(b *testing.B) {
		b.ResetTimer()
		for range b.N {
			s := make([]byte, 65_000, 66_000)
			s[0], s[8], s[9] = 1, 8, 9
			s = GrowBytesLen(s, 5)
			s = GrowBytesLen(s, 40_000)
			runtime.KeepAlive(s)
		}
	})
}
