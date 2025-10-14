# Lowkey Examples

This directory contains example `.lowkey` ignore pattern files for various project types.

## Available Examples

### Basic Patterns (`.lowkey-basic`)
General-purpose patterns suitable for most projects:
- Version control directories (`.git/`, `.svn/`)
- Build outputs (`dist/`, `build/`)
- Temporary files (`*.tmp`, `*.swp`)
- OS-specific files (`.DS_Store`, `Thumbs.db`)

**Use case**: Starting point for any project type

### Web Development (`.lowkey-webdev`)
Optimized for JavaScript/TypeScript projects:
- Node.js dependencies (`node_modules/`)
- Framework build outputs (`.next/`, `.nuxt/`, `.svelte-kit/`)
- Package manager lock files
- Environment variables (`.env*`)

**Use case**: React, Vue, Svelte, Next.js, Remix, Nuxt projects

### Go Projects (`.lowkey-golang`)
Tailored for Go development:
- Compiled binaries (`*.exe`, `*.so`, `*.test`)
- Vendor dependencies (`vendor/`)
- Test coverage files (`*.coverprofile`)
- Go build cache (`.gocache/`)

**Use case**: Go applications and libraries

### Python Projects (`.lowkey-python`)
Designed for Python development:
- Byte-compiled files (`__pycache__/`, `*.pyc`)
- Virtual environments (`venv/`, `.venv/`)
- Distribution packages (`dist/`, `*.egg-info/`)
- Testing artifacts (`.pytest_cache/`, `.coverage`)

**Use case**: Python applications, Django, Flask, FastAPI

### Minimal Patterns (`.lowkey-minimal`)
Only essential patterns for maximum performance:
- `.git/` (version control)
- `node_modules/` (dependencies)
- `dist/` (build output)
- `.DS_Store` (OS cruft)

**Use case**: Performance-critical scenarios or when you want explicit control

## Usage

### Copy to Your Project

```bash
# Choose the appropriate example
cp examples/.lowkey-basic .lowkey

# Or for web development
cp examples/.lowkey-webdev .lowkey

# Make it yours by editing as needed
vim .lowkey
```

### Combine Multiple Examples

You can combine patterns from multiple examples:

```bash
# Start with basic patterns
cat examples/.lowkey-basic > .lowkey

# Add language-specific patterns
tail -n +2 examples/.lowkey-python >> .lowkey
```

### Verify Your Patterns

Test which files are ignored:

```bash
# Start watching with your patterns
lowkey watch .

# Create test files to see if they're ignored
touch test.tmp         # Should be ignored by basic
touch test.log         # Should be ignored by basic
mkdir __pycache__      # Should be ignored by python
```

## Pattern Syntax

Lowkey supports glob patterns:

- `*` - Matches any sequence of non-separator characters
- `**` - Matches zero or more directories
- `?` - Matches any single non-separator character
- `[abc]` or `[a-z]` - Character classes

### Examples

```
# Ignore all .log files anywhere
**/*.log

# Ignore node_modules at any depth
**/node_modules/

# Ignore specific file in root only
/README.tmp

# Ignore all test files
**/*_test.go
**/*.test.js
```

## Performance Tips

1. **Specific is Better**: Use `node_modules/` instead of `**/node_modules/` when possible
2. **Directory Matching**: Always end directory patterns with `/`
3. **Fewer Patterns**: More patterns = more Bloom filter checks
4. **Prioritize Large Directories**: Ignore `node_modules/`, `.git/`, `vendor/` first

## Customization

Each project is unique. Start with an example and customize:

```bash
# Copy example
cp examples/.lowkey-webdev .lowkey

# Add project-specific patterns
echo "# Project-specific" >> .lowkey
echo "/local-cache/" >> .lowkey
echo "*.generated.ts" >> .lowkey
```

## Testing Patterns

Use `lowkey watch` in foreground mode to verify patterns:

```bash
# Start watching
lowkey watch .

# In another terminal, create files
touch should-be-ignored.log
touch should-be-watched.txt

# Check if events appear in the watch output
```

## Related Documentation

- [Configuration & State](../README.md#configuration--state) - Full pattern syntax reference
- [Performance](../README.md#performance) - Impact of ignore patterns on performance
- [Manifest Schema](../docs/guides/manifest-schema.md) - Daemon configuration
