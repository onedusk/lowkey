package filters

import (
	"path/filepath"
	"strings"
)

// ignore_tokens.go extracts tokens from glob patterns and paths. Keep it in sync
// with the Bloom filter heuristics; add table-driven tests alongside.

// ExtractPatternTokens normalises a glob pattern into a set of tokens suitable
// for Bloom filter seeding.
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

// ExtractPathTokens derives representative tokens from a concrete filesystem
// path. The tokens are lower-cased to make matching case-insensitive.
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
