package filters

import (
	"crypto/sha256"
	"encoding/binary"
	"math"
)

// bloom_filter.go builds the ignore Bloom filter discussed in algorithm_design.
// Implement Add/Contains helpers tuned for CLI patterns. Benchmark with `go test`.

// BloomFilter implements a probabilistic set with configurable size/hash count.
type BloomFilter struct {
	bits []uint64
	m    uint64
	k    uint64
}

// NewBloomFilter constructs a Bloom filter sized for the expected cardinality and
// false positive rate. When invalid values are supplied sensible defaults are
// used.
func NewBloomFilter(expectedItems int, falsePositiveRate float64) *BloomFilter {
	if expectedItems <= 0 {
		expectedItems = 1024
	}
	if falsePositiveRate <= 0 || falsePositiveRate >= 1 {
		falsePositiveRate = 0.01
	}

	m := optimalBitCount(expectedItems, falsePositiveRate)
	k := optimalHashFunctions(m, expectedItems)

	slots := int((m + 63) / 64)
	if slots <= 0 {
		slots = 1
	}

	return &BloomFilter{
		bits: make([]uint64, slots),
		m:    m,
		k:    k,
	}
}

// Add inserts a token into the filter.
func (bf *BloomFilter) Add(token string) {
	if bf == nil || bf.m == 0 {
		return
	}
	h1, h2 := hashPair(token)
	bf.lockBits(h1, h2)
}

// Contains reports whether the token might be present in the filter.
func (bf *BloomFilter) Contains(token string) bool {
	if bf == nil || bf.m == 0 {
		return false
	}
	h1, h2 := hashPair(token)
	return bf.checkBits(h1, h2)
}

func (bf *BloomFilter) lockBits(h1, h2 uint64) {
	if bf.m == 0 {
		return
	}
	for i := uint64(0); i < bf.k; i++ {
		combined := (h1 + i*h2) % bf.m
		index := combined / 64
		mask := uint64(1) << (combined % 64)
		bf.bits[index] |= mask
	}
}

func (bf *BloomFilter) checkBits(h1, h2 uint64) bool {
	if bf.m == 0 {
		return false
	}
	for i := uint64(0); i < bf.k; i++ {
		combined := (h1 + i*h2) % bf.m
		index := combined / 64
		mask := uint64(1) << (combined % 64)
		if bf.bits[index]&mask == 0 {
			return false
		}
	}
	return true
}

func hashPair(token string) (uint64, uint64) {
	sum := sha256.Sum256([]byte(token))
	h1 := binary.LittleEndian.Uint64(sum[0:8])
	h2 := binary.LittleEndian.Uint64(sum[8:16])
	if h2 == 0 {
		h2 = 0x9e3779b185ebca87 // golden ratio fallback prevents zero step
	}
	return h1, h2
}

func optimalBitCount(n int, p float64) uint64 {
	m := -float64(n) * math.Log(p) / (math.Ln2 * math.Ln2)
	if m < 64 {
		m = 64
	}
	return uint64(math.Ceil(m))
}

func optimalHashFunctions(m uint64, n int) uint64 {
	k := math.Ln2 * float64(m) / float64(n)
	if k < 1 {
		k = 1
	}
	return uint64(math.Round(k))
}
