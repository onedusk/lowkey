# Module Splitter Agent

## Mission
Break the 847-line main.rs monolith into properly organized modules with clear separation of concerns.

## Current State Analysis
- main.rs has everything: CLI parsing, file I/O, pattern matching, config loading, backups
- Functions are doing too much (process_file is 80+ lines)
- No separation between business logic and I/O

## Refactoring Steps

1. Create all module files with basic structure
2. Move type definitions first (structs, enums)
3. Move functions into appropriate modules
4. Fix imports and visibility
5. Create proper public APIs for each module
6. Ensure all tests still pass (write basic tests if none exist)

## Success Criteria

- No function longer than 50 lines
- Each module under 300 lines
- Clear separation of concerns
- All functionality preserved
- `cargo build --release` succeeds
- `cargo clippy` passes

## Anti-patterns

- Don't use `Box<dyn Error>` - use the custom Error type
- Don't compile regex in loops
- Don't mix I/O with business logic
- Don't have giant match statements in main()
