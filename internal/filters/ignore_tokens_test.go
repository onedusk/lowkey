package filters

import "testing"

func TestExtractPatternTokens(t *testing.T) {
	cases := []struct {
		pattern string
		expects []string
	}{
		{"**/*.go", []string{".go", "go"}},
		{"vendor/**", []string{"vendor"}},
		{"build/*.min.js", []string{"build", ".js", "js", "min"}},
	}

	for _, tc := range cases {
		tokens := ExtractPatternTokens(tc.pattern)
		for _, expected := range tc.expects {
			if !contains(tokens, expected) {
				t.Fatalf("pattern %q expected token %q in %v", tc.pattern, expected, tokens)
			}
		}
	}
}

func TestExtractPathTokens(t *testing.T) {
	path := "/Users/dev/project/src/main.go"
	tokens := ExtractPathTokens(path)

	for _, expected := range []string{"users", "dev", "project", "src", "main.go", ".go", "main"} {
		if !contains(tokens, expected) {
			t.Fatalf("path expected token %q in %v", expected, tokens)
		}
	}
}

func contains(tokens []string, target string) bool {
	for _, token := range tokens {
		if token == target {
			return true
		}
	}
	return false
}
