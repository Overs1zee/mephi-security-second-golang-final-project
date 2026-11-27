package main

import (
	"runtime"
	"testing"
	"time"
)

func TestGenerateRandomElements(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		size int
		want int
	}{
		{name: "zero size", size: 0, want: 0},
		{name: "negative size", size: -5, want: 0},
		{name: "single element", size: 1, want: 1},
		{name: "hundred elements", size: 100, want: 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data := generateRandomElements(tt.size)

			if len(data) != tt.want {
				t.Fatalf("generateRandomElements(%d) length = %d, want %d", tt.size, len(data), tt.want)
			}
			for i, value := range data {
				if value <= 0 {
					t.Errorf("generateRandomElements(%d)[%d] = %d, want > 0", tt.size, i, value)
				}
			}
		})
	}
}

func TestMaximum(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		data []int
		want int
	}{
		{name: "nil slice", data: nil, want: 0},
		{name: "empty slice", data: []int{}, want: 0},
		{name: "single element", data: []int{42}, want: 42},
		{name: "all elements are equal", data: []int{5, 5, 5, 5}, want: 5},
		{name: "max at the start", data: []int{9, 1, 2, 3}, want: 9},
		{name: "max in the middle", data: []int{1, 2, 9, 3, 4}, want: 9},
		{name: "max at the end", data: []int{1, 2, 3, 9}, want: 9},
		{name: "negative elements", data: []int{-5, -3, -9}, want: -3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := maximum(tt.data); got != tt.want {
				t.Errorf("maximum(%v) = %d, want %d", tt.data, got, tt.want)
			}
		})
	}
}

// withTailMax returns a slice of n ones whose last element is 2,
// so the maximum sits in the tail that the last chunk must not lose.
func withTailMax(n int) []int {
	data := make([]int, n)
	for i := range data {
		data[i] = 1
	}
	data[n-1] = 2

	return data
}

func TestMaxChunks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		data []int
		want int
	}{
		{name: "nil slice", data: nil, want: 0},
		{name: "empty slice", data: []int{}, want: 0},
		{name: "shorter than CHUNKS", data: []int{7, 3, 9}, want: 9},
		{name: "shorter than CHUNKS, max in tail", data: withTailMax(3), want: 2},
		{name: "length equal to CHUNKS, max in tail", data: withTailMax(CHUNKS), want: 2},
		{name: "length 10, max in tail", data: withTailMax(10), want: 2},
		{name: "length 17, max in tail", data: withTailMax(17), want: 2},
		{name: "length 1003, max in tail", data: withTailMax(1003), want: 2},
		// An all-negative slice shorter than CHUNKS is out of the task domain
		// (elements are > 0): empty chunks yield 0, which wins over negatives.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := maxChunks(tt.data); got != tt.want {
				t.Errorf("maxChunks(%v) = %d, want %d", tt.data, got, tt.want)
			}
		})
	}

	t.Run("matches maximum on large random slice", func(t *testing.T) {
		t.Parallel()

		data := generateRandomElements(1_000_000)

		sequential := maximum(data)
		parallel := maxChunks(data)

		if parallel != sequential {
			t.Errorf("maxChunks() = %d, want %d (maximum())", parallel, sequential)
		}
	})
}

// TestMaxChunks_GoroutinesFinished checks that maxChunks does not leave
// goroutines running after it returns.
func TestMaxChunks_GoroutinesFinished(t *testing.T) {
	t.Parallel()

	data := generateRandomElements(100_000)

	before := runtime.NumGoroutine()
	maxChunks(data)

	// A goroutine may still be tearing down right after wg.Done,
	// so give it a moment before failing.
	deadline := time.Now().Add(time.Second)
	for runtime.NumGoroutine() > before {
		if time.Now().After(deadline) {
			t.Fatalf("goroutines still running after maxChunks: %d, want <= %d", runtime.NumGoroutine(), before)
		}
		time.Sleep(time.Millisecond)
	}
}

// FuzzMaxChunks checks that the chunked search agrees with the sequential
// one on arbitrary input. Fuzzing supports only []byte, so each byte
// becomes one element.
func FuzzMaxChunks(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{1})
	f.Add([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9})
	f.Add([]byte{1, 1, 2})

	f.Fuzz(func(t *testing.T, raw []byte) {
		data := make([]int, len(raw))
		for i, value := range raw {
			data[i] = int(value)
		}

		sequential := maximum(data)
		parallel := maxChunks(data)

		if parallel != sequential {
			t.Errorf("maxChunks(%v) = %d, want %d (maximum())", data, parallel, sequential)
		}
	})
}
