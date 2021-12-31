package vsohash

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"time"

	"testing"
)

func BenchmarkVSOHash(b *testing.B) {
	// Need quite a bit of data to give our hashes a chance to shine.
	const size = 2 * 1024 * 1024 * 1024

	data := make([]byte, size)
	for i := uint64(0); i < size; i += 8 {
		binary.LittleEndian.PutUint64(data[i:], i)
	}
	b.ResetTimer()

	b.Run("SHA256", func(b *testing.B) {
		start := time.Now()
		for i := 0; i < b.N; i++ {
			sha256.Sum256(data)
		}
		b.ReportMetric(float64(size*b.N)/(1024*1024*time.Since(start).Seconds()), "MB/s")
	})
	for _, parallelism := range []int{1, 2, 4, 8, 16, 24} {
		b.Run(fmt.Sprintf("VSOParallel%d", parallelism), func(b *testing.B) {
			start := time.Now()
			for i := 0; i < b.N; i++ {
				h := NewParallel(parallelism)
				h.Write(data)
				h.Sum(nil)
			}
			b.ReportMetric(float64(size*b.N)/(1024*1024*time.Since(start).Seconds()), "MB/s")
		})
	}
}
