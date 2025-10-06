package filters

// bloom_filter.go builds the ignore Bloom filter discussed in algorithm_design.
// Implement Add/Contains helpers tuned for CLI patterns. Benchmark with `go test`.

// TODO: Implement the Bloom filter for efficient glob pattern matching.
// - Choose a suitable Bloom filter library or implement one from scratch.
// - Implement `Add(string)` to add tokens from ignore patterns to the filter.
// - Implement `Contains(string)` to check if a file path token is likely in the set.
// - Write benchmarks to evaluate the filter's performance and tune its parameters
//   (e.g., size, number of hash functions).
