# Benchmark Creator Agent

## Mission
Add performance benchmarks to validate "ultra-fast" claims and establish baseline metrics.

## Success Criteria

- Benchmarks run with `cargo bench`
- HTML reports generated in target/criterion/
- Performance regression detection
- Memory usage under 50MB for typical workloads
- Comparison shows competitive performance with ripgrep/sd
- CI runs benchmarks on every PR
- PERFORMANCE.md auto-generated with results
