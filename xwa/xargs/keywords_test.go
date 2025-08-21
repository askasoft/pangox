package xargs

import (
	"testing"
)

func TestKeywordsContains(t *testing.T) {
	tests := []struct {
		name     string
		keywords Keywords
		input    string
		expected bool
	}{
		{"Empty list", Keywords{}, "go", false},
		{"Exact match", Keywords{"go", "lang"}, "go", true},
		{"Case-insensitive", Keywords{"GoLang"}, "golang", true},
		{"Partial match", Keywords{"gram"}, "programming", true},
		{"No match", Keywords{"code", "run"}, "build", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := tt.keywords.Contains(tt.input)
			if a != tt.expected {
				t.Errorf("Keywords.Contains(%q) = %v, want %v", tt.input, a, tt.expected)
			}
		})
	}
}

func TestKeywordsContainsAny(t *testing.T) {
	tests := []struct {
		name     string
		keywords Keywords
		input    []string
		expected bool
	}{
		{"Empty keywords", Keywords{}, []string{"go"}, false},
		{"Empty input", Keywords{"go"}, []string{}, false},
		{"One matches", Keywords{"go", "lang"}, []string{"build", "go"}, true},
		{"Case-insensitive match", Keywords{"GoLang"}, []string{"run", "GOLANG"}, true},
		{"Partial match", Keywords{"gram"}, []string{"programming"}, true},
		{"No matches", Keywords{"run", "code"}, []string{"compile", "link"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := tt.keywords.ContainsAny(tt.input...)
			if a != tt.expected {
				t.Errorf("Keywords.ContainsAny(%v) = %v, want %v", tt.input, a, tt.expected)
			}
		})
	}
}

func TestParseKeywords(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Keywords
	}{
		{"Empty input", "", Keywords{}},
		{"Single word", "hello", Keywords{"hello"}},
		{"Multiple words", "go lang", Keywords{"go", "lang"}},
		{"Quoted and unquoted", `"go lang" test`, Keywords{"go lang", "test"}},
		{"Mixed spacing", `   go    "hello world"   test  `, Keywords{"go", "hello world", "test"}},
		{"Incomplete quote", `"hello world`, Keywords{`"hello`, `world`}}, // because quote doesn't close
		{"Only spaces", "     ", Keywords{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := ParseKeywords(tt.input).String()
			w := Keywords(tt.expected).String()
			if w != a {
				t.Errorf("ParseKeywords(%v) = %q, want %q", tt.input, a, w)
			}
		})
	}
}

func TestNextKeyword(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantKey    string
		wantRest   string
		wantQuoted bool
	}{
		{"Empty input", "", "", "", false},
		{"Only spaces", "    ", "", "", false},
		{"Single word", "hello", "hello", "", false},
		{"Multiple words", "hello world", "hello", " world", false},
		{"Quoted word", `"hello world" test`, "hello world", " test", true},
		{"Quoted no close", `"hello world`, `"hello`, " world", false},
		{"No space", "nowordboundary", "nowordboundary", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, rest, quoted := NextKeyword(tt.input)
			if tt.wantKey != key || tt.wantRest != rest || tt.wantQuoted != quoted {
				t.Fatalf("NextKeyword(%q) = (%q, %q, %v), want (%q, %q, %v)", tt.input,
					key, rest, quoted, tt.wantKey, tt.wantRest, tt.wantQuoted,
				)
			}
		})
	}
}
