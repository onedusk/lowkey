// Package filters provides probabilistic data structures and heuristics for
// efficiently ignoring file paths that match user-defined glob patterns. It is
// designed to reduce the overhead of path matching by quickly filtering out
// paths that are unlikely to match any ignore patterns.
//
// The core component is a Bloom filter, which is populated with tokens
// extracted from the ignore patterns. Paths are then checked against this
// filter before performing more expensive glob matching.
package filters

import (
	"path/filepath"
	"strings"
)

// ExtractPatternTokens normalizes a glob pattern and extracts a set of
// representative tokens from it. These tokens are used to populate the Bloom
// filter, allowing for fast, probabilistic checks against file paths.
func ExtractPatternTokens(pattern string) []string {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return nil
	}

	pattern = filepath.ToSlash(pattern)
	segments := strings.Split(pattern, "/")
	seen := make(map[string]struct{})
	tokens := make([]string, 0, len(segments)*2)

	for _, segment := range segments {
		for _, token := range segmentTokens(segment) {
			tokens = appendUnique(tokens, seen, token)
		}
	}

	return tokens
}

// ExtractPathTokens derives a set of representative tokens from a file system
// path. These tokens can then be checked against the Bloom filter to quickly
// determine if the path is unlikely to match any ignore patterns. The tokens
// are lowercased to ensure case-insensitive matching.
func ExtractPathTokens(path string) []string {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil
	}

	clean := filepath.Clean(path)
	clean = filepath.ToSlash(clean)
	segments := strings.Split(clean, "/")
	seen := make(map[string]struct{})
	tokens := make([]string, 0, len(segments)+3)

	for _, segment := range segments {
		for _, token := range segmentTokens(segment) {
			tokens = appendUnique(tokens, seen, token)
		}
	}

	if len(segments) > 0 {
		appendUnique(tokens, seen, segments[0])
	}

	return tokens
}

func segmentTokens(segment string) []string {
	segment = strings.TrimSpace(segment)
	if segment == "" || segment == "." {
		return nil
	}

	segment = strings.ToLower(segment)
	base := stripGlobs(segment)
	result := make([]string, 0, 4)

	if base != "" {
		result = append(result, base)

		ext := filepath.Ext(base)
		if ext := strings.ToLower(ext); ext != "" {
			result = append(result, ext)
			stem := strings.TrimSuffix(base, ext)
			if stem != "" && stem != base {
				result = append(result, stem)
			}
		}

		parts := strings.FieldsFunc(base, func(r rune) bool { return r == '-' || r == '_' || r == '.' })
		for _, part := range parts {
			if part != "" && part != base {
				result = append(result, part)
			}
		}
	}

	return result
}

func stripGlobs(input string) string {
	var builder strings.Builder
	for _, r := range input {
		switch r {
		case '*', '?', '[', ']', '{', '}', '!':
			continue
		default:
			builder.WriteRune(r)
		}
	}
	return strings.Trim(builder.String(), " ")
}

func appendUnique(tokens []string, seen map[string]struct{}, token string) []string {
	token = strings.TrimSpace(strings.ToLower(token))
	if token == "" {
		return tokens
	}
	if _, ok := seen[token]; ok {
		return tokens
	}
	seen[token] = struct{}{}
	return append(tokens, token)
}
