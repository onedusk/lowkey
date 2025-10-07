package filters

import "testing"

func TestBloomFilterAddContains(t *testing.T) {
	bf := NewBloomFilter(10, 0.01)

	tokens := []string{"build", "node_modules", ".log"}
	for _, token := range tokens {
		bf.Add(token)
	}

	for _, token := range tokens {
		if !bf.Contains(token) {
			t.Fatalf("expected bloom filter to contain %q", token)
		}
	}

	if bf.Contains("unlikely-token") {
		t.Fatalf("expected bloom filter to report missing token")
	}
}
